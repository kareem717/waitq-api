package health

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterHumaRoutes(
	humaApi huma.API,
) {
	handler := &httpHandler{}

	huma.Register(humaApi, huma.Operation{
		OperationID: "health-check",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Health check",
		Description: "Health check.",
		Tags:        []string{"Health"},
	}, handler.healthCheck)
}
