package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/securecookie"
	"github.com/supabase-community/supabase-go"
	"go.uber.org/zap"
	"yakubu-llc/waitlist/pkg/entities/account"
	"yakubu-llc/waitlist/pkg/entities/waitlist"
	"yakubu-llc/waitlist/pkg/storage/postgres/shared"
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
	GetByAccountId(ctx context.Context, accountId uuid.UUID, input shared.PaginationRequest) ([]waitlist.Waitlist, error)
	GetById(ctx context.Context, id uuid.UUID) (waitlist.Waitlist, error)
	GetAnalytics(ctx context.Context, waitlistId uuid.UUID) (waitlist.WaitlistAnalytics, error)
	AddEmails(ctx context.Context, emails []waitlist.Email) ([]waitlist.Email, error)
	DeleteEmail(ctx context.Context, waitlistId uuid.UUID, email string) error
	UpdateEmail(ctx context.Context, input waitlist.Email) (waitlist.Email, error)
	GetEmailsByWaitlistID(ctx context.Context, waitlistId uuid.UUID, input shared.PaginationRequest) ([]waitlist.Email, error)
	UnsubscribeEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error)
	GetEmailsByWaitlistIDAndEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error)
}

// Service storage of all services.
type Service struct {
	AccountService  AccountService
	WaitlistService WaitlistService
	Logger          *zap.Logger
	CookieStore     *securecookie.SecureCookie
	SupabaseClient  *supabase.Client
}
