package waitlist

import (
	"context"
	"yakubu-llc/waitlist/pkg/entities/waitlist"
	"yakubu-llc/waitlist/pkg/storage/postgres/shared"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type WaitlistRepository struct {
	db  bun.IDB
	ctx context.Context
}

// NewAccountRepository returns a new instance of the repository.
func NewWaitlistRepository(db bun.IDB, ctx context.Context) *WaitlistRepository {
	return &WaitlistRepository{
		db:  db,
		ctx: ctx,
	}
}

func (r *WaitlistRepository) Create(ctx context.Context, input waitlist.Waitlist) (waitlist.Waitlist, error) {
	resp := waitlist.Waitlist{}

	err := shared.ExcludeInsertColumns(
		r.db.
			NewInsert().
			Model(&input).
			Returning("*"),
	).Scan(ctx, &resp)

	return resp, err
}

func (r *WaitlistRepository) Update(ctx context.Context, id uuid.UUID, input waitlist.Waitlist) (waitlist.Waitlist, error) {
	resp := waitlist.Waitlist{}

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

func (r *WaitlistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err :=
		r.db.
			NewUpdate().
			Model(&waitlist.Waitlist{}).
			Set("deleted_at = clock_timestamp()").
			Where("id = ?", id).
			Exec(ctx)

	return err
}

func (r *WaitlistRepository) GetById(ctx context.Context, id uuid.UUID) (waitlist.Waitlist, error) {
	resp := waitlist.Waitlist{}

	err := r.db.NewSelect().Model(&resp).Where("id = ?", id).Scan(ctx)

	return resp, err
}

func (r *WaitlistRepository) GetByAccountId(ctx context.Context, accountId uuid.UUID, input shared.PaginationRequest) ([]waitlist.Waitlist, error) {
	resp := []waitlist.Waitlist{}

	err := r.db.
		NewSelect().
		Model(&resp).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			query := q.Where("account_id = ?", accountId)
			if !input.IncludeDeleted {
				query = query.Where("deleted_at IS NULL")
			}
			return query
		}).
		OrderExpr("created_at DESC").
		Limit(input.PageSize).
		Offset((input.Page - 1) * input.PageSize).
		Scan(ctx)

	return resp, err
}

func (r *WaitlistRepository) GetAnalytics(ctx context.Context, waitlistId uuid.UUID) (waitlist.WaitlistAnalytics, error) {
	resp := waitlist.WaitlistAnalytics{}

	err := r.db.
		NewSelect().
		Model(&resp).
		ModelTableExpr("waitlist_emails").
		ColumnExpr("waitlist_id").
		ColumnExpr("COUNT(waitlist_id) AS total_emails").
		ColumnExpr("COUNT(CASE WHEN deleted_at IS NULL AND unsubscribed_at IS NULL THEN 1 END) AS active_emails").
		ColumnExpr("COUNT(CASE WHEN deleted_at IS NOT NULL THEN 1 END) AS deleted_emails").
		ColumnExpr("COUNT(CASE WHEN unsubscribed_at IS NOT NULL THEN 1 END) AS unsubscribed_emails").
		Where("waitlist_id = ?", waitlistId).
		GroupExpr("waitlist_id").
		Scan(ctx, &resp)

	return resp, err
}
