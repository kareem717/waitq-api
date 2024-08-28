package http

import (
	"waitq/api/internal/server/http/handler/account"
	"waitq/api/internal/server/http/handler/billing"
	"waitq/api/internal/server/http/handler/health"
	"waitq/api/internal/server/http/handler/waitlist"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

func (s *Server) routes() chi.Router {
	router := chi.NewMux()

	config := huma.DefaultConfig(s.apiName, s.apiVersion)
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:         "http",
			Scheme:       "bearer",
			BearerFormat: "JWT",
		},
	}

	// Create a new Huma API instance
	humaApi := humachi.New(router, config)

	// Register account routes with Huma
	account.RegisterHumaRoutes(
		s.services,
		humaApi,
		s.logger,
		s.supabaseClient,
	)

	waitlist.RegisterHumaRoutes(
		s.services,
		humaApi,
		s.logger,
		s.supabaseClient,
	)

	billing.RegisterHumaRoutes(
		s.services,
		s.logger,
		humaApi,
		s.stripeWebhookSecret,
		s.supabaseClient,
	)

	health.RegisterHumaRoutes(
		humaApi,
		s.logger,
	)

	return router
}
