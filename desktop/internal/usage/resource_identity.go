package usage

import (
	"context"
	"net/http"
	"strings"

	"github.com/superops-team/OpenMPA/desktop/internal/dto"
)

// ResourceInstanceID reads the current Runtime's identity, not its platform ID.
// Studio agent-info does not currently publish an identity field.
func (s *Service) ResourceInstanceID(ctx context.Context, epoch uint64) (string, *dto.SafeError) {
	var wire struct {
		MPAInstanceID string `json:"mpaInstanceId"`
	}
	if err := s.request(ctx, epoch, http.MethodGet, "/api/v1/agents", nil, &wire); err != nil {
		return "", err
	}
	if !resourceIdentifierValid(wire.MPAInstanceID) || strings.HasPrefix(strings.ToLower(wire.MPAInstanceID), "r-") {
		return "", resourceResponseError(false)
	}
	if err := s.resourceEpochError(epoch); err != nil {
		return "", err
	}
	return wire.MPAInstanceID, nil
}
