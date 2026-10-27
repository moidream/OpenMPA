package nativeui

import (
	"context"
	"fmt"
	"regexp"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"github.com/superops-team/OpenMPA/desktop/internal/dto"
	"github.com/superops-team/OpenMPA/desktop/internal/localization"
	"github.com/superops-team/OpenMPA/desktop/internal/usage"
)

const resourcePageSize = 20

var resourceSkillName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

type resourceEditor struct {
	draft, baseline string
	revision        int64
	loaded          bool
	needsRead       bool
}

func (e *resourceEditor) dirty() bool { return e.loaded && e.draft != e.baseline }

func (e *resourceEditor) accept(content string, revision int64) {
	if !e.dirty() {
		e.draft = content
	}
	e.baseline, e.revision, e.loaded, e.needsRead = content, revision, true, false
}

type resourceState struct {
	tab            string
	epoch, version uint64
	mounts         *usage.SkillMounts
	skills         *usage.SkillPage
	spaceID        string
	mcp            *usage.MCPDirectory
	mcpSession     string
	agents         usage.AgentsMDDocument
	profile        usage.MainProfileState
	agentsEditor   resourceEditor
	profileEditor  resourceEditor
	errors         map[string]*dto.SafeError
}

type resourceLoad struct {
	tab, region, sessionID string
	mounts                 *usage.SkillMounts
	skills                 *usage.SkillPage
	mcp                    *usage.MCPDirectory
	agents                 *usage.AgentsMDDocument
	profile                usage.MainProfileState
}

func (p *Controller) resourcesReady() bool {
	return !p.closed && p.services.Usage != nil && p.services.Connection != nil &&
		p.services.Connection.Client != nil && p.services.Connection.Client.State().Connected &&
		p.resources.epoch == p.epoch()
}

func (p *Controller) resourceSession() string {
	for _, item := range p.currentSessions() {
		if item.ID != "" && item.ID == p.shell.SelectedSession {
			return item.ID
		}
	}
	return ""
}

// Entry never reloads an existing editor: page navigation must not discard edits
// or silently renew a consumed write precondition.
func (p *Controller) loadResources() {
	if p.busy || p.closed {
		return
	}
	s := &p.resources
	if s.epoch != p.epoch() {
		*s = resourceState{epoch: p.epoch()}
	}
	if s.tab == "" {
		s.tab = "Skills"
	}
	switch s.tab {
	case "Skills":
		if s.skills != nil && s.skills.Region == p.settings.region {
			return
		}
	case "MCP":
		if s.mcp != nil && s.mcp.SessionID == p.resourceSession() {
			return
		}
	case "AGENTS.md":
		if s.agentsEditor.loaded || s.agentsEditor.needsRead {
			return
		}
	case "Main Profile":
		if s.profileEditor.loaded || s.profileEditor.needsRead {
			return
		}
	}
	p.loadResourcePage(1)
}

func (p *Controller) selectResourceTab(tab string) {
	if p.busy || p.closed {
		return
	}
	switch tab {
	case "Skills", "MCP", "AGENTS.md", "Main Profile":
		p.resources.tab = tab
		p.confirm, p.confirmed, p.err, p.message = false, nil, nil, ""
		p.loadResources()
	}
}

func (p *Controller) setResourceError(tab string, err *dto.SafeError) {
	if p.resources.errors == nil {
		p.resources.errors = make(map[string]*dto.SafeError)
	}
	p.resources.errors[tab] = err
}

func (p *Controller) loadResourcePage(page int) {
	if p.busy || p.closed {
		return
	}
	s := &p.resources
	if !p.resourcesReady() {
		p.setResourceError(s.tab, dto.Failure("capabilityUnavailable", "desktop", "readOnly", 0))
		return
	}
	if page < 1 || page > 10000 {
		p.setResourceError(s.tab, dto.Failure("invalidInput", "desktop", "never", 0))
		return
	}
	s.version++
	version, epoch := s.version, s.epoch
	tab, region, spaceID, sessionID := s.tab, p.settings.region, s.spaceID, p.resourceSession()
	switch tab {
	case "Skills":
		s.skills, s.mounts = nil, nil
	case "MCP":
		s.mcp = nil
		s.mcpSession = sessionID
	case "AGENTS.md":
		s.agentsEditor.needsRead = true
	case "Main Profile":
		s.profileEditor.needsRead = true
	}
	p.setResourceError(tab, nil)
	service := p.services.Usage
	p.async(func(ctx context.Context) (func(), *dto.SafeError) {
		result, err := readResourcePage(ctx, service, epoch, tab, region, spaceID, sessionID, page)
		return func() { p.acceptResourceLoad(version, epoch, result, err) }, nil
	})
}

func readResourcePage(ctx context.Context, service *usage.Service, epoch uint64, tab, region, spaceID, sessionID string, page int) (resourceLoad, *dto.SafeError) {
	result := resourceLoad{tab: tab, region: region, sessionID: sessionID}
	switch tab {
	case "Skills":
		mounts, err := service.DiscoverSkillSpaces(ctx, epoch)
		if err != nil {
			return result, err
		}
		result.mounts = &mounts
		mounted := false
		for _, space := range mounts.Spaces {
			mounted = mounted || space.ID == spaceID
		}
		if !mounted {
			spaceID, page = "", 1
			if len(mounts.Spaces) > 0 {
				spaceID = mounts.Spaces[0].ID
			}
		}
		if spaceID == "" {
			return result, nil
		}
		skills, err := service.ListSkills(ctx, epoch, region, spaceID, page, resourcePageSize)
		if err == nil {
			result.skills = &skills
		}
		return result, err
	case "MCP":
		if sessionID == "" {
			return result, dto.Failure("conversationRequired", "desktop", "never", 0)
		}
		value, err := service.ListMCPServers(ctx, epoch, sessionID)
		if err == nil {
			result.mcp = &value
		}
		return result, err
	case "AGENTS.md":
		value, err := service.ReadAgentsMD(ctx, epoch)
		if err == nil {
			result.agents = &value
		}
		return result, err
	case "Main Profile":
		id, err := service.ResourceInstanceID(ctx, epoch)
		if err != nil {
			result.profile.Status = usage.MainProfileUnavailable
			if err.HTTPStatus == 404 {
				result.profile.Status = usage.MainProfileUnsupportedOrUnconfigured
			}
			return result, err
		}
		result.profile, err = service.ReadMainProfile(ctx, epoch, id)
		return result, err
	}
	return result, dto.Failure("invalidInput", "desktop", "never", 0)
}

func (p *Controller) acceptResourceLoad(version, epoch uint64, result resourceLoad, err *dto.SafeError) {
	s := &p.resources
	if !p.resourcesReady() || s.version != version || s.epoch != epoch || s.tab != result.tab {
		return
	}
	if (result.tab == "MCP" && result.sessionID != p.resourceSession()) ||
		(result.tab == "Skills" && result.region != p.settings.region) {
		return
	}
	if result.skills != nil && (result.skills.Region != p.settings.region ||
		result.skills.ConnectionEpoch != epoch || result.skills.PlatformEpoch != p.services.Platform.Status().Value.Epoch) {
		return
	}
	if result.mcp != nil && result.mcp.SessionID != p.resourceSession() {
		return
	}
	p.setResourceError(result.tab, err)
	switch result.tab {
	case "Skills":
		s.mounts, s.skills = result.mounts, result.skills
		if result.skills != nil {
			s.spaceID = result.skills.SpaceID
		}
	case "MCP":
		s.mcp = result.mcp
	case "AGENTS.md":
		if err == nil && result.agents != nil {
			s.agents = *result.agents
			s.agentsEditor.accept(result.agents.Content, result.agents.Revision)
		}
	case "Main Profile":
		s.profile = result.profile
		if err == nil && result.profile.Status == usage.MainProfileConfigured && result.profile.Record != nil {
			s.profileEditor.accept(result.profile.Record.Profile.AgentsMD, result.profile.Record.Revision)
		}
	}
}

func (p *Controller) resourceEditor(tab string) *resourceEditor {
	if tab == "Main Profile" {
		return &p.resources.profileEditor
	}
	return &p.resources.agentsEditor
}

func (p *Controller) reloadResourceInstructions(tab string) {
	if p.busy || p.closed || p.resources.tab != tab {
		return
	}
	if !p.resourceEditor(tab).dirty() {
		p.loadResourcePage(1)
		return
	}
	epoch, version := p.resources.epoch, p.resources.version
	p.ask("Reload instructions", "Read the latest instructions? Your draft will be kept for review.", func() {
		if p.resourcesReady() && p.resources.tab == tab && p.resources.epoch == epoch && p.resources.version == version {
			p.loadResourcePage(1)
		}
	})
}

func (p *Controller) discardResourceDraft(tab string) {
	epoch, version := p.resources.epoch, p.resources.version
	p.ask("Discard instruction draft", "Replace your draft with the last successfully read instructions?", func() {
		if p.resourcesReady() && p.resources.tab == tab && p.resources.epoch == epoch && p.resources.version == version {
			editor := p.resourceEditor(tab)
			editor.draft = editor.baseline
		}
	})
}

func (p *Controller) resourceCanSave(tab string) bool {
	editor := p.resourceEditor(tab)
	if !p.resourcesReady() || !editor.loaded || editor.needsRead || !editor.dirty() {
		return false
	}
	if tab == "Main Profile" {
		return p.resources.profile.Status == usage.MainProfileConfigured && p.resources.profile.Record != nil
	}
	return tab == "AGENTS.md"
}

// prepareResourceSave consumes UI write authority before starting background
// work, including when navigation cancels delivery of its eventual result.
func (p *Controller) prepareResourceSave(tab string) (func(context.Context) (func(), *dto.SafeError), *dto.SafeError) {
	if p.busy || p.resources.tab != tab || !p.resourceCanSave(tab) {
		return nil, dto.Failure("conflict", "desktop", "reload", 0)
	}
	editor := p.resourceEditor(tab)
	if !utf8.ValidString(editor.draft) || utf8.RuneCountInString(editor.draft) > 40000 {
		return nil, dto.Failure("invalidInput", "desktop", "never", 0)
	}
	s := &p.resources
	s.version++
	version, epoch, draft := s.version, s.epoch, editor.draft
	document, profile, service := s.agents, s.profile, p.services.Usage
	document.Content = draft
	editor.needsRead = true
	p.setResourceError(tab, nil)
	return func(ctx context.Context) (func(), *dto.SafeError) {
		result := resourceLoad{tab: tab}
		var err *dto.SafeError
		if tab == "Main Profile" {
			result.profile, err = service.SaveMainProfileInstructions(ctx, epoch, profile, draft)
		} else {
			var saved usage.AgentsMDDocument
			saved, err = service.SaveAgentsMD(ctx, epoch, document)
			if err == nil {
				result.agents = &saved
			}
		}
		return func() {
			p.acceptResourceLoad(version, epoch, result, err)
			if err == nil && p.resourcesReady() && p.resources.version == version && p.resources.tab == tab {
				p.message = "Instructions saved."
			}
		}, nil
	}, nil
}

func (p *Controller) saveResourceInstructions(tab string) {
	work, err := p.prepareResourceSave(tab)
	if err != nil {
		p.setResourceError(tab, err)
		return
	}
	p.async(work)
}

func (p *Controller) chooseResourceSkill(name string) {
	s := &p.resources
	if p.busy || !p.resourcesReady() || s.skills == nil || s.skills.SpaceID != s.spaceID ||
		s.skills.Region != p.settings.region || s.skills.ConnectionEpoch != p.epoch() ||
		s.skills.PlatformEpoch != p.services.Platform.Status().Value.Epoch {
		p.setResourceError("Skills", dto.Failure("staleConnection", "desktop", "reload", 0))
		return
	}
	if resourceSkillName.MatchString(name) {
		for _, skill := range s.skills.Items {
			if skill.Name == name {
				p.useInConversation("$" + name)
				return
			}
		}
	}
	p.setResourceError("Skills", dto.Failure("invalidInput", "desktop", "never", 0))
}

func (p *Controller) renderResources(c *ui.Context) {
	s := &p.resources
	tabs := []string{"Skills", "MCP", "AGENTS.md", "Main Profile"}
	selected, labels := 0, make([]string, len(tabs))
	for i, tab := range tabs {
		labels[i] = localization.UI(tab)
		if s.tab == tab {
			selected = i
		}
	}
	if ui.Tabs(c.Key("resource-tabs"), &selected, labels...).Label(localization.UI("Resource sections")).Disabled(p.busy).Changed() {
		p.selectResourceTab(tabs[selected])
	}
	if !p.resourcesReady() {
		p.note(c, "Connect a Runtime to use resources.")
		return
	}
	if err := s.errors[s.tab]; err != nil {
		ui.Column(c.Key("resource-error")).Label(localization.UI("Resource error")).Children(func() {
			ui.Text(c, localization.Error(err.Code, err.Message)+statusSuffix(err)).
				TextColor(ui.RGB(175, 60, 50)).Selectable()
		})
	}
	switch s.tab {
	case "MCP":
		p.renderResourceMCP(c)
	case "AGENTS.md", "Main Profile":
		p.renderResourceInstructions(c, s.tab)
	default:
		p.renderResourceSkills(c)
	}
}

func (p *Controller) renderResourceSkills(c *ui.Context) {
	s := &p.resources
	p.button(c, "Refresh skills", func() { p.loadResourcePage(1) })
	p.note(c, "Only Skill Spaces mounted by this Runtime are shown.")
	ui.Text(c, localization.UI("Region")+": "+p.settings.region).Selectable()
	if s.mounts == nil {
		p.note(c, "Skill Spaces have not been loaded.")
		return
	}
	if len(s.mounts.Spaces) == 0 {
		p.note(c, "No Skill Spaces are mounted.")
		return
	}
	options, choice := []string{}, ""
	for _, space := range s.mounts.Spaces {
		label := space.Name + " (" + space.ID + ")"
		options = append(options, label)
		if space.ID == s.spaceID {
			choice = label
		}
	}
	ui.Select(c.Key("Mounted Skill Space"), &choice, options).Label(localization.UI("Mounted Skill Space")).Disabled(p.busy).OnChange(func() {
		for _, space := range s.mounts.Spaces {
			if choice == space.Name+" ("+space.ID+")" {
				s.spaceID = space.ID
				p.loadResourcePage(1)
				return
			}
		}
	})
	if s.skills == nil {
		p.note(c, "The skill directory is unavailable. Check platform access and refresh.")
		return
	}
	p.note(c, "A skill choice is added to the current conversation draft. It is not sent automatically.")
	if len(s.skills.Items) == 0 {
		p.note(c, "No skills are visible on this page.")
	}
	for i, skill := range s.skills.Items {
		ui.Column(c.Key(fmt.Sprintf("resource-skill-%d", i))).Gap(6).Padding(10).Children(func() {
			ui.Text(c, skill.Name).Bold().Selectable()
			ui.Text(c, skill.Description).Selectable()
			valid := resourceSkillName.MatchString(skill.Name)
			ui.Button(c.Key("Use skill "+skill.Name), localization.UI("Use in conversation")).
				Disabled(p.busy || !valid || p.resourceSession() == "").
				OnClick(func() { p.chooseResourceSkill(skill.Name) })
			if !valid {
				p.note(c, "This skill name cannot be used as a Runtime command.")
			}
		})
	}
	ui.Row(c.Key("skill-pagination")).Gap(8).Children(func() {
		page := s.skills
		ui.Text(c, fmt.Sprintf("%s %d / %d", localization.UI("Page"), page.PageNumber, max(1, (page.TotalCount+resourcePageSize-1)/resourcePageSize)))
		ui.Button(c.Key("Previous skills page"), localization.UI("Previous skills page")).
			Disabled(p.busy || page.PageNumber <= 1).OnClick(func() { p.loadResourcePage(page.PageNumber - 1) })
		ui.Button(c.Key("Next skills page"), localization.UI("Next skills page")).
			Disabled(p.busy || !page.HasMore || page.PageNumber >= 10000).OnClick(func() { p.loadResourcePage(page.PageNumber + 1) })
	})
}

func (p *Controller) renderResourceMCP(c *ui.Context) {
	sessionID := p.resourceSession()
	if sessionID == "" {
		p.note(c, "Select a conversation to view its MCP servers.")
		return
	}
	ui.Text(c, sessionID).Selectable()
	p.button(c, "Refresh MCP servers", func() { p.loadResourcePage(1) })
	p.button(c, "Configure conversation MCP", func() {
		p.resources.mcp = nil
		p.sessionAction("configureSession")
	})
	directory := p.resources.mcp
	if directory == nil || directory.SessionID != sessionID {
		p.note(c, "MCP servers have not been loaded for this conversation.")
		if p.resources.mcpSession != sessionID && !p.busy && p.window != nil {
			generation := p.generation
			p.window.Update(func() {
				if p.generation == generation && p.resources.tab == "MCP" && p.resourceSession() == sessionID {
					p.loadResourcePage(1)
				}
			})
		}
		return
	}
	if len(directory.Servers) == 0 {
		p.note(c, "No MCP servers are configured for this conversation.")
	}
	for _, server := range directory.Servers {
		ui.Column(c.Key("mcp-" + server.ID)).Gap(6).Padding(10).Children(func() {
			ui.Text(c, server.Name).Bold().Selectable()
			ui.Text(c, server.Description).Selectable()
			if !server.ToolsKnown {
				p.note(c, "Tool details are unavailable in this Runtime.")
			} else if len(server.Tools) == 0 {
				p.note(c, "No tools are listed for this server.")
			}
			for _, tool := range server.Tools {
				ui.Text(c, tool.Name).Bold().Selectable()
				ui.Text(c, tool.Description).Selectable()
			}
		})
	}
}

func (p *Controller) renderResourceInstructions(c *ui.Context, tab string) {
	editor := p.resourceEditor(tab)
	if tab == "AGENTS.md" {
		p.note(c, "Shared AGENTS.md can affect both Main and Worker.")
	} else {
		p.note(c, "Main Profile affects only Main. Saving instructions preserves its model and skills.")
	}
	if editor.loaded {
		ui.Text(c, fmt.Sprintf("%s: %d", localization.UI("Loaded revision"), editor.revision)).Selectable()
	}
	if tab == "Main Profile" {
		switch p.resources.profile.Status {
		case usage.MainProfileUnsupportedOrUnconfigured:
			p.note(c, "Main Profile is unsupported or not configured. No profile will be created.")
		case usage.MainProfileUnconfigured:
			p.note(c, "Main Profile is not configured. No profile will be created.")
		case usage.MainProfileUnsupported:
			p.note(c, "Main Profile is not supported by this Runtime.")
		case usage.MainProfileConfigured:
			if record := p.resources.profile.Record; record != nil {
				ui.Text(c, record.MPAInstanceID).Selectable()
				if record.Profile.Model != nil {
					ui.Text(c, localization.UI("Model")+": "+record.Profile.Model.ID).Selectable()
				}
				for _, skill := range record.Profile.Skills {
					ui.Text(c, skill.Name+" @ "+skill.Version).Selectable()
				}
			}
		default:
			p.note(c, "Main Profile identity or configuration is unavailable.")
		}
	}
	p.button(c, "Reload "+tab, func() { p.reloadResourceInstructions(tab) })
	if editor.loaded {
		ui.TextArea(c.Key(tab+" instructions"), &editor.draft).Label(localization.UI(tab+" instructions")).
			Lines(8, 20).FillWidth().Disabled(p.busy)
		if editor.dirty() {
			p.note(c, "Unsaved instruction draft.")
			p.note(c, "Latest saved instructions")
			ui.Text(c, editor.baseline).Selectable()
			p.button(c, "Discard instruction draft", func() { p.discardResourceDraft(tab) })
		}
	}
	if editor.needsRead {
		p.note(c, "Read the latest instructions before saving. Your draft is retained.")
	}
	ui.Button(c.Key("Save "+tab), localization.UI("Save "+tab)).Disabled(p.busy || !p.resourceCanSave(tab)).
		OnClick(func() { p.saveResourceInstructions(tab) })
}
