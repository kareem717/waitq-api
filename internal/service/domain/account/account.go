package account

import (
	"context"

	"waitq/api/internal/entities/account"
	"waitq/api/internal/storage"
	"waitq/api/internal/storage/postgres/shared"

	"github.com/google/uuid"
)

type AccountService struct {
	repositories storage.Repository
}

// NewAccountService returns a new instance of account service.
func NewAccountService(repositories storage.Repository) *AccountService {
	return &AccountService{
		repositories: repositories,
	}
}

func (s *AccountService) GetById(ctx context.Context, id uuid.UUID) (account.Account, error) {
	return s.repositories.Account().GetById(ctx, id)
}

func (s *AccountService) GetByUserId(ctx context.Context, userId uuid.UUID, input shared.GetManyRequest) ([]account.Account, error) {
	return s.repositories.Account().GetByUserId(ctx, userId, input)
}

func (s *AccountService) Create(ctx context.Context, input account.Account) (account.Account, error) {
	return s.repositories.Account().Create(ctx, input)
}

func (s *AccountService) Delete(ctx context.Context, accountId uuid.UUID) error {
	return s.repositories.Account().Delete(ctx, accountId)
}

func (s *AccountService) Update(ctx context.Context, id uuid.UUID, input account.Account) (account.Account, error) {
	return s.repositories.Account().Update(ctx, id, input)
}
