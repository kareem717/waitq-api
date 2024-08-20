package subscription

import (
	"context"
	"database/sql"
	"errors"

	"waitq/api/internal/entities/subscription"
	"waitq/api/internal/storage"
	stripeClient "waitq/api/pkg/stripe"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v79"
	"go.uber.org/zap"
)

type subscriptionService struct {
	subscriptionRepository storage.SubscriptionRepository
	waitlistRepository     storage.WaitlistRepository
	stripeClient           *stripeClient.Client
	logger                 *zap.Logger
}

func NewSubscriptionService(subscriptionRepository storage.SubscriptionRepository, waitlistRepository storage.WaitlistRepository, stripeClient *stripeClient.Client, logger *zap.Logger) *subscriptionService {
	return &subscriptionService{
		subscriptionRepository: subscriptionRepository,
		waitlistRepository:     waitlistRepository,
		stripeClient:           stripeClient,
		logger:                 logger,
	}
}

func (s *subscriptionService) CreateStripeCheckoutSession(ctx context.Context, priceId string, accountId uuid.UUID, redirectUrl string) (*stripe.CheckoutSession, error) {
	return s.stripeClient.CreateCheckoutSession(priceId, accountId.String(), redirectUrl)
}

func (s *subscriptionService) HandleStripeCheckoutSuccess(ctx context.Context, sessionId string) (subscription.AccountSubscription, string, error) {
	session, err := s.stripeClient.GetSession(sessionId)
	if err != nil {
		s.logger.Error("failed to get checkout session", zap.Error(err))
		return subscription.AccountSubscription{}, "", err
	}

	customerAccountId := session.Metadata["customer_account_id"]
	redirectUrl := session.Metadata["redirect_url"]
	parsedAccountId, err := uuid.Parse(customerAccountId)
	if err != nil {
		s.logger.Error("failed to parse account id during checkout success", zap.Error(err))
		return subscription.AccountSubscription{}, "", err
	}

	existingSub, err := s.subscriptionRepository.GetRelationshipByAccountId(ctx, parsedAccountId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Info("account does not have a subscription, creating new subscription", zap.String("account_id", parsedAccountId.String()))
		} else {
			s.logger.Error("failed to get account subscription", zap.Error(err))
			return subscription.AccountSubscription{}, "", err
		}
	}

	if existingSub.StripeSubscriptionID != "" {
		s.logger.Error("account already has a subscription, need to upgrade instead of create new subscription", zap.String("account_id", parsedAccountId.String()))
		return subscription.AccountSubscription{}, "", errors.New("account already has a subscription, need to upgrade instead of create new subscription")
	}

	lineItems, err := s.stripeClient.GetSessionLineItemIter(session)
	if err != nil {
		s.logger.Error("failed to get session line item iter", zap.Error(err))
		return subscription.AccountSubscription{}, "", err
	}

	item := lineItems.LineItem()

	subRecord, err := s.subscriptionRepository.GetByStripeProductId(ctx, item.Price.Product.ID)
	if err != nil {
		s.logger.Error("failed to get subscription record", zap.Error(err))
		return subscription.AccountSubscription{}, "", err
	}

	newSubParams := subscription.AccountSubscription{
		AccountID:            parsedAccountId,
		SubscriptionID:       subRecord.ID,
		StripeSubscriptionID: session.Subscription.ID,
		StripeCustomerID:     session.Customer.ID,
		StripePriceID:        item.Price.ID,
	}

	newSub, err := s.subscriptionRepository.CreateAccountSubscription(ctx, newSubParams)
	if err != nil {
		s.logger.Error("failed to create account subscription", zap.Error(err))
		return subscription.AccountSubscription{}, "", err
	}

	return newSub, redirectUrl, nil
}

func (s *subscriptionService) UpdateAccountSubscription(ctx context.Context, accountId uuid.UUID, newPriceId string) (subscription.AccountSubscription, error) {
	existingSub, err := s.subscriptionRepository.GetRelationshipByAccountId(ctx, accountId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("account does not have a subscription, cannot update", zap.String("account_id", accountId.String()))
			return subscription.AccountSubscription{}, err
		} else {
			s.logger.Error("failed to get account subscription", zap.Error(err))
			return subscription.AccountSubscription{}, err
		}
	}

	stripeSubIter, err := s.stripeClient.GetCustomerSubscriptions(existingSub.StripeCustomerID)
	if err != nil {
		s.logger.Error("failed to get customer subscriptions", zap.Error(err))
		return subscription.AccountSubscription{}, err
	}

	if !stripeSubIter.Next() {
		s.logger.Error("no subscriptions found for customer", zap.String("customer_id", existingSub.StripeCustomerID))
		return subscription.AccountSubscription{}, err
	}

	stripeSub := stripeSubIter.Subscription()

	newStripeSub, err := s.stripeClient.UpdateCustomerSubscription(stripeSub.ID, newPriceId)
	if err != nil {
		s.logger.Error("faile	d to update customer subscription", zap.Error(err))
		return subscription.AccountSubscription{}, err
	}

	price := newStripeSub.Items.Data[0].Price

	newSub, err := s.subscriptionRepository.GetByStripeProductId(ctx, price.Product.ID)
	if err != nil {
		s.logger.Error("failed to create account subscription", zap.Error(err))
	}

	newAccSub, err := s.subscriptionRepository.UpdateAccountSubscription(ctx, existingSub.AccountID, newSub.ID, price.ID)
	if err != nil {
		s.logger.Error("failed to create account subscription", zap.Error(err))
	}

	return newAccSub, err
}

func (s *subscriptionService) GetAccountSubscriptionRelationship(ctx context.Context, accountId uuid.UUID) (subscription.AccountSubscription, error) {
	return s.subscriptionRepository.GetRelationshipByAccountId(ctx, accountId)
}

func (s *subscriptionService) GetAccountSubscription(ctx context.Context, accountId uuid.UUID) (subscription.Subscription, error) {
	return s.subscriptionRepository.GetByAccountId(ctx, accountId)
}

func (s *subscriptionService) CancelAccountSubscription(ctx context.Context, accountId uuid.UUID) error {
	existingSub, err := s.subscriptionRepository.GetRelationshipByAccountId(ctx, accountId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("account does not have a subscription to cancel", zap.String("account_id", accountId.String()))
			return err
		} else {
			s.logger.Error("failed to get account subscription", zap.Error(err))
			return err
		}
	}

	_, err = s.stripeClient.CancelSubscription(existingSub.StripeSubscriptionID)
	if err != nil {
		s.logger.Error("failed to cancel subscription", zap.Error(err))
		return err
	}

	// Optionally, update your local subscription record to reflect the cancellation
	err = s.subscriptionRepository.DeleteRelationship(ctx, accountId)
	if err != nil {
		s.logger.Error("failed to delete account subscription relationship", zap.Error(err))
		return err
	}

	return nil
}

func (s *subscriptionService) GetSubscriptionByWaitlistId(ctx context.Context, waitlistId uuid.UUID) (subscription.Subscription, error) {
	waitlist, err := s.waitlistRepository.GetById(ctx, waitlistId)
	if err != nil {
		return subscription.Subscription{}, err
	}

	sub, err := s.subscriptionRepository.GetByAccountId(ctx, waitlist.AccountID)

	return sub, nil
}
