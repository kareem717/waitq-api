package storage

import (
	"context"
	"waitq/api/internal/entities/account"
	"waitq/api/internal/entities/billing"
	"waitq/api/internal/entities/waitlist"
	"waitq/api/internal/storage/postgres/shared"

	"github.com/google/uuid"
)

type AccountRepository interface {
	Create(ctx context.Context, accountParams account.Account) (account.Account, error)
	Delete(ctx context.Context, accountId uuid.UUID) error
	Update(ctx context.Context, accountParams account.Account) (account.Account, error)
	GetByUserId(ctx context.Context, userId uuid.UUID, queryParams shared.GetManyRequest) (account.Account, error)
	GetById(ctx context.Context, accountId uuid.UUID) (account.Account, error)
}

type WaitlistRepository interface {
	Create(ctx context.Context, input waitlist.Waitlist) (waitlist.Waitlist, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, input waitlist.Waitlist) (waitlist.Waitlist, error)
	GetByAccountId(ctx context.Context, accountId uuid.UUID, input shared.PaginationRequest) ([]waitlist.Waitlist, error)
	GetById(ctx context.Context, id uuid.UUID) (waitlist.Waitlist, error)
	AddEmails(ctx context.Context, emails []waitlist.Email) ([]waitlist.Email, error)
	DeleteEmail(ctx context.Context, waitlistId uuid.UUID, email string) error
	UpdateEmail(ctx context.Context, input waitlist.Email) (waitlist.Email, error)
	GetEmailsByWaitlistID(ctx context.Context, waitlistId uuid.UUID, input shared.EmailPaginationRequest) ([]waitlist.Email, error)
	UnsubscribeEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error)
	GetEmailsByWaitlistIDAndEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error)
	GetAnalytics(ctx context.Context, waitlistId uuid.UUID) (waitlist.WaitlistAnalytics, error)
	GetEmailCountByWaitlistID(ctx context.Context, waitlistId uuid.UUID, includeDeleted bool, includeUnsubscribed bool) (int, error)
	GetActiveWaitlistCountByAccountId(ctx context.Context, accountId uuid.UUID) (int, error)
	GetByURLAlias(ctx context.Context, urlAlias string) (waitlist.Waitlist, error)
	IsURLAliasAvailable(ctx context.Context, urlAlias string) (bool, error)
	GetPublicMany(ctx context.Context, input shared.CursorPaginationRequest) ([]waitlist.PublicWaitlist, error)
}

type BillingRepository interface {
	CreateAccountSubscription(ctx context.Context, sub billing.AccountSubscription) (billing.AccountSubscription, error)
	GetRelationshipByAccountId(ctx context.Context, accountId uuid.UUID) (billing.AccountSubscription, error)
	GetByAccountId(ctx context.Context, accountId uuid.UUID) (billing.Subscription, error)
	GetByStripeProductId(ctx context.Context, stripeProductId string) (billing.Subscription, error)
	GetRelationshipByCustomerId(ctx context.Context, customerId string) (billing.AccountSubscription, error)
	DeleteRelationship(ctx context.Context, accountId uuid.UUID) error
	UpdateAccountSubscription(ctx context.Context, accountId uuid.UUID, newSubscriptionId uuid.UUID, newPriceId string) (billing.AccountSubscription, error)
}

type RepositoryProvider interface {
	Account() AccountRepository
	Waitlist() WaitlistRepository
	Billing() BillingRepository
}

type Transaction interface {
	RepositoryProvider
	Commit() error
	Rollback() error
}

type Repository interface {
	RepositoryProvider
	HealthCheck(ctx context.Context) error
	NewTransaction() (Transaction, error)
	RunInTx(ctx context.Context, fn func(ctx context.Context, tx Transaction) error) error
}
