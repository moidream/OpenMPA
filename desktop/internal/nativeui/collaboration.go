package nativeui

import (
	"context"
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/superops-team/OpenMPA/desktop/internal/dto"
	"github.com/superops-team/OpenMPA/desktop/internal/localization"
	"github.com/superops-team/OpenMPA/desktop/internal/usage"
)

type collaborationDraft struct {
	alias, task string
	aliasDirty  bool
}

type collaborationState struct {
	binding          *usage.CollaborationBinding
	spaces           *usage.CollaborationSpacePage
	agents           *usage.CollaborationAgentPage
	spaceID, agentID string
	epoch, version   uint64
	bindingFresh     bool
	needsReload      bool
	drafts           map[string]*collaborationDraft
}

type collaborationSnapshot struct {
	binding  *usage.CollaborationBinding
	spaces   *usage.CollaborationSpacePage
	agents   *usage.CollaborationAgentPage
	spaceID  string
	complete bool
}

func (s *collaborationState) member() *usage.CollaborationAgent {
	if s.agents != nil {
		for i := range s.agents.Items {
			if s.agents.Items[i].ID == s.agentID {
				return &s.agents.Items[i]
			}
		}
	}
	return nil
}

func (s *collaborationState) draft() *collaborationDraft {
	if s.drafts == nil {
		s.drafts = make(map[string]*collaborationDraft)
	}
	key := s.spaceID + "/" + s.agentID
	if s.drafts[key] == nil {
		s.drafts[key] = &collaborationDraft{}
	}
	draft := s.drafts[key]
	if member := s.member(); member != nil && !draft.aliasDirty {
		draft.alias = ""
		if member.Alias != nil {
			draft.alias = *member.Alias
		}
	}
	return draft
}

func (p *Controller) collaborationReady() bool {
	return p.collaborationBindingReady() && !p.collaboration.needsReload
}

func (p *Controller) collaborationBindingReady() bool {
	s := &p.collaboration
	return !p.closed && p.services.Usage != nil && p.services.Connection.Status().Value.Connected &&
		s.epoch == p.epoch() && s.binding != nil && s.bindingFresh
}

func (p *Controller) loadCollaboration() {
	spacesPage, agentsPage := 1, 1
	if p.collaboration.spaces != nil {
		spacesPage = p.collaboration.spaces.Page
	}
	if p.collaboration.agents != nil {
		agentsPage = p.collaboration.agents.Page
	}
	p.loadCollaborationPage(spacesPage, agentsPage)
}

func (p *Controller) loadCollaborationPage(spacesPage, agentsPage int) {
	if p.busy || p.closed {
		return
	}
	if p.services.Usage == nil {
		p.err = dto.Failure("capabilityUnavailable", "desktop", "never", 0)
		return
	}
	s := &p.collaboration
	s.version++
	s.needsReload, s.bindingFresh = true, false
	version, epoch, spaceID := s.version, p.epoch(), s.spaceID
	service := p.services.Usage
	p.async(func(ctx context.Context) (func(), *dto.SafeError) {
		snapshot, err := readCollaboration(ctx, service, epoch, spaceID, spacesPage, agentsPage)
		return func() { p.acceptCollaborationLoad(version, epoch, snapshot) }, err
	})
}

func readCollaboration(ctx context.Context, service *usage.Service, epoch uint64, spaceID string, spacesPage, agentsPage int) (*collaborationSnapshot, *dto.SafeError) {
	binding, err := service.CollaborationBinding(ctx, epoch)
	if err != nil {
		return nil, err
	}
	snapshot := &collaborationSnapshot{binding: binding, spaceID: spaceID}
	spaces, err := service.CollaborationSpaces(ctx, epoch, spacesPage)
	if err != nil {
		return snapshot, err
	}
	if spaceID == "" && binding.SpaceID != nil {
		spaceID = *binding.SpaceID
	}
	snapshot.spaces, snapshot.spaceID = spaces, spaceID
	if spaceID != "" {
		snapshot.agents, err = service.CollaborationAgents(ctx, epoch, spaceID, agentsPage)
		if err != nil {
			return snapshot, err
		}
		// Binding and directory are separate reads; never expose a mixed snapshot.
		current, err := service.CollaborationBinding(ctx, epoch)
		if err != nil {
			return nil, err
		}
		if current.Revision != binding.Revision || current.Region != binding.Region ||
			collaborationBoundSpace(current) != collaborationBoundSpace(binding) {
			snapshot.binding = current
			return snapshot, dto.Failure("conflict", "runtimeIngress", "reload", 409)
		}
	}
	snapshot.complete = true
	return snapshot, nil
}

func collaborationBoundSpace(binding *usage.CollaborationBinding) string {
	if binding != nil && binding.SpaceID != nil {
		return *binding.SpaceID
	}
	return ""
}

func (p *Controller) acceptCollaborationLoad(version, epoch uint64, snapshot *collaborationSnapshot) {
	s := &p.collaboration
	if snapshot == nil || s.version != version || p.epoch() != epoch {
		return
	}
	s.binding, s.epoch, s.bindingFresh = snapshot.binding, epoch, true
	if !snapshot.complete {
		s.needsReload = true
		return
	}
	s.binding, s.spaces, s.agents, s.spaceID = snapshot.binding, snapshot.spaces, snapshot.agents, snapshot.spaceID
	s.epoch, s.needsReload = epoch, false
	if s.member() == nil {
		s.agentID = ""
	} else {
		s.draft()
	}
}

func (p *Controller) selectCollaborationSpace(id string) {
	if p.busy || p.closed || p.collaboration.spaces == nil {
		return
	}
	for _, space := range p.collaboration.spaces.Items {
		if space.ID == id {
			s := &p.collaboration
			s.spaceID, s.agentID, s.agents = id, "", nil
			p.loadCollaborationPage(s.spaces.Page, 1)
			return
		}
	}
}

func (p *Controller) renderCollaboration(c *ui.Context) {
	s := &p.collaboration
	if !p.services.Connection.Status().Value.Connected {
		p.note(c, "Connect a Runtime to manage collaboration.")
		return
	}
	p.button(c, "Refresh collaboration", p.loadCollaboration)
	p.note(c, "Browsing and refreshing do not change the Runtime. Writes require confirmation.")
	if s.needsReload || !s.bindingFresh {
		p.note(c, "Refresh collaboration before submitting changes. Your drafts are retained.")
	}
	p.section(c, "Current binding", func() {
		if s.binding == nil {
			p.note(c, "The current binding has not been loaded.")
			return
		}
		p.note(c, s.binding.Region)
		if id := collaborationBoundSpace(s.binding); id != "" {
			ui.Text(c, id).Selectable()
			if s.binding.SpaceName != nil {
				ui.Text(c, *s.binding.SpaceName).Selectable()
			}
			p.collaborationAction(c, "Unbind space", p.collaborationBindingReady(), func() { p.reviewCollaborationBinding(true) })
		} else {
			p.note(c, "No A2A space is bound.")
		}
	})
	p.section(c, "A2A spaces", func() {
		if s.spaces == nil {
			p.note(c, "Refresh collaboration to load spaces.")
			return
		}
		options, choice := []string{}, ""
		for _, space := range s.spaces.Items {
			label := space.Name + " (" + space.ID + ")"
			options = append(options, label)
			if space.ID == s.spaceID {
				choice = label
			}
		}
		ui.Select(c.Key("Collaboration space"), &choice, options).Label(localization.UI("Collaboration space")).Disabled(p.busy).OnChange(func() {
			for _, space := range s.spaces.Items {
				if choice == space.Name+" ("+space.ID+")" {
					p.selectCollaborationSpace(space.ID)
					return
				}
			}
		})
		if len(s.spaces.Items) == 0 {
			p.note(c, "No A2A spaces are visible.")
		}
		p.collaborationPagination(c, s.spaces.Page, s.spaces.PageSize, s.spaces.Total, true)
		if s.spaceID != "" {
			ui.Text(c, s.spaceID).Selectable()
			p.collaborationAction(c, "Bind selected space", p.collaborationReady(), func() { p.reviewCollaborationBinding(false) })
		}
	})
	p.section(c, "Space members", func() {
		if s.agents == nil {
			p.note(c, "Select a space to read its members.")
			return
		}
		if len(s.agents.Items) == 0 {
			p.note(c, "No members are visible in this space.")
		}
		p.rows(c, "collaboration-members", len(s.agents.Items), 220, func(i int) {
			member := s.agents.Items[i]
			ui.Row(c.Key(member.ID)).Gap(8).Children(func() {
				ui.Text(c, member.Name+" ("+member.ID+")").Grow(1)
				ui.Text(c, member.Status)
				p.buttonWithKey(c, "Select member "+member.ID, localization.UI("Select member"), func() {
					s.agentID = member.ID
					s.draft()
				})
			})
		})
		p.collaborationPagination(c, s.agents.Page, s.agents.PageSize, s.agents.Total, false)
	})
	if member := s.member(); member != nil {
		p.renderCollaborationMember(c, *member)
	}
}

func (p *Controller) collaborationPagination(c *ui.Context, page, size, total int, spaces bool) {
	ui.Row(c).Gap(8).Children(func() {
		pages := total / size
		if total%size != 0 {
			pages++
		}
		p.note(c, localization.UI("Page")+" "+fmt.Sprint(page)+" / "+fmt.Sprint(max(1, pages)))
		p.note(c, localization.UI("Total")+" "+fmt.Sprint(total))
		change := func(next int) {
			spacePage, agentPage := 1, 1
			if p.collaboration.spaces != nil {
				spacePage = p.collaboration.spaces.Page
			}
			if p.collaboration.agents != nil {
				agentPage = p.collaboration.agents.Page
			}
			if spaces {
				spacePage = next
			} else {
				agentPage = next
			}
			p.loadCollaborationPage(spacePage, agentPage)
		}
		previous, next := "Previous members page", "Next members page"
		if spaces {
			previous, next = "Previous spaces page", "Next spaces page"
		}
		if page > 1 {
			p.button(c, previous, func() { change(page - 1) })
		}
		if page < 10000 && page*size < total {
			p.button(c, next, func() { change(page + 1) })
		}
	})
}

func (p *Controller) collaborationAction(c *ui.Context, label string, enabled bool, action func()) {
	if enabled {
		p.button(c, label, action)
	} else {
		ui.Button(c.Key(label), localization.UI(label)).Disabled(true)
	}
}

func (p *Controller) renderCollaborationMember(c *ui.Context, member usage.CollaborationAgent) {
	s := &p.collaboration
	draft := s.draft()
	p.section(c, "Member details", func() {
		ui.Text(c, member.Name).Bold().Selectable()
		ui.Text(c, member.Description).Selectable()
		ui.Text(c, member.ID+" / "+member.Status).Selectable()
		if member.Card == nil {
			p.note(c, "Agent Card unavailable in this Runtime.")
		} else {
			ui.Text(c, member.Card.Name).Selectable()
			ui.Text(c, member.Card.Description).Selectable()
			ui.Text(c, member.Card.Version+" / "+member.Card.ProtocolVersion).Selectable()
		}
		p.note(c, "Current alias")
		if member.Alias != nil {
			ui.Text(c, *member.Alias).Selectable()
		}
		if member.AliasAvailable {
			ui.Text(c, localization.UI("Member alias")).Label(localization.UI("Member alias") + " caption").FontSize(12)
			ui.TextInput(c.Key("Member alias"), &draft.alias).Label(localization.UI("Member alias")).FillWidth().Disabled(p.busy).
				OnChange(func() { draft.aliasDirty = true })
			p.note(c, "Leave the alias empty to remove it.")
			p.collaborationAction(c, "Save member alias", p.collaborationReady(), p.reviewCollaborationAlias)
		} else {
			p.note(c, "Aliases are unavailable in this Runtime.")
		}
		ui.Text(c, localization.UI("Collaboration task")).Label(localization.UI("Collaboration task") + " caption").FontSize(12)
		ui.TextArea(c.Key("Collaboration task"), &draft.task).Label(localization.UI("Collaboration task")).Lines(3, 8).FillWidth().Disabled(p.busy)
		p.note(c, "This adds an @mention to the conversation draft. Review and send it in the conversation.")
		if member.Alias == nil || *member.Alias == "" {
			p.note(c, "Save a unique member alias before using this collaborator.")
		}
		usable := p.collaborationReady() && collaborationBoundSpace(s.binding) == s.spaceID && member.Status == "running" &&
			member.AliasAvailable && member.Alias != nil && *member.Alias != "" &&
			(member.RuntimeID == "" || member.RuntimeID != p.services.Connection.Status().Value.RuntimeID)
		p.collaborationAction(c, "Use in conversation", usable, p.useCollaborationMember)
	})
}

func (p *Controller) reviewCollaborationBinding(unbind bool) {
	s := &p.collaboration
	if !p.collaborationBindingReady() || (!unbind && !p.collaborationReady()) || p.busy {
		p.err = dto.Failure("conflict", "desktop", "reload", 0)
		return
	}
	var target *string
	title := "Unbind space"
	if !unbind {
		visible := false
		if s.spaces != nil {
			for _, space := range s.spaces.Items {
				visible = visible || space.ID == s.spaceID
			}
		}
		if !visible {
			p.err = dto.Failure("invalidInput", "desktop", "never", 0)
			return
		}
		id := s.spaceID
		target, title = &id, "Bind selected space"
	} else if s.binding.SpaceID == nil {
		return
	}
	epoch, version, revision := s.epoch, s.version, s.binding.Revision
	targetID := collaborationBoundSpace(s.binding)
	if target != nil {
		targetID = *target
	}
	p.ask(title, localization.UI("Change the Runtime's A2A space binding?")+"\n"+targetID, func() {
		if !p.collaborationBindingReady() || p.busy || p.epoch() != epoch || s.version != version || s.binding.Revision != revision {
			p.err = dto.Failure("staleConnection", "desktop", "reload", 0)
			return
		}
		s.bindingFresh = false
		p.async(func(ctx context.Context) (func(), *dto.SafeError) {
			value, err := p.services.Usage.SetCollaborationBinding(ctx, epoch, target, revision)
			return func() {
				if value != nil && s.version == version {
					s.binding, s.bindingFresh = value, true
					p.message = "Space binding saved."
				}
			}, err
		})
	})
}

func (p *Controller) reviewCollaborationAlias() {
	s := &p.collaboration
	member := s.member()
	if !p.collaborationReady() || p.busy || member == nil || !member.AliasAvailable {
		p.err = dto.Failure("conflict", "desktop", "reload", 0)
		return
	}
	epoch, version, spaceID, agentID, revision := s.epoch, s.version, s.spaceID, member.ID, member.AliasRevision
	draft := s.draft()
	alias := draft.alias
	p.ask("Save member alias", localization.UI("Update this member's alias in the selected space?")+"\n"+spaceID+" / "+agentID+"\n"+alias, func() {
		if !p.collaborationReady() || p.busy || p.epoch() != epoch || s.version != version || s.spaceID != spaceID || s.agentID != agentID {
			p.err = dto.Failure("staleConnection", "desktop", "reload", 0)
			return
		}
		s.needsReload = true
		p.async(func(ctx context.Context) (func(), *dto.SafeError) {
			value, err := p.services.Usage.SetCollaborationAlias(ctx, epoch, spaceID, agentID, &alias, revision)
			return func() {
				if value != nil && s.version == version && s.spaceID == spaceID && s.agentID == agentID {
					member.Alias, member.AliasRevision = value.Alias, value.Revision
					draft.alias, draft.aliasDirty = "", false
					if value.Alias != nil {
						draft.alias = *value.Alias
					}
					s.needsReload = false
					p.message = "Member alias saved."
				}
			}, err
		})
	})
}

func (p *Controller) useCollaborationMember() {
	s := &p.collaboration
	member := s.member()
	if !p.collaborationReady() || p.busy || member == nil || collaborationBoundSpace(s.binding) != s.spaceID ||
		(member.RuntimeID != "" && member.RuntimeID == p.services.Connection.Status().Value.RuntimeID) {
		p.err = dto.Failure("conflict", "desktop", "reload", 0)
		return
	}
	text, err := usage.CollaborationMention(*member, s.draft().task)
	if err != nil {
		p.err = err
		return
	}
	p.useInConversation(text)
}
