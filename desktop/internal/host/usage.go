package host

import (
	"strings"

	"github.com/egoist/mygo"
)

type UsageIntent struct {
	ConnectionEpoch uint64 `json:"connectionEpoch"`
	SessionID       string `json:"sessionId"`
	Text            string `json:"text"`
}

var UsageRequested = mygo.NewEvent[UsageIntent]("UsageRequested")

// SendUsage inserts a choice in the current conversation. It never runs a turn.
func SendUsage(h *Host, intent UsageIntent) {
	if h == nil || intent.SessionID == "" || len(intent.SessionID) > 256 || len(intent.Text) > 64<<10 || strings.ContainsRune(intent.Text, 0) {
		return
	}
	h.mu.Lock()
	r := h.rail
	h.mu.Unlock()
	if r != nil {
		state := r.snapshot()
		if state.ConnectionEpoch == intent.ConnectionEpoch && state.SelectedSession == intent.SessionID {
			UsageRequested.Emit(r.window, intent)
		}
	}
}
