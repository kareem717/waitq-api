package storage

import (
	"context"
	"waitq/api/internal/entities/account"
	"waitq/api/internal/entities/subscription"
	"waitq/api/internal/entities/waitlist"
	"waitq/api/internal/storage/postgres/shared"

	"github.com/google/uuid"
)

type AccountRepository interface {
	Create(ctx context.Context, account account.Account) (account.Account, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Update(ctx context.Context, id uuid.UUID, input account.Account) (account.Account, error)
	GetByUserId(ctx context.Context, userId uuid.UUID, input shared.GetManyRequest) ([]account.Account, error)
	GetById(ctx context.Context, id uuid.UUID) (account.Account, error)
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

type SubscriptionRepository interface {
	CreateAccountSubscription(ctx context.Context, sub subscription.AccountSubscription) (subscription.AccountSubscription, error)
	GetRelationshipByAccountId(ctx context.Context, accountId uuid.UUID) (subscription.AccountSubscription, error)
	GetByAccountId(ctx context.Context, accountId uuid.UUID) (subscription.Subscription, error)
	GetByStripeProductId(ctx context.Context, stripeProductId string) (subscription.Subscription, error)
	GetRelationshipByCustomerId(ctx context.Context, customerId string) (subscription.AccountSubscription, error)
	DeleteRelationship(ctx context.Context, accountId uuid.UUID) error
	UpdateAccountSubscription(ctx context.Context, accountId uuid.UUID, newSubscriptionId uuid.UUID, newPriceId string) (subscription.AccountSubscription, error)
}

type RepositoryProvider interface {
	Account() AccountRepository
	Waitlist() WaitlistRepository
	Subscription() SubscriptionRepository
}

type UnitOfWork interface {
	RepositoryProvider
	Commit() error
	Rollback() error
}

type Repository interface {
	RepositoryProvider
	HealthCheck(ctx context.Context) error
	NewUnitOfWork() (UnitOfWork, error)
	RunInTx(ctx context.Context, fn func(ctx context.Context, uow UnitOfWork) error) error
}
