package account

import (
	"context"
	"waitq/api/internal/entities/account"
	"waitq/api/internal/storage/postgres/shared"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AccountRepository struct {
	db  bun.IDB
	ctx context.Context
}

// NewAccountRepository returns a new instance of the repository.
func NewAccountRepository(db bun.IDB, ctx context.Context) *AccountRepository {
	return &AccountRepository{
		db:  db,
		ctx: ctx,
	}
}

func (r *AccountRepository) Create(ctx context.Context, accountParams account.Account) (account.Account, error) {
	resp := account.Account{}

	err := shared.ExcludeInsertColumns(
		r.db.
			NewInsert().
			Model(&accountParams).
			ExcludeColumn("id").
			ExcludeColumn("parsed_email").
			Returning("*"),
	).Scan(ctx, &resp)

	return resp, err
}

func (r *AccountRepository) Update(ctx context.Context, accountParams account.Account) (account.Account, error) {
	resp := account.Account{}

	err :=
		shared.ExcludeUpdateColumns(
			r.db.
				NewUpdate().
				Model(&accountParams).
				ExcludeColumn("parsed_email").
				OmitZero().
				Where("id = ?", accountParams.ID).
				Returning("*"),
		).
			Scan(ctx, &resp)

	return resp, err
}

func (r *AccountRepository) Delete(ctx context.Context, accountId uuid.UUID) error {
	_, err :=
		r.db.
			NewUpdate().
			Model(&account.Account{}).
			Set("deleted_at = clock_timestamp()").
			Where("id = ?", accountId).
			Exec(ctx)

	return err
}

func (r *AccountRepository) GetById(ctx context.Context, accountId uuid.UUID) (account.Account, error) {
	resp := account.Account{}

	err := r.db.
		NewSelect().
		Model(&resp).
		Where("id = ?", accountId).
		Where("deleted_at IS NULL").
		Scan(ctx)

	return resp, err
}

func (r *AccountRepository) GetByUserId(ctx context.Context, userId uuid.UUID, requestParams shared.GetManyRequest) (account.Account, error) {
	resp := account.Account{}

	err := r.db.
		NewSelect().
		Model(&resp).
		Where("user_id = ?", userId).
		Where("deleted_at IS NULL").
		Scan(ctx)

	return resp, err
}
