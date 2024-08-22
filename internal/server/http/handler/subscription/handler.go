package subscription

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"waitq/api/internal/entities/subscription"
	"waitq/api/internal/server/http/handler/shared"
	"waitq/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/webhook"
	"go.uber.org/zap"
)

type httpHandler struct {
	subscriptionService service.SubscriptionService
	logger              *zap.Logger
	stripeWebhookSecret string
}

// TODO: handle subscription override
func newHTTPHandler(subscriptionService service.SubscriptionService, logger *zap.Logger, stripeWebhookSecret string) *httpHandler {
	return &httpHandler{
		subscriptionService: subscriptionService,
		logger:              logger,
		stripeWebhookSecret: stripeWebhookSecret,
	}
}

type PriceIDPathParam struct {
	PriceID string `path:"priceId"`
}

type GetStripeCheckoutLinkInput struct {
	AccountID uuid.UUID `query:"accountId" minLength:"36" maxLength:"36" format:"uuid"`
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

	if input.AccountID != ctxAccount.ID {
		h.logger.Error("attempted to get checkout link for another account", zap.Any("accountId", input.AccountID), zap.Any("ctxAccountId", ctxAccount.ID))
		return nil, huma.Error403Forbidden("Cannot get checkout link for another account")
	}

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

type GetStripeBillingPortalLinkInput struct {
	AccountIDPathParam
	RedirectURL string `query:"redirectUrl"`
}

type GetStripeBillingPortalLinkOutput struct {
	Body struct {
		Message string `json:"message"`
		Link    string `json:"link"`
	}
}

func (h *httpHandler) getStripeBillingPortalLink(ctx context.Context, input *GetStripeBillingPortalLinkInput) (*GetStripeBillingPortalLinkOutput, error) {
	ctxAccount := shared.GetAuthenticatedAccount(ctx)

	if input.AccountID != ctxAccount.ID {
		h.logger.Error("attempted to get billing portal link for another account", zap.Any("accountId", input.AccountID), zap.Any("ctxAccountId", ctxAccount.ID))
		return nil, huma.Error403Forbidden("Cannot get billing portal link for another account")
	}

	sess, err := h.subscriptionService.CreateStripeBillingPortalSession(ctx, ctxAccount.StripeCustomerID, input.RedirectURL)
	if err != nil {
		h.logger.Error("failed to create stripe billing portal session", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to create stripe billing portal session")
	}

	resp := &GetStripeBillingPortalLinkOutput{}
	resp.Body.Message = "Stripe billing portal link created successfully"
	resp.Body.Link = sess.URL

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

type HandleStripeWebhookInput struct {
	Signature string `header:"Stripe-Signature"`
	Body      json.RawMessage
}

type HandleStripeWebhookOutput struct {
	Status int
}

func (h *httpHandler) handleStripeWebhook(ctx context.Context, input *HandleStripeWebhookInput) (*HandleStripeWebhookOutput, error) {
	const MaxBodyBytes = int64(65536)
	reader := bytes.NewReader(input.Body)
	limitedReader := io.LimitReader(reader, MaxBodyBytes)

	payload, err := io.ReadAll(limitedReader)
	if err != nil {
		h.logger.Error("failed to read webhook body", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to read webhook body")
	}

	event, err := webhook.ConstructEvent(payload, input.Signature, h.stripeWebhookSecret)
	if err != nil {
		h.logger.Error("webhook signature verification failed", zap.Error(err))
		return nil, huma.Error400BadRequest("Webhook signature verification failed")
	}

	switch event.Type {
	case stripe.EventTypeCustomerSubscriptionUpdated:
		eventBody, err := parseStripeWebhook[stripe.Subscription](event)
		if err != nil {
			h.logger.Error("failed to parse webhook json", zap.Error(err), zap.String("eventType", string(event.Type)))
			return nil, huma.Error500InternalServerError("Failed to parse webhook json")
		}

		account, err := h.subscriptionService.GetAccountByCustomerId(ctx, eventBody.Customer.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				h.logger.Error("account subscription relationship not found", zap.Error(err))
				return nil, huma.Error404NotFound("Account subscription relationship not found")
			}
			h.logger.Error("failed to get account by customer id", zap.Error(err))
			return nil, huma.Error500InternalServerError("Failed to get account by customer id")
		}

		_, err = h.subscriptionService.UpdateAccountSubscription(ctx, account.ID, eventBody.Items.Data[0].Price.ID)
		if err != nil {
			h.logger.Error("failed to update account subscription", zap.Error(err))
			return nil, huma.Error500InternalServerError("Failed to update account subscription")
		}

		h.logger.Info("Subscription was updated!", zap.Any("subscription", eventBody))
	case stripe.EventTypeCustomerSubscriptionDeleted:
		eventBody, err := parseStripeWebhook[stripe.Subscription](event)
		if err != nil {
			h.logger.Error("failed to parse webhook json", zap.Error(err), zap.String("eventType", string(event.Type)))
			return nil, huma.Error500InternalServerError("Failed to parse webhook json")
		}

		account, err := h.subscriptionService.GetAccountByCustomerId(ctx, eventBody.Customer.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				h.logger.Error("account subscription relationship not found", zap.Error(err))
				return nil, huma.Error404NotFound("Account subscription relationship not found")
			}
			h.logger.Error("failed to get account by customer id", zap.Error(err))
			return nil, huma.Error500InternalServerError("Failed to get account by customer id")
		}

		err = h.subscriptionService.DeleteAccountSubscriptionRelationship(ctx, account.ID)
		if err != nil {
			h.logger.Error("failed to delete account subscription relationship", zap.Error(err))
			return nil, huma.Error500InternalServerError("Failed to delete account subscription relationship")
		}

		h.logger.Info("Subscription was deleted!", zap.Any("subscription", eventBody))
	case stripe.EventTypeCheckoutSessionCompleted:
		eventBody, err := parseStripeWebhook[stripe.CheckoutSession](event)
		if err != nil {
			h.logger.Error("failed to parse webhook json", zap.Error(err), zap.String("eventType", string(event.Type)))
			return nil, huma.Error500InternalServerError("Failed to parse webhook json")
		}

		sub, err := h.subscriptionService.HandleStripeCheckoutSuccess(ctx, eventBody.ID)
		if err != nil {
			h.logger.Error("failed to handle stripe checkout success", zap.Error(err))
			return nil, huma.Error500InternalServerError("Failed to handle stripe checkout success")
		}

		h.logger.Info("subscription created", zap.Any("subscription", sub))

	default:
		h.logger.Error("unhandled event type", zap.String("eventType", string(event.Type)))
		return nil, huma.Error501NotImplemented("Unhandled event type")
	}

	resp := &HandleStripeWebhookOutput{}
	resp.Status = http.StatusOK
	return resp, nil
}

func parseStripeWebhook[T any](event stripe.Event) (*T, error) {
	var data T
	err := json.Unmarshal(event.Data.Raw, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
