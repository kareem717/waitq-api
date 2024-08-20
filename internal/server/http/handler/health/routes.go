package health

import (
	"net/http"

	"waitq/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterHumaRoutes(
	humaApi huma.API,
	services *service.Service,
) {
	handler := &httpHandler{
		logger: services.Logger,
	}

	huma.Register(humaApi, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health check",
		Description: "Health check.",
		Tags:        []string{"Health"},
	}, handler.healthCheck)
}
