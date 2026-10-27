// Package usage adapts existing Runtime resources for the native desktop.
package usage

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/superops-team/OpenMPA/desktop/internal/cloud"
	"github.com/superops-team/OpenMPA/desktop/internal/connection"
	"github.com/superops-team/OpenMPA/desktop/internal/dto"
)

type Service struct {
	conn  *connection.Client
	cloud *cloud.Client
}

func New(conn *connection.Client, platform *cloud.Client) *Service {
	return &Service{conn: conn, cloud: platform}
}

func (s *Service) request(ctx context.Context, epoch uint64, method, path string, body, output any) *dto.SafeError {
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			return dto.Failure("invalidInput", "desktop", "never", 0)
		}
	}
	resp, err := s.conn.Request(ctx, epoch, method, path, data, nil)
	if err != nil {
		if method != "GET" && err.HTTPStatus >= 500 {
			return dto.Failure("outcomeUnknown", err.Plane, "verifyOnly", err.HTTPStatus)
		}
		return err
	}
	defer resp.Body.Close()
	raw, readErr := io.ReadAll(io.LimitReader(resp.Body, (4<<20)+1))
	if !s.conn.Current(epoch) {
		return dto.Failure("staleConnection", "transport", "reload", 0)
	}
	if readErr != nil || len(raw) > 4<<20 || !json.Valid(raw) || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, output) != nil {
		code, retry := "invalidResponse", "readOnly"
		if method != "GET" {
			code, retry = "outcomeUnknown", "verifyOnly"
		}
		return dto.Failure(code, "runtimeIngress", retry, resp.StatusCode)
	}
	return nil
}
