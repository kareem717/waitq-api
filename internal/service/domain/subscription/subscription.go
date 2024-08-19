package subscription

import (
	"context"

	"waitq/api/internal/entities/subscription"
	"waitq/api/internal/storage"
	stripeClient "waitq/api/pkg/stripe"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v79"
	"go.uber.org/zap"
)

type subscriptionService struct {
	subscriptionRepository storage.SubscriptionRepository
	stripeClient           *stripeClient.Client
	logger                 *zap.Logger
}

func NewSubscriptionService(subscriptionRepository storage.SubscriptionRepository, stripeClient *stripeClient.Client, logger *zap.Logger) *subscriptionService {
	return &subscriptionService{
		subscriptionRepository: subscriptionRepository,
		stripeClient:           stripeClient,
		logger:                 logger,
	}
}

func (s *subscriptionService) CreateStripeCheckoutSession(ctx context.Context, priceId string, accountId uuid.UUID) (*stripe.CheckoutSession, error) {
	return s.stripeClient.CreateCheckoutSession(priceId, accountId.String())
}

func (s *subscriptionService) HandleStripeCheckoutSuccess(ctx context.Context, sessionId string) (subscription.AccountSubscription, error) {
	session, err := s.stripeClient.GetCheckoutSession(sessionId)
	if err != nil {
		s.logger.Error("failed to get checkout session", zap.Error(err))
		return subscription.AccountSubscription{}, err
	}

	customerAccountId := session.Metadata["customer_account_id"]
	parsedAccountId, err := uuid.Parse(customerAccountId)
	if err != nil {
		s.logger.Error("failed to parse account id during checkout success", zap.Error(err))
		return subscription.AccountSubscription{}, err
	}

	lineItems, err := s.stripeClient.GetSessionLineItemIter(session)
	if err != nil {
		s.logger.Error("failed to get session line item iter", zap.Error(err))
		return subscription.AccountSubscription{}, err
	}

	item := lineItems.LineItem()

	subRecord, err := s.subscriptionRepository.GetByStripeProductId(ctx, item.Price.Product.ID)
	if err != nil {
		s.logger.Error("failed to get subscription record", zap.Error(err))
		return subscription.AccountSubscription{}, err
	}

	sub := subscription.AccountSubscription{
		AccountID:            parsedAccountId,
		SubscriptionID:       subRecord.ID,
		StripeSubscriptionID: session.Subscription.ID,
		StripeCustomerID:     session.Customer.ID,
	}

	accSub, err := s.subscriptionRepository.UpdateAccountSubscription(ctx, sub)
	if err != nil {
		s.logger.Error("failed to update account subscription", zap.Error(err))
		return subscription.AccountSubscription{}, err
	}

	return accSub, nil
}

func (s *subscriptionService) GetAccountSubscription(ctx context.Context, accountId uuid.UUID) (subscription.Subscription, error) {
	return s.subscriptionRepository.GetByAccountId(ctx, accountId)
}

func (s *subscriptionService) IsProProduct(productID string) bool {
	return s.stripeClient.IsProProduct(productID)
}

func (s *subscriptionService) IsEntrepreneurProduct(productID string) bool {
	return s.stripeClient.IsEntrepreneurProduct(productID)
}