package subscription

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"waitq/api/internal/entities/subscription"
	"waitq/api/internal/server/http/handler/shared"
	"waitq/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
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

type PriceIDPathParam struct {
	PriceID string `path:"priceId"`
}

type GetStripeCheckoutLinkInput struct {
	PriceIDPathParam
	RedirectURL string `query:"redirectUrl"`
}

type GetStripeCheckoutLinkOutput struct {
	Body struct {
		Message string `json:"message"`
		Link    string `json:"link"`
	}
}

func (h *httpHandler) getStripeCheckoutLink(ctx context.Context, input *GetStripeCheckoutLinkInput) (*GetStripeCheckoutLinkOutput, error) {
	ctxAccount := shared.GetAuthenticatedAccount(ctx)

	sess, err := h.subscriptionService.CreateStripeCheckoutSession(ctx, input.PriceID, ctxAccount.ID, input.RedirectURL)
	if err != nil {
		h.logger.Error("failed to create stripe checkout session", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to create stripe checkout session")
	}

	resp := &GetStripeCheckoutLinkOutput{}
	resp.Body.Message = "Stripe checkout link created successfully"
	resp.Body.Link = sess.URL

	return resp, nil
}

type HandleStripeSubscriptionCallbackInput struct {
	SessionID string `query:"session_id"`
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
	sub, redirectUrl, err := h.subscriptionService.HandleStripeCheckoutSuccess(ctx, input.SessionID)
	if err != nil {
		h.logger.Error("failed to handle stripe checkout success", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to handle stripe checkout success")
	}

	h.logger.Info("subscription created", zap.Any("subscription", sub))

	resp := &HandleStripeSubscriptionCallbackOutput{}
	resp.Body.Message = "Subscription callback handled successfully"
	resp.RedirectHeader = redirectUrl
	resp.Status = http.StatusSeeOther
	resp.Body.RedirectURL = redirectUrl

	return resp, nil
}

type AccountIDPathParam struct {
	AccountID uuid.UUID `path:"accountId" minLength:"36" maxLength:"36" format:"uuid"`
}

type GetAccountSubscriptionOutput struct {
	Body struct {
		Message      string                            `json:"message"`
		Subscription *subscription.Subscription        `json:"subscription"`
		Relations    *subscription.AccountSubscription `json:"subscriptionRelationship"`
	}
}

func (h *httpHandler) getAccountSubscription(ctx context.Context, input *AccountIDPathParam) (*GetAccountSubscriptionOutput, error) {
	if ctxAccount := shared.GetAuthenticatedAccount(ctx); ctxAccount.ID != input.AccountID {
		h.logger.Error("unauthorized access", zap.Any("accountId", input.AccountID), zap.Any("ctxAccountId", ctxAccount.ID))
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	resp := &GetAccountSubscriptionOutput{}

	sub, err := h.subscriptionService.GetAccountSubscription(ctx, input.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			resp.Body.Message = "Account has no subscription"
			resp.Body.Subscription = nil
			resp.Body.Relations = nil
			return resp, nil
		}
		h.logger.Error("failed to get account subscription", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to get account subscription")
	}

	subRel, err := h.subscriptionService.GetAccountSubscriptionRelationship(ctx, input.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			resp.Body.Message = "Account has no subscription"
			resp.Body.Subscription = nil
			resp.Body.Relations = nil
			return resp, nil
		}
		h.logger.Error("failed to get account subscription", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to get account subscription")
	}

	resp.Body.Message = "Subscription retrieved successfully"
	resp.Body.Subscription = &sub
	resp.Body.Relations = &subRel

	return resp, nil
}

type UpdateAccountSubscriptionInput struct {
	PriceID string `query:"priceId"`
	AccountIDPathParam
}

type UpdateAccountSubscriptionOutput struct {
	Body struct {
		Message      string                            `json:"message"`
		Subscription *subscription.AccountSubscription `json:"subscription"`
	}
}

func (h *httpHandler) updateAccountSubscription(ctx context.Context, input *UpdateAccountSubscriptionInput) (*UpdateAccountSubscriptionOutput, error) {
	if ctxAccount := shared.GetAuthenticatedAccount(ctx); ctxAccount.ID != input.AccountID {
		h.logger.Error("attempted to update subscription for another account", zap.Any("accountId", input.AccountID), zap.Any("ctxAccountId", ctxAccount.ID))
		return nil, huma.Error403Forbidden("Cannot update subscription for another account")
	}

	sub, err := h.subscriptionService.UpdateAccountSubscription(ctx, input.AccountID, input.PriceID)
	if err != nil {
		h.logger.Error("failed to update account subscription", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to update account subscription")
	}

	resp := &UpdateAccountSubscriptionOutput{}
	resp.Body.Message = "Subscription updated successfully"
	resp.Body.Subscription = &sub

	return resp, nil
}

type CancelAccountSubscriptionOutput struct {
	Body struct {
		Message string `json:"message"`
	}
}

func (h *httpHandler) cancelAccountSubscription(ctx context.Context, input *AccountIDPathParam) (*CancelAccountSubscriptionOutput, error) {
	if ctxAccount := shared.GetAuthenticatedAccount(ctx); ctxAccount.ID != input.AccountID {
		h.logger.Error("attempted to cancel subscription for another account", zap.Any("accountId", input.AccountID), zap.Any("ctxAccountId", ctxAccount.ID))
		return nil, huma.Error403Forbidden("Cannot cancel subscription for another account")
	}

	err := h.subscriptionService.CancelAccountSubscription(ctx, input.AccountID)
	if err != nil {
		h.logger.Error("failed to cancel account subscription", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to cancel account subscription")
	}

	resp := &CancelAccountSubscriptionOutput{}
	resp.Body.Message = "Subscription cancelled successfully"

	return resp, nil
}
