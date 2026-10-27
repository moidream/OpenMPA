package nativeui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/superops-team/OpenMPA/desktop/internal/cloud"
	"github.com/superops-team/OpenMPA/desktop/internal/connection"
	"github.com/superops-team/OpenMPA/desktop/internal/dto"
	"github.com/superops-team/OpenMPA/desktop/internal/host"
	"github.com/superops-team/OpenMPA/desktop/internal/ipc"
	"github.com/superops-team/OpenMPA/desktop/internal/localization"
	"github.com/superops-team/OpenMPA/desktop/internal/usage"
)

func nativeResourceController(t *testing.T, handler http.HandlerFunc) *Controller {
	t.Helper()
	previousLocale := localization.CurrentLocale()
	localization.SetCurrentLocale(localization.English)
	t.Cleanup(func() { localization.SetCurrentLocale(previousLocale) })
	p := testController(t)
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	epoch, err := p.services.Connection.Client.Bind(connection.Binding{
		Endpoint: server.URL, RuntimeID: "r-platform", RuntimeKey: "fixture-key",
	}, true)
	if err != nil {
		t.Fatal(err)
	}
	p.services.Usage = usage.New(p.services.Connection.Client, nil)
	p.resources = resourceState{tab: "AGENTS.md", epoch: epoch}
	p.shell = host.ShellState{View: "resources", ConnectionEpoch: epoch, SelectedSession: "s-one",
		Sessions: []host.NavigationItem{{ID: "s-one", Title: "First"}, {ID: "s-two", Title: "Second"}}}
	return p
}

func nativeReadResources(p *Controller, tab string) *dto.SafeError {
	p.resources.tab = tab
	p.resources.version++
	p.resourceEditor(tab).needsRead = true
	result, err := readResourcePage(context.Background(), p.services.Usage, p.epoch(), tab, p.settings.region, p.resources.spaceID, p.resourceSession(), 1)
	p.acceptResourceLoad(p.resources.version, p.epoch(), result, err)
	return err
}

func nativeSaveResources(t *testing.T, p *Controller, tab string) {
	t.Helper()
	work, err := p.prepareResourceSave(tab)
	if err != nil || work == nil {
		t.Fatalf("prepare save: %v", err)
	}
	apply, err := work(context.Background())
	p.complete(p.generation, p.epoch(), p.services.Platform.Status().Value.Epoch, false, apply, err)
}

func TestNativeResourcesEditorsInputNavigationAndConfirmation(t *testing.T) {
	var calls atomic.Int32
	p := nativeResourceController(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		io.WriteString(w, `{"content":"saved","revision":3}`)
	})
	if err := nativeReadResources(p, "AGENTS.md"); err != nil {
		t.Fatal(err)
	}
	p.resources.profileEditor = resourceEditor{loaded: true, baseline: "main", draft: "main"}
	p.resources.profile.Status = usage.MainProfileUnsupportedOrUnconfigured
	tt := ui.NewTester(func(c *ui.Context) {
		p.Render(c, p.shell)
		p.RenderOverlay(c, p.shell)
	}, 1100, 1500)
	if !tt.HasText("Loaded revision: 3") {
		t.Fatal("shared instructions hide the loaded CAS revision")
	}
	for _, label := range []string{"Skills", "MCP", "AGENTS.md", "Main Profile", "AGENTS.md instructions", "Save AGENTS.md", "Reload AGENTS.md"} {
		if _, ok := tt.Find(label); !ok {
			t.Fatalf("missing native control %q", label)
		}
	}
	if err := tt.Click("AGENTS.md instructions"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Super, ui.KeyA)
	tt.Type("unsaved")
	if p.resources.agentsEditor.draft != "unsaved" || !p.resources.agentsEditor.dirty() {
		t.Fatal("native editor did not retain input")
	}
	for _, label := range []string{"Reload AGENTS.md", "Discard instruction draft"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
		if !p.confirm || calls.Load() != 1 {
			t.Fatal("dirty reload/discard bypassed confirmation or started a request")
		}
		if err := tt.Click("Cancel"); err != nil {
			t.Fatal(err)
		}
	}
	if err := tt.Click("Main Profile"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("Main Profile is unsupported or not configured. No profile will be created.") {
		t.Fatal("ambiguous Main Profile status hidden")
	}
	if err := tt.Click("AGENTS.md"); err != nil {
		t.Fatal(err)
	}
	p.observe(host.ShellState{View: "chat", ConnectionEpoch: p.epoch()})
	p.observe(host.ShellState{View: "resources", ConnectionEpoch: p.epoch()})
	p.loadResources()
	if p.resources.agentsEditor.draft != "unsaved" || calls.Load() != 1 {
		t.Fatal("tab/page re-entry reloaded or lost the draft")
	}
	p.discardResourceDraft("AGENTS.md")
	p.confirmed()
	if p.resources.agentsEditor.draft != "saved" || p.resources.agentsEditor.dirty() {
		t.Fatal("explicit discard failed")
	}
}

func TestNativeResourcesAgentsMDSaveAndReadRecovery(t *testing.T) {
	for _, tc := range []struct {
		name, body, code string
		status           int
	}{
		{"success", `{"content":"draft","revision":8}`, "", 200},
		{"conflict", `{"detail":"secret"}`, "conflict", 409},
		{"precondition", `{}`, "conflict", 412},
		{"unknown", `{}`, "outcomeUnknown", 200},
		{"server", `{}`, "outcomeUnknown", 500},
		{"auth", `{"detail":"secret"}`, "authRequired", 401},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var reads, writes atomic.Int32
			p := nativeResourceController(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					if reads.Add(1) == 1 {
						io.WriteString(w, `{"content":"old","revision":7}`)
					} else {
						io.WriteString(w, `{"content":"remote edit","revision":9}`)
					}
					return
				}
				writes.Add(1)
				var document usage.AgentsMDDocument
				if r.Method != http.MethodPost || json.NewDecoder(r.Body).Decode(&document) != nil ||
					document.Revision != 7 || document.Content != "draft" {
					t.Errorf("CAS input lost: %+v", document)
				}
				w.WriteHeader(tc.status)
				io.WriteString(w, tc.body)
			})
			if err := nativeReadResources(p, "AGENTS.md"); err != nil {
				t.Fatal(err)
			}
			p.resources.agentsEditor.draft = "draft"
			nativeSaveResources(t, p, "AGENTS.md")
			editor, err := &p.resources.agentsEditor, p.resources.errors["AGENTS.md"]
			if editor.draft != "draft" || writes.Load() != 1 {
				t.Fatal("save lost draft or replayed a write")
			}
			if tc.code == "" {
				if err != nil || editor.dirty() || editor.needsRead || p.resources.agents.Revision != 8 {
					t.Fatalf("save not accepted: %+v %v", editor, err)
				}
				return
			}
			if err == nil || err.Code != tc.code || !editor.needsRead || !editor.dirty() || p.resourceCanSave("AGENTS.md") {
				t.Fatalf("write failure not fenced: %+v %v", editor, err)
			}
			tt := ui.NewTester(p.renderResources, 1000, 1500)
			if _, ok := tt.Find("Resource error"); !ok || tt.HasText("secret") {
				t.Fatal("safe native error missing or provider details leaked")
			}
			if work, err := p.prepareResourceSave("AGENTS.md"); work != nil || err == nil || writes.Load() != 1 {
				t.Fatal("write could be retried before a read")
			}
			p.loadResources()
			if reads.Load() != 1 {
				t.Fatal("re-entry silently renewed write authority")
			}
			if err := nativeReadResources(p, "AGENTS.md"); err != nil {
				t.Fatal(err)
			}
			if editor.draft != "draft" || editor.baseline != "remote edit" || p.resources.agents.Revision != 9 ||
				!p.resourceCanSave("AGENTS.md") || reads.Load() != 2 {
				t.Fatal("read recovery lost draft or did not renew CAS")
			}
		})
	}
}

func TestNativeResourcesInputLimitsAndFailedReads(t *testing.T) {
	var calls atomic.Int32
	p := nativeResourceController(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(403)
		io.WriteString(w, `{"detail":"private"}`)
	})
	p.resources.agentsEditor = resourceEditor{loaded: true, baseline: "old", draft: "draft"}
	for _, input := range []string{strings.Repeat("x", 40001), string([]byte{0xff})} {
		p.resources.agentsEditor.draft = input
		if work, err := p.prepareResourceSave("AGENTS.md"); work != nil || err == nil || err.Code != "invalidInput" {
			t.Fatal("invalid instructions accepted")
		}
		if p.resources.agentsEditor.needsRead || calls.Load() != 0 {
			t.Fatal("local validation consumed write authority or used network")
		}
	}
	for _, input := range []string{"", strings.Repeat("\u4e2d", 40000)} {
		p.resources.agentsEditor.draft = input
		work, err := p.prepareResourceSave("AGENTS.md")
		if work == nil || err != nil {
			t.Fatal("valid boundary rejected", err)
		}
		p.resources.agentsEditor.needsRead = false
	}
	p.resources.agentsEditor.draft = "retained"
	if err := nativeReadResources(p, "AGENTS.md"); err == nil || err.Code != "denied" {
		t.Fatal("read failure hidden")
	}
	if p.resources.agentsEditor.draft != "retained" || p.resources.agentsEditor.baseline != "old" || p.resourceCanSave("AGENTS.md") {
		t.Fatal("failed read replaced draft or allowed save")
	}
}

const nativeProfileFixture = `{"mpaInstanceId":"mi-current","revision":7,"etag":"\"7\"","digest":"old",
	"profile":{"agentsMd":"old main","model":{"id":"chosen-model"},"skills":[{"name":"writer","version":"v1","content":"skill body"}]},
	"sourceProfileId":"source","sourceProfileDigest":"digest"}`

func TestNativeResourcesMainProfileIdentityAndPreservation(t *testing.T) {
	var identityReads, profileReads, writes atomic.Int32
	p := nativeResourceController(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/agents":
			identityReads.Add(1)
			io.WriteString(w, `{"mpaInstanceId":"mi-current","basePrivateUrl":"private"}`)
		case "/api/v1/agents/mi-current/main-profile":
			if r.Method == http.MethodGet {
				profileReads.Add(1)
				io.WriteString(w, nativeProfileFixture)
				return
			}
			writes.Add(1)
			var body struct {
				Profile             usage.MainProfile `json:"profile"`
				SourceProfileID     string            `json:"sourceProfileId"`
				SourceProfileDigest string            `json:"sourceProfileDigest"`
			}
			if json.NewDecoder(r.Body).Decode(&body) != nil || r.Method != http.MethodPut ||
				r.Header.Get("If-Match") != `"7"` || r.Header.Get("Idempotency-Key") == "" ||
				body.Profile.AgentsMD != "updated main" || body.Profile.Model.ID != "chosen-model" ||
				len(body.Profile.Skills) != 1 || body.Profile.Skills[0].Content != "skill body" ||
				body.SourceProfileID != "source" || body.SourceProfileDigest != "digest" {
				t.Errorf("existing-only profile update lost settings: %+v", body)
			}
			io.WriteString(w, strings.NewReplacer(`"revision":7`, `"revision":8`, `\"7\"`, `\"8\"`, "old main", "updated main").Replace(nativeProfileFixture))
		default:
			t.Error("wrong identity route:", r.URL.Path)
			w.WriteHeader(404)
		}
	})
	if err := nativeReadResources(p, "Main Profile"); err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(p.renderResources, 1000, 1500)
	for _, text := range []string{"mi-current", "Model: chosen-model", "writer @ v1", "Loaded revision: 7"} {
		if !tt.HasText(text) {
			t.Errorf("missing preserved profile setting %q", text)
		}
	}
	p.resources.profileEditor.draft = "updated main"
	nativeSaveResources(t, p, "Main Profile")
	if p.resources.errors["Main Profile"] != nil || p.resources.profile.Record.Revision != 8 ||
		p.resources.profileEditor.dirty() || p.resources.profileEditor.needsRead ||
		identityReads.Load() != 1 || profileReads.Load() != 1 || writes.Load() != 1 {
		t.Fatalf("profile update failed: %+v", p.resources)
	}
}

func TestNativeResourcesMainProfileUnknownNeverBlindResaves(t *testing.T) {
	var writes atomic.Int32
	p := nativeResourceController(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/agents" {
			io.WriteString(w, `{"mpaInstanceId":"mi-current"}`)
		} else if r.Method == http.MethodGet {
			io.WriteString(w, nativeProfileFixture)
		} else {
			writes.Add(1)
			io.WriteString(w, `{}`)
		}
	})
	if err := nativeReadResources(p, "Main Profile"); err != nil {
		t.Fatal(err)
	}
	p.resources.profileEditor.draft = "kept"
	nativeSaveResources(t, p, "Main Profile")
	if p.resources.errors["Main Profile"].Code != "outcomeUnknown" || p.resourceCanSave("Main Profile") ||
		p.resources.profileEditor.draft != "kept" || !p.resources.profileEditor.needsRead || p.resources.profileEditor.revision != 7 {
		t.Fatal("unknown Main Profile save was not fenced")
	}
	tt := ui.NewTester(p.renderResources, 1000, 1500)
	if !tt.HasText("Loaded revision: 7") {
		t.Fatal("failed save lost the loaded revision")
	}
	p.loadResources()
	if err := nativeReadResources(p, "Main Profile"); err != nil || !p.resourceCanSave("Main Profile") ||
		p.resources.profileEditor.draft != "kept" || writes.Load() != 1 {
		t.Fatal("Main Profile read recovery failed", err)
	}
}

func TestNativeResourcesIdentityAnd404FailClosed(t *testing.T) {
	for _, tc := range []struct {
		name, identity string
		status         int
		wantProfiles   int32
	}{
		{"platform-id", `{"mpaInstanceId":"r-platform"}`, 200, 0},
		{"missing-id", `{"name":"main"}`, 200, 0},
		{"escape", `{"mpaInstanceId":"../other"}`, 200, 0},
		{"wrong-type", `{"mpaInstanceId":9}`, 200, 0},
		{"identity-404", `{}`, 404, 0},
		{"profile-404", `{"mpaInstanceId":"mi-current"}`, 200, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var profiles, writes atomic.Int32
			p := nativeResourceController(t, func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					writes.Add(1)
				}
				if r.URL.Path == "/api/v1/agents" {
					w.WriteHeader(tc.status)
					io.WriteString(w, tc.identity)
				} else {
					profiles.Add(1)
					w.WriteHeader(404)
				}
			})
			if err := nativeReadResources(p, "Main Profile"); err == nil {
				t.Fatal("missing identity/profile was accepted")
			}
			tt := ui.NewTester(p.renderResources, 1000, 1500)
			if err := tt.Click("Save Main Profile"); err != nil {
				t.Fatal(err)
			}
			if p.resourceCanSave("Main Profile") || p.resources.profileEditor.loaded || profiles.Load() != tc.wantProfiles || writes.Load() != 0 {
				t.Fatal("missing/ambiguous profile was writable")
			}
			if strings.Contains(tc.name, "404") && !tt.HasText("Main Profile is unsupported or not configured. No profile will be created.") {
				t.Fatal("404 ambiguity hidden")
			}
		})
	}
}

func TestNativeResourcesMCPUsesSelectedSessionAndShowsTools(t *testing.T) {
	p := nativeResourceController(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/sessions/s-one/mcp-servers" {
			t.Error("wrong session request", r.Method, r.URL.Path)
		}
		io.WriteString(w, `{"sessionId":"s-one","mcpServers":[
			{"id":"search","name":"Search","description":"Find pages","tools":[{"name":"query","description":"Query pages"}]},
			{"id":"unknown","name":"Unknown tools"},{"id":"empty","name":"Empty tools","tools":[]}]}`)
	})
	if err := nativeReadResources(p, "MCP"); err != nil {
		t.Fatal(err)
	}
	tt := ui.NewTester(p.renderResources, 1100, 1500)
	for _, text := range []string{"Search", "Find pages", "query", "Query pages", "Tool details are unavailable in this Runtime.", "No tools are listed for this server."} {
		if !tt.HasText(text) {
			t.Errorf("MCP detail missing: %s", text)
		}
	}
	if _, ok := tt.Find("Configure conversation MCP"); !ok {
		t.Fatal("existing configuration entry missing")
	}
	p.shell.SelectedSession = "s-two"
	tt.Frame()
	if tt.HasText("Query pages") || !tt.HasText("MCP servers have not been loaded for this conversation.") {
		t.Fatal("MCP results leaked across sessions")
	}
	old := resourceLoad{tab: "MCP", sessionID: "s-one", mcp: p.resources.mcp}
	p.acceptResourceLoad(p.resources.version, p.epoch(), old, nil)
	p.resources.mcp = nil
	old.mcp = nil
	p.acceptResourceLoad(p.resources.version, p.epoch(), old, dto.Failure("denied", "runtimeIngress", "never", 403))
	if p.resources.errors["MCP"] != nil {
		t.Fatal("stale session error applied")
	}
	p.shell.ConnectionEpoch++
	tt.Frame()
	if !tt.HasText("Select a conversation to view its MCP servers.") {
		t.Fatal("stale session membership accepted")
	}
}

func TestNativeResourcesSkillNamesAndSelectionGuards(t *testing.T) {
	for _, name := range []string{"writer", "Code.v1_test-2", strings.Repeat("a", 128)} {
		if !resourceSkillName.MatchString(name) {
			t.Errorf("safe Runtime name rejected: %q", name)
		}
	}
	for _, name := range []string{"", "-writer", "writer task", "writer\n$other", "$writer", "a/b", "a;run", "\u4e2d", strings.Repeat("a", 129)} {
		if resourceSkillName.MatchString(name) {
			t.Errorf("unsafe Runtime name accepted: %q", name)
		}
	}
	for _, boundary := range []string{"valid", "unlisted", "bad-name", "epoch", "platform", "region", "space", "no-session"} {
		t.Run(boundary, func(t *testing.T) {
			p := nativeResourceController(t, func(w http.ResponseWriter, r *http.Request) { t.Error("selection used network") })
			p.resources.tab, p.resources.spaceID = "Skills", "mounted"
			p.resources.skills = &usage.SkillPage{SpaceID: "mounted", Region: p.settings.region, ConnectionEpoch: p.epoch(),
				PlatformEpoch: p.services.Platform.Status().Value.Epoch, Items: []usage.Skill{{Name: "writer"}, {Name: "bad name"}}}
			name := "writer"
			switch boundary {
			case "unlisted":
				name = "other"
			case "bad-name":
				name = "bad name"
			case "epoch":
				p.resources.skills.ConnectionEpoch++
			case "platform":
				p.resources.skills.PlatformEpoch++
			case "region":
				p.settings.region = "cn-shanghai"
			case "space":
				p.resources.spaceID = "other"
			case "no-session":
				p.shell.SelectedSession = ""
			}
			p.chooseResourceSkill(name)
			failed := p.err != nil || p.resources.errors["Skills"] != nil
			if failed != (boundary != "valid") || p.busy {
				t.Fatalf("selection boundary %s: error=%v busy=%v", boundary, failed, p.busy)
			}
		})
	}
}

type nativeResourceTransport func(*http.Request) (*http.Response, error)

func (f nativeResourceTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestNativeResourcesSkillDiscoveryRegionAndPagination(t *testing.T) {
	const card = `{"name":"main","capabilities":{"extensions":[{"uri":"urn:veadk:mpa:resource-topology:v1","params":{
		"nodes":[{"id":"agent:main","name":"main","kind":"agent"},{"id":"space:mounted","kind":"skill-space","name":"Mounted","resourceId":"mounted"},
		{"id":"space:other","kind":"skill-space","name":"Other","resourceId":"unmounted"}],
		"edges":[{"source":"agent:main","target":"space:mounted","relation":"mounts"}]}}]}}`
	p := nativeResourceController(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/agent-card.json" || r.Method != http.MethodGet {
			t.Error("wrong discovery request", r.URL.Path)
		}
		io.WriteString(w, card)
	})
	var actions atomic.Int32
	platformServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("Action") == "GetCallerIdentity" {
			io.WriteString(w, `{"Result":{"AccountId":"account","IdentityType":"User","UserId":"user"}}`)
			return
		}
		actions.Add(1)
		var body struct {
			SkillSpaceID string `json:"SkillSpaceId"`
		}
		if r.URL.Query().Get("Action") != "ListSkillsBySpaceId" || json.NewDecoder(r.Body).Decode(&body) != nil ||
			body.SkillSpaceID != "mounted" || !strings.Contains(r.Header.Get("Authorization"), "/cn-shanghai/agentkit/request") {
			t.Error("wrong mounted skill query or Region")
		}
		items := make([]map[string]string, 21)
		for i := range items {
			items[i] = map[string]string{"Name": fmt.Sprintf("skill-%d", i), "Description": fmt.Sprintf("Description %d", i)}
		}
		json.NewEncoder(w).Encode(map[string]any{"Result": map[string]any{"Items": items}})
	}))
	defer platformServer.Close()
	platform := cloud.NewWithTransport(nativeResourceTransport(func(r *http.Request) (*http.Response, error) {
		// Keep all I/O on httptest while retaining signed Region/Action inputs.
		target, _ := http.NewRequestWithContext(r.Context(), r.Method, platformServer.URL+"?"+r.URL.RawQuery, r.Body)
		target.Header = r.Header.Clone()
		return platformServer.Client().Transport.RoundTrip(target)
	}))
	if _, err := platform.SignIn(context.Background(), dto.SignInRequest{Region: "cn-shanghai", AccessKey: "fixture-access", SecretKey: "fixture-secret"}); err != nil {
		t.Fatal(err)
	}
	p.services.Platform = ipc.NewPlatform(platform, p.services.Connection.Client)
	p.services.Usage = usage.New(p.services.Connection.Client, platform)
	p.settings.region, p.resources.tab = "cn-shanghai", "Skills"
	for _, page := range []int{1, 2} {
		result, err := readResourcePage(context.Background(), p.services.Usage, p.epoch(), "Skills", p.settings.region, "mounted", "", page)
		p.acceptResourceLoad(p.resources.version, p.epoch(), result, err)
		if err != nil || p.resources.skills == nil || p.resources.skills.PageNumber != page || len(p.resources.mounts.Spaces) != 1 {
			t.Fatal("skill page not loaded", err)
		}
		tt := ui.NewTester(p.renderResources, 1100, 4000)
		for _, label := range []string{"Mounted Skill Space", "Previous skills page", "Next skills page"} {
			if _, ok := tt.Find(label); !ok {
				t.Errorf("missing skill control %q", label)
			}
		}
		want := "Description 0"
		if page == 2 {
			want = "Description 20"
		}
		if !tt.HasText(want) || p.resources.skills.HasMore != (page == 1) {
			t.Fatal("wrong page description or continuation")
		}
	}
	if actions.Load() != 2 {
		t.Fatal("unexpected cloud action count")
	}
}

func TestNativeResourcesLateSaveAndEpochIsolation(t *testing.T) {
	p := nativeResourceController(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"content":"draft","revision":2}`)
	})
	p.resources.agents = usage.AgentsMDDocument{Content: "old", Revision: 1}
	p.resources.agentsEditor = resourceEditor{loaded: true, baseline: "old", draft: "draft"}
	work, err := p.prepareResourceSave("AGENTS.md")
	if err != nil || !p.resources.agentsEditor.needsRead {
		t.Fatal("write authority not consumed before dispatch")
	}
	generation, epoch := p.generation, p.epoch()
	apply, err := work(context.Background())
	p.cancelWork()
	p.complete(generation, epoch, p.services.Platform.Status().Value.Epoch, false, apply, err)
	if p.resources.agents.Revision != 1 || p.resources.agentsEditor.draft != "draft" || !p.resources.agentsEditor.needsRead {
		t.Fatal("late save applied after page cancellation")
	}
	result := resourceLoad{tab: "AGENTS.md", agents: &usage.AgentsMDDocument{Content: "late", Revision: 9}}
	p.acceptResourceLoad(p.resources.version+1, epoch, result, nil)
	p.acceptResourceLoad(p.resources.version, epoch+1, result, nil)
	if p.resources.agentsEditor.draft != "draft" || !p.resources.agentsEditor.needsRead {
		t.Fatal("stale load changed draft or CAS authority")
	}
	p.services.Connection.Client.Disconnect()
	p.complete(p.generation, epoch, p.services.Platform.Status().Value.Epoch, false, apply, nil)
	if p.resources.agentsEditor.loaded || p.resources.profileEditor.loaded || p.resources.skills != nil {
		t.Fatal("connection invalidation retained resource state")
	}
}

func TestNativeResourcesUnknownErrorHasVisibleFallback(t *testing.T) {
	p := nativeResourceController(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("rendering an error used network")
	})
	p.setResourceError("AGENTS.md", &dto.SafeError{Code: "futureResourceError", Message: "Resource read is unavailable."})
	for _, locale := range []string{localization.English, localization.Chinese} {
		localization.SetCurrentLocale(locale)
		tt := ui.NewTester(p.renderResources, 1000, 1500)
		if !tt.HasText("Resource read is unavailable.") {
			t.Fatal("unknown safe error became blank")
		}
		if _, ok := tt.Find(localization.UI("Resource error")); !ok {
			t.Fatal("resource error label was not localized")
		}
	}
}
