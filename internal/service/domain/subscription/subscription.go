package subscription

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"waitq/api/internal/entities/subscription"
	"waitq/api/internal/storage"
	stripeClient "waitq/api/pkg/stripe"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v79"
)

type subscriptionService struct {
	repo         storage.Repository
	stripeClient *stripeClient.Client
}

func NewSubscriptionService(
	repo storage.Repository,
	stripeClient *stripeClient.Client,
) *subscriptionService {
	return &subscriptionService{
		repo:         repo,
		stripeClient: stripeClient,
	}
}

func (s *subscriptionService) CreateStripeCheckoutSession(ctx context.Context, priceId string, accountId uuid.UUID, redirectUrl string) (*stripe.CheckoutSession, error) {
	return s.stripeClient.CreateCheckoutSession(priceId, accountId, redirectUrl)
}

func (s *subscriptionService) CreateStripeBillingPortalSession(ctx context.Context, customerID string, returnURL string) (*stripe.BillingPortalSession, error) {
	return s.stripeClient.GetBillingPortalURL(customerID, returnURL)
}

func (s *subscriptionService) HandleStripeCheckoutSuccess(ctx context.Context, sessionId string) (subscription.AccountSubscription, string, error) {
	log.Printf("sessionId: %v", sessionId)
	var sub subscription.AccountSubscription
	var redirectURL string

	session, err := s.stripeClient.GetSession(sessionId)
	if err != nil {
		return sub, redirectURL, err
	}

	redirectURL = session.Metadata[stripeClient.RedirectURLMetadataKey]
	if redirectURL == "" {
		return sub, redirectURL, errors.New("redirect url is empty")
	}

	accountId, err := uuid.Parse(session.Metadata[stripeClient.AccountIDMetadataKey])
	if err != nil {
		return sub, redirectURL, err
	}

	newSub := subscription.AccountSubscription{}
	err = s.repo.RunInTx(ctx, func(ctx context.Context, uow storage.UnitOfWork) error {
		log.Printf("accountId: %v", accountId)
		existingSub, err := uow.Subscription().GetRelationshipByAccountId(ctx, accountId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		log.Printf("existingSub: %+v", existingSub)
		// non empty stripe subscription id means the account already has a subscription
		if existingSub.StripeSubscriptionID != "" {
			return errors.New("account already has a subscription, need to upgrade instead of create new subscription")
		}

		lineItems, err := s.stripeClient.GetSessionLineItemIter(session)
		if err != nil {
			return err
		}

		item := lineItems.LineItem()

		subRecord, err := uow.Subscription().GetByStripeProductId(ctx, item.Price.Product.ID)
		if err != nil {
			return err
		}

		sub = subscription.AccountSubscription{
			AccountID:            accountId,
			SubscriptionID:       subRecord.ID,
			StripeSubscriptionID: session.Subscription.ID,
			StripePriceID:        item.Price.ID,
			StripeCustomerID:     session.Customer.ID,
		}

		log.Printf("sub: %+v", sub)
		newSub, err = uow.Subscription().CreateAccountSubscription(ctx, sub)
		if err != nil {
			return err
		}

		log.Printf("new sub: %+v", newSub)

		return nil
	})

	return newSub, redirectURL, err
}

func (s *subscriptionService) UpdateAccountSubscription(ctx context.Context, sub *stripe.Subscription) (subscription.AccountSubscription, error) {
	var newSub subscription.AccountSubscription

	err := s.repo.RunInTx(ctx, func(ctx context.Context, uow storage.UnitOfWork) error {
		accountSub, err := uow.Subscription().GetRelationshipByCustomerId(ctx, sub.Customer.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return err
			} else {
				return err
			}
		}

		stripeSubIter, err := s.stripeClient.GetCustomerSubscriptions(accountSub.StripeCustomerID)
		if err != nil {
			return err
		}

		if !stripeSubIter.Next() {
			return err
		}

		stripeSub := stripeSubIter.Subscription()

		newStripeSub, err := s.stripeClient.UpdateCustomerSubscription(stripeSub.ID, sub.Items.Data[0].Price.ID)
		if err != nil {
			return err
		}

		price := newStripeSub.Items.Data[0].Price

		dbSub, err := uow.Subscription().GetByStripeProductId(ctx, price.Product.ID)
		if err != nil {
			return err
		}

		newSub, err = uow.Subscription().UpdateAccountSubscription(ctx, accountSub.AccountID, dbSub.ID, price.ID)
		if err != nil {
			return err
		}

		if err := uow.Commit(); err != nil {
			return err
		}

		return nil
	})

	return newSub, err
}

func (s *subscriptionService) GetAccountSubscriptionRelationship(ctx context.Context, accountId uuid.UUID) (subscription.AccountSubscription, error) {
	return s.repo.Subscription().GetRelationshipByAccountId(ctx, accountId)
}

func (s *subscriptionService) GetAccountSubscription(ctx context.Context, accountId uuid.UUID) (subscription.Subscription, error) {
	return s.repo.Subscription().GetByAccountId(ctx, accountId)
}

func (s *subscriptionService) CancelAccountSubscription(ctx context.Context, accountId uuid.UUID) error {
	return s.repo.RunInTx(ctx, func(ctx context.Context, uow storage.UnitOfWork) error {
		// Do this here because idk how to undo the stripe call if it fails
		err := uow.Subscription().DeleteRelationship(ctx, accountId)
		if err != nil {
			return err
		}

		existingSub, err := uow.Subscription().GetRelationshipByAccountId(ctx, accountId)
		if err != nil {
			return err
		}

		_, err = s.stripeClient.CancelSubscription(existingSub.StripeSubscriptionID)
		if err != nil {
			return err
		}

		if err := uow.Commit(); err != nil {
			return err
		}

		return nil
	})
}

func (s *subscriptionService) GetSubscriptionByWaitlistId(ctx context.Context, waitlistId uuid.UUID) (subscription.Subscription, error) {
	waitlist, err := s.repo.Waitlist().GetById(ctx, waitlistId)
	if err != nil {
		return subscription.Subscription{}, err
	}

	return s.repo.Subscription().GetByAccountId(ctx, waitlist.AccountID)

}

func (s *subscriptionService) DeleteAccountSubscription(ctx context.Context, subscription *stripe.Subscription) error {
	accountSub, err := s.repo.Subscription().GetRelationshipByCustomerId(ctx, subscription.Customer.ID)
	if err != nil {
		return err
	}

	return s.repo.Subscription().DeleteRelationship(ctx, accountSub.AccountID)
}
