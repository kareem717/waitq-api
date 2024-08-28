package waitlist

import (
	"net/http"
	"waitq/api/internal/service"

	"waitq/api/internal/server/http/middleware"

	"github.com/danielgtaylor/huma/v2"
	"github.com/supabase-community/supabase-go"
	"go.uber.org/zap"
)

func RegisterHumaRoutes(
	service *service.Service,
	humaApi huma.API,
	logger *zap.Logger,
	supabaseClient *supabase.Client,
) {
	handler := &httpHandler{
		waitlistService: service.WaitlistService,
		billingService:  service.BillingService,
		logger:          logger,
	}

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-waitlist-by-id",
		Method:      http.MethodGet,
		Path:        "/waitlists/{id}",
		Summary:     "Get a waitlist by ID",
		Description: "Get a waitlist by ID.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, logger)
			},
		},
	}, handler.getByID)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-waitlist-api-key-by-id",
		Method:      http.MethodGet,
		Path:        "/waitlists/{id}/api",
		Summary:     "Get a waitlist API key by ID",
		Description: "Get a waitlist API key by ID.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithUser(humaApi)(ctx, next, logger, supabaseClient)
			},
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithAccount(humaApi)(ctx, next, logger, service)
			},
		},
	}, handler.getApiKeys)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-waitlists-by-account-id",
		Method:      http.MethodPost,
		Path:        "/waitlists/account/{accountId}",
		Summary:     "Get waitlists by account ID",
		Description: "Get waitlists by account ID.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithUser(humaApi)(ctx, next, logger, supabaseClient)
			},
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithAccount(humaApi)(ctx, next, logger, service)
			},
		},
	}, handler.getByAccountID)

	huma.Register(humaApi, huma.Operation{
		OperationID: "create-waitlist",
		Method:      http.MethodPost,
		Path:        "/waitlists",
		Summary:     "Create a waitlist",
		Description: "Create a waitlist.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithUser(humaApi)(ctx, next, logger, supabaseClient)
			},
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithAccount(humaApi)(ctx, next, logger, service)
			},
		},
	}, handler.create)

	huma.Register(humaApi, huma.Operation{
		OperationID: "update-waitlist",
		Method:      http.MethodPut,
		Path:        "/waitlists/{id}",
		Summary:     "Update a waitlist",
		Description: "Update a waitlist.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, logger)
			},
		},
	}, handler.update)

	huma.Register(humaApi, huma.Operation{
		OperationID: "generate-new-waitlist-jwt-secret",
		Method:      http.MethodPut,
		Path:        "/waitlists/{id}/jwt/generate",
		Summary:     "Generate a new waitlist JWT secret",
		Description: "Generate a new waitlist JWT secret.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, logger)
			},
		},
	}, handler.updateJWTSecret)

	huma.Register(humaApi, huma.Operation{
		OperationID: "delete-waitlist",
		Method:      http.MethodDelete,
		Path:        "/waitlists/{id}",
		Summary:     "Delete a waitlist",
		Description: "Delete a waitlist.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, logger)
			},
		},
	}, handler.delete)

	huma.Register(humaApi, huma.Operation{
		OperationID: "add-email-to-waitlist",
		Method:      http.MethodPost,
		Path:        "/waitlists/{id}/emails/add",
		Summary:     "Add emails to a waitlist",
		Description: "Add emails to a waitlist.",
		Tags:        []string{"Waitlists"},
	}, handler.addEmail)

	huma.Register(humaApi, huma.Operation{
		OperationID: "unsubscribe-from-waitlist",
		Method:      http.MethodPost,
		Path:        "/waitlists/{id}/emails/unsubscribe",
		Summary:     "Unsubscribe from a waitlist",
		Description: "Unsubscribe from a waitlist.",
		Tags:        []string{"Waitlists"},
	}, handler.unsubscribeEmail)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-unsubscribed-email-jwt",
		Method:      http.MethodGet,
		Path:        "/waitlists/{id}/emails/unsubscribed",
		Summary:     "Get unsubscribed email JWT",
		Description: "Get unsubscribed email JWT.",
		Tags:        []string{"Waitlists"},
	}, handler.getUnsubscribedEmailJWT)

	huma.Register(humaApi, huma.Operation{
		OperationID: "delete-email-from-waitlist",
		Method:      http.MethodDelete,
		Path:        "/waitlists/{id}/emails",
		Summary:     "Delete email from a waitlist",
		Description: "Delete email from a waitlist.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, logger)
			},
		},
	}, handler.deleteEmail)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-emails-in-waitlist",
		Method:      http.MethodPost,
		Path:        "/waitlists/{id}/emails",
		Summary:     "Get emails in a waitlist",
		Description: "Get emails in a waitlist.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, logger)
			},
		},
	}, handler.getEmailsByWaitlistID)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-waitlist-analytics",
		Method:      http.MethodGet,
		Path:        "/waitlists/{id}/analytics",
		Summary:     "Get waitlist analytics",
		Description: "Get waitlist analytics.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, logger)
			},
		},
	}, handler.getWaitlistAnalytics)

	huma.Register(humaApi, huma.Operation{
		OperationID: "export-waitlist-emails-to-csv",
		Method:      http.MethodGet,
		Path:        "/waitlists/{id}/emails/export",
		Summary:     "Export waitlist emails to CSV",
		Description: "Export waitlist emails to CSV.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, logger)
			},
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithWaitlistOwnerBilling(humaApi)(ctx, next, logger, service)
			},
		},
	}, handler.exportEmailsToCSV)

	huma.Register(humaApi, huma.Operation{
		OperationID: "check-url-alias-available",
		Method:      http.MethodGet,
		Path:        "/waitlists/public/{urlAlias}/available",
		Summary:     "Check if a URL alias is available",
		Description: "Check if a URL alias is available.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithUser(humaApi)(ctx, next, logger, supabaseClient)
			},
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithAccount(humaApi)(ctx, next, logger, service)
			},
		},
	}, handler.urlAliasAvailable)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-waitlist-by-url-alias",
		Method:      http.MethodGet,
		Path:        "/waitlists/public/{urlAlias}",
		Summary:     "Get a waitlist by URL alias",
		Description: "Get a waitlist by URL alias.",
		Tags:        []string{"Waitlists"},
	}, handler.getWaitlistByURLAlias)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-public-waitlist-main-page",
		Method:      http.MethodPost,
		Path:        "/waitlists/public",
		Summary:     "Get many public waitlists",
		Description: "Get SEO sitemap data.",
		Tags:        []string{"Waitlists"},
	}, handler.getPublicManyWaitlists)

}
