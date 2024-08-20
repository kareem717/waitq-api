package health

import (
	"context"

	"go.uber.org/zap"
)

type httpHandler struct {
	logger *zap.Logger
}

func newHTTPHandler(logger *zap.Logger) *httpHandler {
	return &httpHandler{
		logger: logger,
	}
}

type HealthCheckOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

func (h *httpHandler) healthCheck(ctx context.Context, input *struct{}) (*HealthCheckOutput, error) {
	h.logger.Info("healthy :)")

	resp := &HealthCheckOutput{}
	resp.Body.Message = "OK"

	return resp, nil
}
