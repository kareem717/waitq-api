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
}
