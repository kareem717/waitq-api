package account

import (
	"context"

	"waitq/api/internal/entities/account"
	"waitq/api/internal/storage"
	"waitq/api/internal/storage/postgres/shared"
	"waitq/api/pkg/stripe"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
)

type AccountService struct {
	accountRepository storage.AccountRepository // interface of the repository, not implementation
	sb                *supabase.Client
	stripeClient      *stripe.Client
}

// NewAccountService returns a new instance of account service.
func NewAccountService(accountRepository storage.AccountRepository, sb *supabase.Client, stripeClient *stripe.Client) *AccountService {
	return &AccountService{
		accountRepository: accountRepository,
		sb:                sb,
		stripeClient:      stripeClient,
	}
}

func (s *AccountService) GetById(ctx context.Context, id uuid.UUID) (account.Account, error) {
	return s.accountRepository.GetById(ctx, id)
}

func (s *AccountService) GetByUserId(ctx context.Context, userId uuid.UUID, input shared.GetManyRequest) ([]account.Account, error) {
	return s.accountRepository.GetByUserId(ctx, userId, input)
}

func (s *AccountService) Create(ctx context.Context, input account.Account) (account.Account, error) {
	customer, err := s.stripeClient.CreateCustomer(input.Email, input.Name)
	if err != nil {
		return account.Account{}, err
	}

	input.StripeCustomerID = customer.ID

	return s.accountRepository.Create(ctx, input)
}

func (s *AccountService) Delete(ctx context.Context, accountId uuid.UUID) error {
	account, err := s.accountRepository.GetById(ctx, accountId)
	if err != nil {
		return err
	}

	err = s.stripeClient.DeleteCustomer(account.StripeCustomerID)
	if err != nil {
		return err
	}

	return s.accountRepository.Delete(ctx, accountId)
}

func (s *AccountService) Update(ctx context.Context, id uuid.UUID, input account.Account) (account.Account, error) {
	return s.accountRepository.Update(ctx, id, input)
}
