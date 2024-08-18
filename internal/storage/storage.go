package storage

import (
	"context"
	"go/token"
	"waitq/api/internal/entities/account"
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
}

type TokenRepository interface {
	Create(ctx context.Context, input token.Token) (token.Token, error)
	GetByToken(ctx context.Context, token int) (token.Token, error)
}

type Repository struct {
	Account  AccountRepository
	Waitlist WaitlistRepository
}
