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
) {

	handler := &httpHandler{
		subscriptionService: service.SubscriptionService,
		logger:              service.Logger,
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
		OperationID: "handle-stripe-subscription-callback",
		Method:      http.MethodGet,
		Path:        "/subscriptions/callback",
		Summary:     "Handle a subscription callback",
		Description: "Handle a subscription callback.",
		Tags:        []string{"Subscriptions"},
	}, handler.handleStripeSubscriptionCallback)

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

	huma.Register(humaApi, huma.Operation{
		OperationID: "cancel-account-subscription",
		Method:      http.MethodDelete,
		Path:        "/subscriptions/account/{accountId}",
		Summary:     "Cancel a subscription",
		Description: "Cancel a subscription.",
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
	}, handler.cancelAccountSubscription)

	huma.Register(humaApi, huma.Operation{
		OperationID: "update-account-subscription",
		Method:      http.MethodPut,
		Path:        "/subscriptions/account/{accountId}",
		Summary:     "Update a subscription",
		Description: "Update a subscription.",
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
	}, handler.updateAccountSubscription)
}
