package waitlist

import (
	"net/http"
	"waitq/api/pkg/service"

	"waitq/api/pkg/server/http/middleware"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterHumaRoutes(
	service *service.Service,
	humaApi huma.API,
) {
	handler := &httpHandler{
		waitlistService: service.WaitlistService,
		logger:          service.Logger,
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
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, service.Logger)
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
				middleware.WithUser(humaApi)(ctx, next, service.SupabaseClient, service.Logger)
			},
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithAccount(humaApi)(ctx, next, service.AccountService, service.Logger)
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
				middleware.WithUser(humaApi)(ctx, next, service.SupabaseClient, service.Logger)
			},
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithAccount(humaApi)(ctx, next, service.AccountService, service.Logger)
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
				middleware.WithUser(humaApi)(ctx, next, service.SupabaseClient, service.Logger)
			},
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithAccount(humaApi)(ctx, next, service.AccountService, service.Logger)
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
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, service.Logger)
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
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, service.Logger)
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
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, service.Logger)
			},
		},
	}, handler.delete)

	huma.Register(humaApi, huma.Operation{
		OperationID: "add-emails-to-waitlist",
		Method:      http.MethodPost,
		Path:        "/waitlists/{id}/emails/create",
		Summary:     "Add emails to a waitlist",
		Description: "Add emails to a waitlist.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, handler.addEmails)

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
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, service.Logger)
			},
		},
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
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, service.Logger)
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
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, service.Logger)
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
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, service.Logger)
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
				middleware.WithWaitlistServiceKey(humaApi)(ctx, next, service.Logger)
			},
		},
	}, handler.exportEmailsToCSV)

}
