package billing

import (
	"context"

	"waitq/api/internal/entities/billing"
	"waitq/api/internal/storage"
	stripeClient "waitq/api/pkg/stripe"

	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v79"
)

type billingService struct {
	repo         storage.Repository
	stripeClient *stripeClient.Client
}

func NewBillingService(
	repo storage.Repository,
	stripeClient *stripeClient.Client,
) *billingService {
	return &billingService{
		repo:         repo,
		stripeClient: stripeClient,
	}
}

func (s *billingService) CreateStripeCheckoutSession(ctx context.Context, priceId string, accountId uuid.UUID, redirectUrl string) (*stripe.CheckoutSession, error) {
	return s.stripeClient.CreateCheckoutSession(priceId, accountId, redirectUrl)
}

func (s *billingService) CreateStripeBillingPortalSession(ctx context.Context, customerID string, returnURL string) (*stripe.BillingPortalSession, error) {
	return s.stripeClient.GetBillingPortalURL(customerID, returnURL)
}

func (s *billingService) HandleStripeCheckoutSuccess(ctx context.Context, sessionId string) (billing.AccountSubscription, string, error) {
	var sub billing.AccountSubscription
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

	newSub := billing.AccountSubscription{}
	err = s.repo.RunInTx(ctx, func(ctx context.Context, tx storage.Transaction) error {
		existingSub, err := tx.Billing().GetRelationshipByAccountId(ctx, accountId)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		// non empty stripe billing id means the account already has a billing
		if existingSub.StripeSubscriptionID != "" {
			return errors.New("account already has a billing, need to upgrade instead of create new billing")
		}

		lineItems, err := s.stripeClient.GetSessionLineItemIter(session)
		if err != nil {
			return err
		}

		item := lineItems.LineItem()

		subRecord, err := tx.Billing().GetByStripeProductId(ctx, item.Price.Product.ID)
		if err != nil {
			return err
		}

		sub = billing.AccountSubscription{
			AccountID:            accountId,
			SubscriptionID:       subRecord.ID,
			StripeSubscriptionID: session.Subscription.ID,
			StripePriceID:        item.Price.ID,
			StripeCustomerID:     session.Customer.ID,
		}

		newSub, err = tx.Billing().CreateAccountSubscription(ctx, sub)
		if err != nil {
			return err
		}

		return nil
	})

	return newSub, redirectURL, err
}

func (s *billingService) UpdateAccountSubscription(ctx context.Context, sub *stripe.Subscription) (billing.AccountSubscription, error) {
	var newSub billing.AccountSubscription

	err := s.repo.RunInTx(ctx, func(ctx context.Context, tx storage.Transaction) error {
		accountSub, err := tx.Billing().GetRelationshipByCustomerId(ctx, sub.Customer.ID)
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

		dbSub, err := tx.Billing().GetByStripeProductId(ctx, price.Product.ID)
		if err != nil {
			return err
		}

		newSub, err = tx.Billing().UpdateAccountSubscription(ctx, accountSub.AccountID, dbSub.ID, price.ID)
		if err != nil {
			return err
		}

		return nil
	})

	return newSub, err
}

func (s *billingService) GetAccountSubscriptionRelationship(ctx context.Context, accountId uuid.UUID) (billing.AccountSubscription, error) {
	return s.repo.Billing().GetRelationshipByAccountId(ctx, accountId)
}

func (s *billingService) GetAccountSubscription(ctx context.Context, accountId uuid.UUID) (billing.Subscription, error) {
	return s.repo.Billing().GetByAccountId(ctx, accountId)
}

func (s *billingService) CancelAccountSubscription(ctx context.Context, accountId uuid.UUID) error {
	return s.repo.RunInTx(ctx, func(ctx context.Context, tx storage.Transaction) error {
		// Do this here because idk how to undo the stripe call if it fails
		err := tx.Billing().DeleteRelationship(ctx, accountId)
		if err != nil {
			return err
		}

		existingSub, err := tx.Billing().GetRelationshipByAccountId(ctx, accountId)
		if err != nil {
			return err
		}

		_, err = s.stripeClient.CancelSubscription(existingSub.StripeSubscriptionID)
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *billingService) GetSubscriptionByWaitlistId(ctx context.Context, waitlistId uuid.UUID) (billing.Subscription, error) {
	waitlist, err := s.repo.Waitlist().GetById(ctx, waitlistId)
	if err != nil {
		return billing.Subscription{}, err
	}

	return s.repo.Billing().GetByAccountId(ctx, waitlist.AccountID)

}

func (s *billingService) DeleteAccountSubscriptionByCustomerID(ctx context.Context, customerID string) error {
	accountSub, err := s.repo.Billing().GetRelationshipByCustomerId(ctx, customerID)
	if err != nil {
		return err
	}

	return s.repo.Billing().DeleteRelationship(ctx, accountSub.AccountID)
}
