package host

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestUsagePagesAreReachableFromNativeSidebar(t *testing.T) {
	var action string
	rail := &nativeRail{state: ShellState{View: "chat"}, request: func(next, id string) { action = next }}
	tt := ui.NewTester(rail.sidebar, 240, 700)
	for _, item := range []struct{ label, view string }{
		{"Skills, MCP & instructions", "resources"},
		{"A2A collaboration", "collaboration"},
	} {
		if err := tt.Click(item.label); err != nil {
			t.Fatal(err)
		}
		if action != item.view || !validShellState(ShellState{View: item.view}) {
			t.Fatalf("native page unavailable: %s", item.view)
		}
	}
}
