package nativeui

import (
	"strings"

	"github.com/superops-team/OpenMPA/desktop/internal/dto"
	"github.com/superops-team/OpenMPA/desktop/internal/host"
)

func (p *Controller) useInConversation(text string) {
	if strings.TrimSpace(text) == "" || len(text) > 64<<10 || strings.ContainsRune(text, 0) {
		p.err = dto.Failure("invalidInput", "desktop", "never", 0)
		return
	}
	for _, session := range p.currentSessions() {
		if session.ID == p.shell.SelectedSession {
			host.SendUsage(p.host, host.UsageIntent{ConnectionEpoch: p.epoch(), SessionID: session.ID, Text: text})
			return
		}
	}
	p.err = &dto.SafeError{Code: "conversationRequired", Message: "Select a conversation before using this resource.", Plane: "desktop", RetryDisposition: "never"}
}
