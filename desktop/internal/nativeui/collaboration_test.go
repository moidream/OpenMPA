package nativeui

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/superops-team/OpenMPA/desktop/internal/connection"
	"github.com/superops-team/OpenMPA/desktop/internal/dto"
	"github.com/superops-team/OpenMPA/desktop/internal/host"
	"github.com/superops-team/OpenMPA/desktop/internal/usage"
)

func collaborationTestController(t *testing.T) *Controller {
	t.Helper()
	p := testController(t)
	_, err := p.services.Connection.Client.Bind(connection.Binding{Endpoint: "https://example.com", RuntimeID: "r-self", RuntimeKey: "key", Scope: "test"}, false)
	if err != nil {
		t.Fatal(err)
	}
	p.services.Usage = usage.New(p.services.Connection.Client, nil)
	id, alias := "as-team", "Writer"
	p.collaboration = collaborationState{
		epoch: p.epoch(), version: 1, spaceID: id, bindingFresh: true,
		binding: &usage.CollaborationBinding{SpaceID: &id, Region: "cn-beijing", Revision: 3},
		spaces: &usage.CollaborationSpacePage{
			Items: []usage.CollaborationSpace{{ID: id, Name: "Team"}}, Total: 51, Page: 1, PageSize: 50,
		},
		agents: &usage.CollaborationAgentPage{
			Items: []usage.CollaborationAgent{{ID: "a-writer", Name: "Writer", Description: "Writes reports", Status: "running",
				Alias: &alias, AliasAvailable: true, AliasRevision: 4}},
			Total: 51, Page: 1, PageSize: 50,
		},
	}
	return p
}

func TestNativeCollaborationSelectionAndReviewDoNotWrite(t *testing.T) {
	p := collaborationTestController(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(500)
	}))
	defer server.Close()
	_, err := p.services.Connection.Client.Bind(connection.Binding{Endpoint: server.URL, RuntimeID: "r-self", RuntimeKey: "key", Scope: "test"}, true)
	if err != nil {
		t.Fatal(err)
	}
	p.collaboration.epoch = p.epoch()
	tt := ui.NewTester(func(c *ui.Context) {
		p.renderCollaboration(c)
		p.RenderOverlay(c, host.ShellState{View: "collaboration"})
	}, 1000, 1500)
	for _, label := range []string{"Collaboration space", "Next spaces page", "Next members page", "Refresh collaboration"} {
		if _, ok := tt.Find(label); !ok {
			t.Fatalf("%s missing", label)
		}
	}
	if err := tt.Click("Select member"); err != nil {
		t.Fatal(err)
	}
	if p.collaboration.agentID != "a-writer" || !tt.HasText("Writes reports") || !tt.HasText("Agent Card unavailable in this Runtime.") {
		t.Fatal("member details missing after selection")
	}
	if err := tt.Click("Member alias"); err != nil {
		t.Fatal(err)
	}
	tt.Key(ui.Super, ui.KeyA)
	tt.Type("New Writer")
	if p.collaboration.draft().alias != "New Writer" || !p.collaboration.draft().aliasDirty {
		t.Fatalf("alias draft not tracked: %+v", p.collaboration.draft())
	}
	for _, label := range []string{"Save member alias", "Bind selected space", "Unbind space"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
		if !p.confirm || p.confirmed == nil || requests.Load() != 0 {
			t.Fatal("review was skipped or performed a request")
		}
		if err := tt.Click("Cancel"); err != nil {
			t.Fatal(err)
		}
	}
	if requests.Load() != 0 || p.collaboration.draft().alias != "New Writer" {
		t.Fatal("selection or canceled review wrote or discarded the draft")
	}
}

func TestNativeCollaborationReadRefreshAndPaginationAreReadOnly(t *testing.T) {
	p := testController(t)
	var reads, writes atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writes.Add(1)
			w.WriteHeader(500)
			return
		}
		reads.Add(1)
		switch r.URL.Path {
		case "/api/v1/studio/collaborators/registry/binding":
			io.WriteString(w, `{"space_id":"as-team","region":"cn-beijing","revision":3}`)
		case "/api/v1/studio/collaborators/registry/spaces":
			if r.URL.Query().Get("page") != "2" {
				t.Error("space pagination lost")
			}
			io.WriteString(w, `{"items":[{"Id":"as-team","Name":"Team"}],"total":51,"page":2,"page_size":50}`)
		case "/api/v1/studio/collaborators/registry/agents":
			if r.URL.Query().Get("space_id") != "as-team" || r.URL.Query().Get("page") != "3" {
				t.Error("agent pagination or scope lost")
			}
			io.WriteString(w, `{"items":[{"Id":"a-writer","Name":"Writer","Description":"Reports","Status":"running","Alias":"Latest","AliasRevision":6,"AliasAvailable":true}],"total":101,"page":3,"page_size":50}`)
		default:
			t.Error(r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	epoch, err := p.services.Connection.Client.Bind(connection.Binding{Endpoint: server.URL, RuntimeKey: "key"}, true)
	if err != nil {
		t.Fatal(err)
	}
	p.services.Usage = usage.New(p.services.Connection.Client, nil)
	p.collaboration = collaborationState{spaceID: "as-team", agentID: "a-writer", version: 2, needsReload: true}
	draft := p.collaboration.draft()
	draft.alias, draft.task, draft.aliasDirty = "Unsaved alias", "Unsaved task", true
	snapshot, err := readCollaboration(context.Background(), p.services.Usage, epoch, "", 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	p.acceptCollaborationLoad(2, epoch, snapshot)
	if p.collaboration.needsReload || p.collaboration.spaceID != "as-team" ||
		p.collaboration.spaces.Page != 2 || p.collaboration.agents.Page != 3 ||
		p.collaboration.member().AliasRevision != 6 || *p.collaboration.member().Alias != "Latest" {
		t.Fatalf("refresh did not load authoritative versions: %+v", p.collaboration)
	}
	if draft.alias != "Unsaved alias" || draft.task != "Unsaved task" || !draft.aliasDirty || writes.Load() != 0 || reads.Load() != 4 {
		t.Fatal("refresh wrote remotely or lost a draft")
	}
	p.observed = true
	p.shell = host.ShellState{View: "collaboration"}
	p.connectionEpoch, p.platformEpoch = p.epoch(), p.services.Platform.Status().Value.Epoch
	p.observe(host.ShellState{View: "chat"})
	p.observe(host.ShellState{View: "collaboration"})
	p.acceptCollaborationLoad(2, epoch, snapshot)
	if p.collaboration.draft().alias != "Unsaved alias" || p.collaboration.draft().task != "Unsaved task" {
		t.Fatal("navigation/re-entry clobbered a draft")
	}
}

func TestNativeCollaborationRejectsMixedBindingSnapshot(t *testing.T) {
	p := testController(t)
	var bindings atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/binding"):
			fmt.Fprintf(w, `{"space_id":"as-team","region":"cn-beijing","revision":%d}`, bindings.Add(1))
		case strings.HasSuffix(r.URL.Path, "/spaces"):
			io.WriteString(w, `{"items":[],"total":0,"page":1,"page_size":50}`)
		case strings.HasSuffix(r.URL.Path, "/agents"):
			io.WriteString(w, `{"items":[],"total":0,"page":1,"page_size":50}`)
		}
	}))
	defer server.Close()
	epoch, err := p.services.Connection.Client.Bind(connection.Binding{Endpoint: server.URL}, true)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := readCollaboration(context.Background(), usage.New(p.services.Connection.Client, nil), epoch, "", 1, 1)
	if snapshot == nil || snapshot.complete || err == nil || err.Code != "conflict" {
		t.Fatalf("mixed snapshot accepted: %+v %v", snapshot, err)
	}
}

func TestNativeCollaborationInaccessibleSpaceStillAllowsExplicitUnbind(t *testing.T) {
	p := testController(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/binding"):
			io.WriteString(w, `{"space_id":"as-deleted","region":"cn-beijing","revision":9}`)
		case strings.HasSuffix(r.URL.Path, "/spaces"):
			io.WriteString(w, `{"items":[],"total":0,"page":1,"page_size":50}`)
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	epoch, err := p.services.Connection.Client.Bind(connection.Binding{Endpoint: server.URL}, true)
	if err != nil {
		t.Fatal(err)
	}
	p.services.Usage = usage.New(p.services.Connection.Client, nil)
	snapshot, err := readCollaboration(context.Background(), p.services.Usage, epoch, "", 1, 1)
	if err == nil || snapshot == nil || snapshot.complete {
		t.Fatalf("inaccessible space returned success: %+v %v", snapshot, err)
	}
	p.acceptCollaborationLoad(0, epoch, snapshot)
	if p.collaborationReady() || !p.collaborationBindingReady() || p.collaboration.binding.Revision != 9 {
		t.Fatal("failed member read lost the binding or enabled member writes")
	}
	p.reviewCollaborationBinding(true)
	if !p.confirm || p.confirmed == nil || !strings.Contains(p.confirmMessage, "as-deleted") {
		t.Fatal("inaccessible binding cannot be explicitly reviewed for removal")
	}
}

func TestNativeCollaborationFailuresAndStaleResultsRetainDraft(t *testing.T) {
	for _, code := range []string{"conflict", "outcomeUnknown", "authRequired", "invalidResponse"} {
		t.Run(code, func(t *testing.T) {
			p := collaborationTestController(t)
			p.collaboration.agentID = "a-writer"
			draft := p.collaboration.draft()
			draft.alias, draft.task, draft.aliasDirty = "Unsaved", "Task", true
			p.collaboration.needsReload = true
			p.complete(p.generation, p.epoch(), p.services.Platform.Status().Value.Epoch, false, nil, dto.Failure(code, "runtimeIngress", "reload", 409))
			if draft.alias != "Unsaved" || draft.task != "Task" || p.collaborationReady() {
				t.Fatal("failed operation lost the draft or allowed an immediate retry")
			}
			p.reviewCollaborationAlias()
			if p.confirm || p.confirmed != nil {
				t.Fatal("failed write could be resubmitted without a read")
			}
			snapshot := &collaborationSnapshot{binding: p.collaboration.binding, spaces: p.collaboration.spaces,
				agents: p.collaboration.agents, spaceID: "as-other"}
			p.acceptCollaborationLoad(0, p.epoch(), snapshot)
			p.acceptCollaborationLoad(p.collaboration.version, p.epoch()+1, snapshot)
			if p.collaboration.spaceID != "as-team" || draft.alias != "Unsaved" {
				t.Fatal("stale result changed the selected space")
			}
		})
	}
}

func TestNativeCollaborationDraftsStayScopedAndInvalidateWithIdentity(t *testing.T) {
	p := collaborationTestController(t)
	s := &p.collaboration
	s.agentID = "a-writer"
	s.draft().alias, s.draft().aliasDirty, s.draft().task = "Team draft", true, "Team task"
	s.spaceID = "as-other"
	if s.draft().alias == "Team draft" || s.draft().task != "" {
		t.Fatal("draft leaked across spaces")
	}
	s.draft().alias = "Other draft"
	s.spaceID = "as-team"
	if s.draft().alias != "Team draft" || s.draft().task != "Team task" {
		t.Fatal("selection discarded the previous draft")
	}
	p.invalidate()
	if p.collaboration.binding != nil || p.collaboration.agents != nil || p.collaboration.drafts != nil {
		t.Fatal("identity invalidation retained collaboration state")
	}
}

func TestNativeCollaborationUseRequiresBoundRunningNonSelfMemberAndSession(t *testing.T) {
	for _, kind := range []string{"unbound", "other-space", "stopped", "self", "stale", "no-session", "bad-task"} {
		t.Run(kind, func(t *testing.T) {
			p := collaborationTestController(t)
			s := &p.collaboration
			s.agentID = "a-writer"
			s.draft().task = "Summarize"
			switch kind {
			case "unbound":
				s.binding.SpaceID = nil
			case "other-space":
				other := "as-other"
				s.binding.SpaceID = &other
			case "stopped":
				s.member().Status = "stopped"
			case "self":
				s.member().RuntimeID = "r-self"
			case "stale":
				s.epoch++
			case "bad-task":
				s.draft().task = "@other task"
			}
			p.useCollaborationMember()
			if p.err == nil || s.draft().task == "" || p.busy {
				t.Fatal("invalid target accepted, draft cleared, or a Run started")
			}
		})
	}
}

func TestNativeCollaborationStaleConfirmationCannotWrite(t *testing.T) {
	p := collaborationTestController(t)
	p.collaboration.agentID = "a-writer"
	p.reviewCollaborationAlias()
	confirmed := p.confirmed
	if confirmed == nil {
		t.Fatal("review missing")
	}
	p.collaboration.version++
	confirmed()
	if p.err == nil || p.busy {
		t.Fatal("stale confirmation started a request")
	}
}

func TestNativeCollaborationWriteButtonsRemainDisabledUntilRead(t *testing.T) {
	p := collaborationTestController(t)
	s := &p.collaboration
	s.agentID, s.bindingFresh, s.needsReload = "a-writer", false, true
	s.draft().alias, s.draft().aliasDirty = "Unsaved alias", true
	tt := ui.NewTester(p.renderCollaboration, 1000, 1500)
	for _, label := range []string{"Bind selected space", "Unbind space", "Save member alias", "Use in conversation"} {
		if err := tt.Click(label); err != nil {
			t.Fatal(err)
		}
		if p.confirm || p.busy || p.err != nil {
			t.Fatalf("%s was not disabled", label)
		}
	}
	snapshot := &collaborationSnapshot{binding: s.binding, spaces: s.spaces, agents: s.agents, spaceID: s.spaceID, complete: true}
	p.acceptCollaborationLoad(s.version, p.epoch(), snapshot)
	tt.Frame()
	if err := tt.Click("Save member alias"); err != nil {
		t.Fatal(err)
	}
	if !p.confirm || s.draft().alias != "Unsaved alias" {
		t.Fatal("successful read did not re-enable review with the original draft")
	}
}

func TestNativeCollaborationRequiresSavedAliasForSelectedMember(t *testing.T) {
	p := collaborationTestController(t)
	p.collaboration.agentID = "a-writer"
	member := p.collaboration.member()
	member.Alias = nil
	p.collaboration.draft().task = "Summarize the report"
	tt := ui.NewTester(p.renderCollaboration, 1000, 1500)
	if !tt.HasText("Save a unique member alias before using this collaborator.") {
		t.Fatal("missing alias recovery instruction")
	}
	if err := tt.Click("Use in conversation"); err != nil {
		t.Fatal(err)
	}
	if p.busy || p.confirm || p.err != nil {
		t.Fatal("unaddressable member action was enabled")
	}
	p.useCollaborationMember()
	if p.err == nil || p.err.Code != "collaborationAliasRequired" {
		t.Fatal("display name accepted as selected member address")
	}
}
