package billing

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
	logger *zap.Logger,
	humaApi huma.API,
	stripeWebhookSecret string,
	supabaseClient *supabase.Client,
) {
	handler := &httpHandler{
		billingService:      service.BillingService,
		logger:              logger,
		stripeWebhookSecret: stripeWebhookSecret,
	}

	huma.Register(humaApi, huma.Operation{
		OperationID:  "handle-stripe-webhook",
		Method:       http.MethodPost,
		Path:         "/billing/webhook",
		Summary:      "Handle a stripe webhook",
		Description:  "Handle a stripe webhook.",
		Tags:         []string{"Billing"},
		MaxBodyBytes: 64 * 1024,
	}, handler.handleStripeWebhook)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-stripe-billing-portal-link",
		Method:      http.MethodGet,
		Path:        "/billing/account/{accountId}/billing-portal",
		Summary:     "Get a stripe billing portal link for an account",
		Description: "Get a stripe billing portal link for an account.",
		Tags:        []string{"Billing"},
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
	}, handler.getStripeBillingPortalLink)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-account-subscription",
		Method:      http.MethodGet,
		Path:        "/billing/account/{accountId}/checkout/{priceId}",
		Summary:     "Get a stripe checkout link",
		Description: "Get a stripe checkout link.",
		Tags:        []string{"Billing"},
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
	}, handler.getStripeCheckoutLink)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-account-subscription",
		Method:      http.MethodGet,
		Path:        "/billing/account/{accountId}/subscription",
		Summary:     "Get account subscription",
		Description: "Get account subscription.",
		Tags:        []string{"Billing"},
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
	}, handler.getAccountSubscription)
}
