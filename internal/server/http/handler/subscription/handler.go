package subscription

import (
	"context"
	"net/http"
	"waitq/api/internal/server/http/handler/shared"
	"waitq/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

type httpHandler struct {
	subscriptionService service.SubscriptionService
	logger              *zap.Logger
}

// TODO: handle subscription override
func newHTTPHandler(subscriptionService service.SubscriptionService, logger *zap.Logger) *httpHandler {
	return &httpHandler{
		subscriptionService: subscriptionService,
		logger:              logger,
	}
}

type GetStripeCheckoutLinkInput struct {
	PriceID string `path:"priceId"`
}

type GetStripeCheckoutLinkOutput struct {
	Body struct {
		Message string `json:"message"`
		Link    string `json:"link"`
	}
}

func (h *httpHandler) getStripeCheckoutLink(ctx context.Context, input *GetStripeCheckoutLinkInput) (*GetStripeCheckoutLinkOutput, error) {
	ctxAccount := shared.GetAuthenticatedAccount(ctx)

	sess, err := h.subscriptionService.CreateStripeCheckoutSession(ctx, input.PriceID, ctxAccount.ID)
	if err != nil {
		h.logger.Error("failed to create stripe checkout session", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to create stripe checkout session")
	}

	resp := &GetStripeCheckoutLinkOutput{}
	resp.Body.Message = "Subscription callback handled successfully"
	resp.Body.Link = sess.URL

	return resp, nil
}

type HandleStripeSubscriptionCallbackInput struct {
	RedirectURL string `query:"redirect_url"`
	SessionID   string `query:"session_id"`
	AccountID   string `query:"account_id"`
}

type HandleStripeSubscriptionCallbackOutput struct {
	RedirectHeader string `header:"Location"`
	Status         int
	Body           struct {
		Message     string `json:"message"`
		RedirectURL string `json:"redirect_url"`
	}
}

func (h *httpHandler) handleStripeSubscriptionCallback(ctx context.Context, input *HandleStripeSubscriptionCallbackInput) (*HandleStripeSubscriptionCallbackOutput, error) {
	sub, err := h.subscriptionService.HandleStripeCheckoutSuccess(ctx, input.SessionID)
	if err != nil {
		h.logger.Error("failed to handle stripe checkout success", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to handle stripe checkout success")
	}

	h.logger.Info("subscription created", zap.Any("subscription", sub))

	resp := &HandleStripeSubscriptionCallbackOutput{}
	resp.Body.Message = "Subscription callback handled successfully"
	resp.RedirectHeader = input.RedirectURL
	resp.Status = http.StatusSeeOther
	resp.Body.RedirectURL = input.RedirectURL

	return resp, nil
}
