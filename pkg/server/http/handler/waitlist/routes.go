package waitlist

import (
	"net/http"
	"yakubu-llc/waitlist/pkg/service"

	"yakubu-llc/waitlist/pkg/server/http/middleware"

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

	humaApi.UseMiddleware(
		func(ctx huma.Context, next func(huma.Context)) {
			middleware.WithUser(humaApi)(ctx, next, service.SupabaseClient, service.Logger)
		},
		func(ctx huma.Context, next func(huma.Context)) {
			middleware.WithAccount(humaApi)(ctx, next, service.AccountService, service.Logger)
		},
	)

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
	}, handler.getByID)

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
	}, handler.update)

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
		OperationID: "delete-email-from-waitlist",
		Method:      http.MethodDelete,
		Path:        "/waitlists/{id}/emails",
		Summary:     "Delete email from a waitlist",
		Description: "Delete email from a waitlist.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, handler.deleteEmail)

	huma.Register(humaApi, huma.Operation{
		OperationID: "update-email-from-waitlist",
		Method:      http.MethodPut,
		Path:        "/waitlists/{id}/emails",
		Summary:     "Update email from a waitlist",
		Description: "Update email from a waitlist.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, handler.updateEmail)

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
	}, handler.getEmailsByWaitlistID)

	huma.Register(humaApi, huma.Operation{
		OperationID: "unsubscribe-from-waitlist",
		Method:      http.MethodPost,
		Path:        "/waitlists/{id}/emails/unsubscribe",
		Summary:     "Unsubscribe from a waitlist",
		Description: "Unsubscribe from a waitlist.",
		Tags:        []string{"Waitlists"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
	}, handler.unsubscribeEmail)

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
	}, handler.getWaitlistAnalytics)

}
