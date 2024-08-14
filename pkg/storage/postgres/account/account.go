package account

import (
	"context"
	"waitq/api/pkg/entities/account"
	"waitq/api/pkg/storage/postgres/shared"

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

func (r *AccountRepository) Create(ctx context.Context, input account.Account) (account.Account, error) {
	resp := account.Account{}

	err := shared.ExcludeInsertColumns(
		r.db.
			NewInsert().
			Model(&input).
			Returning("*"),
	).Scan(ctx, &resp)

	return resp, err
}

func (r *AccountRepository) Update(ctx context.Context, id uuid.UUID, input account.Account) (account.Account, error) {
	resp := account.Account{}

	err :=
		shared.ExcludeUpdateColumns(
			r.db.
				NewUpdate().
				Model(&input).
				OmitZero().
				Where("id = ?", id).
				Returning("*"),
		).
			Scan(ctx, &resp)

	return resp, err
}

func (r *AccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err :=
		r.db.
			NewUpdate().
			Model(&account.Account{}).
			Set("deleted_at = clock_timestamp()").
			Where("id = ?", id).
			Exec(ctx)

	return err
}

func (r *AccountRepository) GetById(ctx context.Context, id uuid.UUID) (account.Account, error) {
	resp := account.Account{}

	err := r.db.NewSelect().Model(&resp).Where("id = ?", id).Scan(ctx)

	return resp, err
}

func (r *AccountRepository) GetByUserId(ctx context.Context, userId uuid.UUID, input shared.GetManyRequest) ([]account.Account, error) {
	resp := []account.Account{}

	err := r.db.
		NewSelect().
		Model(&resp).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			query := q.Where("user_id = ?", userId)
			if !input.IncludeDeleted {
				query = query.Where("deleted_at IS NULL")
			}
			return query
		}).
		Scan(ctx)

	return resp, err
}
