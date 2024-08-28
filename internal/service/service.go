package service

import (
	"context"
	"errors"
	"waitq/api/internal/entities/account"
	"waitq/api/internal/entities/subscription"
	"waitq/api/internal/entities/waitlist"
	"waitq/api/internal/storage/postgres/shared"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v79"
)

var (
	ErrAccountNotFound = errors.New("account not found")
)

type AccountService interface {
	Create(ctx context.Context, input account.Account) (account.Account, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, input account.Account) (account.Account, error)
	GetById(ctx context.Context, id uuid.UUID) (account.Account, error)
	GetByUserId(ctx context.Context, userId uuid.UUID, input shared.GetManyRequest) ([]account.Account, error)
}

type WaitlistService interface {
	Create(ctx context.Context, input waitlist.Waitlist) (waitlist.Waitlist, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, input waitlist.Waitlist) (waitlist.Waitlist, error)
	UpdateJWTSecret(ctx context.Context, id uuid.UUID, secret string) (waitlist.Waitlist, error)
	GetByAccountId(ctx context.Context, accountId uuid.UUID, input shared.PaginationRequest) ([]waitlist.Waitlist, error)
	GetById(ctx context.Context, id uuid.UUID) (waitlist.Waitlist, error)
	GetAnalytics(ctx context.Context, waitlistId uuid.UUID) (waitlist.WaitlistAnalytics, error)
	AddEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error)
	DeleteEmail(ctx context.Context, waitlistId uuid.UUID, email string) error
	UpdateEmail(ctx context.Context, input waitlist.Email) (waitlist.Email, error)
	GetEmailsByWaitlistID(ctx context.Context, waitlistId uuid.UUID, input shared.EmailPaginationRequest) ([]waitlist.Email, error)
	UnsubscribeEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error)
	GetEmailsByWaitlistIDAndEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error)
	ExportEmails(ctx context.Context, waitlistId uuid.UUID) (chan string, chan error)
	GetActiveWaitlistCountByAccountId(ctx context.Context, accountId uuid.UUID) (int, error)
	GetActiveEmailCountByWaitlistId(ctx context.Context, waitlistId uuid.UUID) (int, error)
	GetByURLAlias(ctx context.Context, urlAlias string) (waitlist.Waitlist, error)
	IsURLAliasAvailable(ctx context.Context, urlAlias string) (bool, error)
	GetPublicMany(ctx context.Context, input shared.CursorPaginationRequest) ([]waitlist.PublicWaitlist, error)
}

type SubscriptionService interface {
	CreateStripeCheckoutSession(ctx context.Context, priceId string, accountId uuid.UUID, redirectUrl string) (*stripe.CheckoutSession, error)
	HandleStripeCheckoutSuccess(ctx context.Context, sessionId string) (subscription.AccountSubscription, string, error)
	GetAccountSubscription(ctx context.Context, accountId uuid.UUID) (subscription.Subscription, error)
	GetAccountSubscriptionRelationship(ctx context.Context, accountId uuid.UUID) (subscription.AccountSubscription, error)
	CancelAccountSubscription(ctx context.Context, accountId uuid.UUID) error
	UpdateAccountSubscription(ctx context.Context, subscription *stripe.Subscription) (subscription.AccountSubscription, error)
	GetSubscriptionByWaitlistId(ctx context.Context, waitlistId uuid.UUID) (subscription.Subscription, error)
	CreateStripeBillingPortalSession(ctx context.Context, customerID string, returnURL string) (*stripe.BillingPortalSession, error)
	DeleteAccountSubscription(ctx context.Context, subscription *stripe.Subscription) error
}

// Service storage of all services.
type Service struct {
	AccountService      AccountService
	WaitlistService     WaitlistService
	SubscriptionService SubscriptionService
}
