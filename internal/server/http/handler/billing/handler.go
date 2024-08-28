package billing

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"waitq/api/internal/entities/billing"
	"waitq/api/internal/server/http/handler/shared"
	"waitq/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/webhook"
	"go.uber.org/zap"
)

type httpHandler struct {
	billingService      service.BillingService
	logger              *zap.Logger
	stripeWebhookSecret string
}

func newHTTPHandler(billingService service.BillingService, logger *zap.Logger, stripeWebhookSecret string) *httpHandler {
	return &httpHandler{
		billingService:      billingService,
		logger:              logger,
		stripeWebhookSecret: stripeWebhookSecret,
	}
}

type AccountIDPathParam struct {
	AccountID uuid.UUID `path:"accountId" minLength:"36" maxLength:"36" format:"uuid"`
}

type GetStripeCheckoutLinkInput struct {
	AccountIDPathParam
	PriceID     string `path:"priceId"`
	RedirectURL string `query:"redirectUrl"`
}

type LinkOutput struct {
	Body struct {
		Message string `json:"message"`
		Link    string `json:"link"`
	}
}

func (h *httpHandler) getStripeCheckoutLink(ctx context.Context, input *GetStripeCheckoutLinkInput) (*LinkOutput, error) {
	ctxAccount := shared.GetAuthenticatedAccount(ctx)

	if input.AccountID != ctxAccount.ID {
		h.logger.Error("attempted to get checkout link for another account", zap.Any("accountId", input.AccountID), zap.Any("ctxAccountId", ctxAccount.ID))
		return nil, huma.Error403Forbidden("Cannot get checkout link for another account")
	}

	sess, err := h.billingService.CreateStripeCheckoutSession(ctx, input.PriceID, ctxAccount.ID, input.RedirectURL)
	if err != nil {
		h.logger.Error("failed to create stripe checkout session", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to create stripe checkout session")
	}

	resp := &LinkOutput{}
	resp.Body.Message = "Stripe checkout link created successfully"
	resp.Body.Link = sess.URL

	return resp, nil
}

type GetStripeBillingPortalLinkInput struct {
	AccountIDPathParam
	RedirectURL string `query:"redirectUrl"`
}

func (h *httpHandler) getStripeBillingPortalLink(ctx context.Context, input *GetStripeBillingPortalLinkInput) (*LinkOutput, error) {
	ctxAccount := shared.GetAuthenticatedAccount(ctx)

	if input.AccountID != ctxAccount.ID {
		h.logger.Error("attempted to get billing portal link for another account", zap.Any("accountId", input.AccountID), zap.Any("ctxAccountId", ctxAccount.ID))
		return nil, huma.Error403Forbidden("Cannot get billing portal link for another account")
	}

	sub, err := h.billingService.GetAccountSubscriptionRelationship(ctx, ctxAccount.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.logger.Error("account does not have a subscription", zap.Error(err))
			return nil, huma.Error404NotFound("Account does not have a subscription")
		}
		h.logger.Error("failed to get account subscription relationship", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to get account subscription relationship")
	}

	sess, err := h.billingService.CreateStripeBillingPortalSession(ctx, sub.StripeCustomerID, input.RedirectURL)
	if err != nil {
		h.logger.Error("failed to create stripe billing portal session", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to create stripe billing portal session")
	}

	resp := &LinkOutput{}
	resp.Body.Message = "Stripe billing portal link created successfully"
	resp.Body.Link = sess.URL

	return resp, nil
}

type GetAccountSubscriptionOutput struct {
	Body struct {
		Message      string                       `json:"message"`
		Subscription *billing.Subscription        `json:"subscription"`
		Relation     *billing.AccountSubscription `json:"subscriptionRelationship"`
	}
}

func (h *httpHandler) getAccountSubscription(ctx context.Context, input *AccountIDPathParam) (*GetAccountSubscriptionOutput, error) {
	if ctxAccount := shared.GetAuthenticatedAccount(ctx); ctxAccount.ID != input.AccountID {
		h.logger.Error("unauthorized access", zap.Any("accountId", input.AccountID), zap.Any("ctxAccountId", ctxAccount.ID))
		return nil, huma.Error401Unauthorized("Unauthorized")
	}

	resp := &GetAccountSubscriptionOutput{}

	subRel, err := h.billingService.GetAccountSubscriptionRelationship(ctx, input.AccountID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			resp.Body.Message = "Account has no subscription"
			resp.Body.Subscription = nil
			resp.Body.Relation = nil
			return resp, nil
		}
		h.logger.Error("failed to get account subscription", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to get account subscription")
	}

	sub, err := h.billingService.GetAccountSubscription(ctx, input.AccountID)
	if err != nil {
		h.logger.Error("failed to get account subscription", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to get account subscription")
	}

	resp.Body.Message = "Subscription retrieved successfully"
	resp.Body.Subscription = &sub
	resp.Body.Relation = &subRel

	return resp, nil
}

type HandleStripeWebhookInput struct {
	Signature string `header:"Stripe-Signature"`
	Body      json.RawMessage
}

type HandleStripeWebhookOutput struct {
	RedirectURL *string `header:"Location"`
	Status      int
}

func (h *httpHandler) handleStripeWebhook(ctx context.Context, input *HandleStripeWebhookInput) (*HandleStripeWebhookOutput, error) {
	reader := bytes.NewReader(input.Body)

	payload, err := io.ReadAll(reader)
	if err != nil {
		h.logger.Error("failed to read webhook body", zap.Error(err))
		return nil, huma.Error500InternalServerError("Failed to read webhook body")
	}

	event, err := webhook.ConstructEvent(payload, input.Signature, h.stripeWebhookSecret)
	if err != nil {
		h.logger.Error("webhook signature verification failed", zap.Error(err))
		return nil, huma.Error400BadRequest("Webhook signature verification failed")
	}

	resp := &HandleStripeWebhookOutput{}

	switch event.Type {
	case stripe.EventTypeCustomerSubscriptionUpdated:
		eventBody, err := parseStripeWebhook[stripe.Subscription](event)
		if err != nil {
			h.logger.Error("failed to parse webhook json", zap.Error(err), zap.String("eventType", string(event.Type)))
			return nil, huma.Error500InternalServerError("Failed to parse webhook json")
		}

		_, err = h.billingService.UpdateAccountSubscription(ctx, eventBody)
		if err != nil {
			h.logger.Error("failed to update account subscription", zap.Error(err))
			return nil, huma.Error500InternalServerError("Failed to update account subscription")
		}

		h.logger.Info("Subscription was updated!", zap.Any("subscription", eventBody))
		resp.Status = http.StatusOK
	case stripe.EventTypeCustomerSubscriptionDeleted:
		eventBody, err := parseStripeWebhook[stripe.Subscription](event)
		if err != nil {
			h.logger.Error("failed to parse webhook json", zap.Error(err), zap.String("eventType", string(event.Type)))
			return nil, huma.Error500InternalServerError("Failed to parse webhook json")
		}

		err = h.billingService.DeleteAccountSubscriptionByCustomerID(ctx, eventBody.Customer.ID)
		if err != nil {
			h.logger.Error("failed to delete account subscription relationship", zap.Error(err))
			return nil, huma.Error500InternalServerError("Failed to delete account subscription relationship")
		}

		h.logger.Info("Subscription was deleted!", zap.Any("subscription", eventBody))
		resp.Status = http.StatusOK
	case stripe.EventTypeCheckoutSessionCompleted:
		eventBody, err := parseStripeWebhook[stripe.CheckoutSession](event)
		if err != nil {
			h.logger.Error("failed to parse webhook json", zap.Error(err), zap.String("eventType", string(event.Type)))
			return nil, huma.Error500InternalServerError("Failed to parse webhook json")
		}

		sub, redirectURL, err := h.billingService.HandleStripeCheckoutSuccess(ctx, eventBody.ID)
		if err != nil {
			h.logger.Error("failed to handle stripe checkout success", zap.Error(err))
			return nil, huma.Error500InternalServerError("Failed to handle stripe checkout success")
		}

		h.logger.Info("subscription created", zap.Any("subscription", sub))
		resp.RedirectURL = &redirectURL
		resp.Status = http.StatusPermanentRedirect
	default:
		h.logger.Error("unhandled event type", zap.String("eventType", string(event.Type)))
		return nil, huma.Error501NotImplemented("Unhandled event type")
	}

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
