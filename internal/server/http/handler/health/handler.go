package health

import (
	"context"
)

type httpHandler struct{}

func newHTTPHandler() *httpHandler {
	return &httpHandler{}
}

type HealthCheckOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

func (h *httpHandler) healthCheck(ctx context.Context, input *struct{}) (*HealthCheckOutput, error) {
	resp := &HealthCheckOutput{}
	resp.Body.Message = "OK"

	return resp, nil
}
