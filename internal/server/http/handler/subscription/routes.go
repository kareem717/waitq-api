package subscription

import (
	"net/http"
	"waitq/api/internal/service"

	"waitq/api/internal/server/http/middleware"

	"github.com/danielgtaylor/huma/v2"
)

func RegisterHumaRoutes(
	service *service.Service,
	humaApi huma.API,
	stripeWebhookSecret string,
) {

	handler := &httpHandler{
		subscriptionService: service.SubscriptionService,
		logger:              service.Logger,
		stripeWebhookSecret: stripeWebhookSecret,
	}

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-stripe-checkout-link",
		Method:      http.MethodGet,
		Path:        "/subscriptions/{priceId}",
		Summary:     "Get a stripe checkout link",
		Description: "Get a stripe checkout link.",
		Tags:        []string{"Subscriptions"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithUser(humaApi)(ctx, next, service)
			},
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithAccount(humaApi)(ctx, next, service)
			},
		},
	}, handler.getStripeCheckoutLink)

	huma.Register(humaApi, huma.Operation{
		OperationID: "handle-stripe-webhook",
		Method:      http.MethodPost,
		Path:        "/subscriptions/webhook",
		Summary:     "Handle a stripe webhook",
		Description: "Handle a stripe webhook.",
		Tags:        []string{"Subscriptions"},
	}, handler.handleStripeWebhook)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-stripe-billing-portal-link",
		Method:      http.MethodGet,
		Path:        "/subscriptions/billing-portal/{accountId}",
		Summary:     "Get a stripe billing portal link",
		Description: "Get a stripe billing portal link.",
		Tags:        []string{"Subscriptions"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithUser(humaApi)(ctx, next, service)
			},
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithAccount(humaApi)(ctx, next, service)
			},
		},
	}, handler.getStripeBillingPortalLink)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-stripe-checkout-link",
		Method:      http.MethodGet,
		Path:        "/subscriptions/{priceId}",
		Summary:     "Get a stripe checkout link",
		Description: "Get a stripe checkout link.",
		Tags:        []string{"Subscriptions"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithUser(humaApi)(ctx, next, service)
			},
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithAccount(humaApi)(ctx, next, service)
			},
		},
	}, handler.getStripeCheckoutLink)

	huma.Register(humaApi, huma.Operation{
		OperationID: "get-account-subscription",
		Method:      http.MethodGet,
		Path:        "/subscriptions/account/{accountId}",
		Summary:     "Get a subscription",
		Description: "Get a subscription.",
		Tags:        []string{"Subscriptions"},
		Security: []map[string][]string{
			{"bearerAuth": {}},
		},
		Middlewares: huma.Middlewares{
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithUser(humaApi)(ctx, next, service)
			},
			func(ctx huma.Context, next func(huma.Context)) {
				middleware.WithAccount(humaApi)(ctx, next, service)
			},
		},
	}, handler.getAccountSubscription)

}
