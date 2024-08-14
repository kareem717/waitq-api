package account

import (
	"context"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
	"yakubu-llc/waitlist/pkg/entities/account"
	"yakubu-llc/waitlist/pkg/storage"
	"yakubu-llc/waitlist/pkg/storage/postgres/shared"
)

type AccountService struct {
	accountRepository storage.AccountRepository // interface of the repository, not implementation
	sb                *supabase.Client
}

// NewAccountService returns a new instance of account service.
func NewAccountService(accountRepository storage.AccountRepository, sb *supabase.Client) *AccountService {
	return &AccountService{
		accountRepository: accountRepository,
		sb:                sb,
	}
}

func (s *AccountService) GetById(ctx context.Context, id uuid.UUID) (account.Account, error) {
	return s.accountRepository.GetById(ctx, id)
}

func (s *AccountService) GetByUserId(ctx context.Context, userId uuid.UUID, input shared.GetManyRequest) ([]account.Account, error) {
	return s.accountRepository.GetByUserId(ctx, userId, input)
}

func (s *AccountService) Create(ctx context.Context, input account.Account) (account.Account, error) {
	return s.accountRepository.Create(ctx, input)
}

func (s *AccountService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.accountRepository.Delete(ctx, id)
}

func (s *AccountService) Update(ctx context.Context, id uuid.UUID, input account.Account) (account.Account, error) {
	return s.accountRepository.Update(ctx, id, input)
}
