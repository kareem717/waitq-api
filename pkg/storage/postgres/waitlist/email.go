package waitlist

import (
	"context"
	"yakubu-llc/waitlist/pkg/entities/waitlist"
	"yakubu-llc/waitlist/pkg/storage/postgres/shared"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

func (r *WaitlistRepository) AddEmails(ctx context.Context, emails []waitlist.Email) ([]waitlist.Email, error) {
	resp := []waitlist.Email{}

	err :=
		r.db.
			NewInsert().
			Model(&emails).
			ExcludeColumn("id").
			ExcludeColumn("created_at").
			ExcludeColumn("updated_at").
			ExcludeColumn("deleted_at").
			ExcludeColumn("unsubscribed_at").
			Returning("*").
			Scan(ctx, &resp)

	return resp, err
}

func (r *WaitlistRepository) UpdateEmail(ctx context.Context, input waitlist.Email) (waitlist.Email, error) {
	resp := waitlist.Email{}

	err :=
		shared.ExcludeUpdateColumns(
			r.db.
				NewUpdate().
				Model(&input).
				OmitZero().
				Where("waitlist_id = ?", input.WaitlistID).
				Where("email = ?", input.Email).
				Returning("*"),
		).
			Scan(ctx, &resp)

	return resp, err
}

func (r *WaitlistRepository) DeleteEmail(ctx context.Context, waitlistId uuid.UUID, email string) error {
	_, err :=
		r.db.
			NewUpdate().
			Model(&waitlist.Email{}).
			Set("deleted_at = clock_timestamp()").
			Where("waitlist_id = ?", waitlistId).
			Where("email = ?", email).
			Exec(ctx)

	return err
}

func (r *WaitlistRepository) GetEmailsByWaitlistID(ctx context.Context, waitlistId uuid.UUID, input shared.PaginationRequest) ([]waitlist.Email, error) {
	resp := []waitlist.Email{}

	err := r.db.NewSelect().
		Model(&resp).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			query := q.Where("waitlist_id = ?", waitlistId)
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

func (r *WaitlistRepository) GetEmailsByWaitlistIDAndEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error) {
	resp := waitlist.Email{}

	err := r.db.NewSelect().
		Model(&resp).
		Where("waitlist_id = ?", waitlistId).
		Where("email = ?", email).
		Scan(ctx)

	return resp, err
}

func (r *WaitlistRepository) UnsubscribeEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error) {
	resp := waitlist.Email{}

	err :=
		r.db.
			NewUpdate().
			Model(&waitlist.Email{}).
			Set("unsubscribed_at = clock_timestamp()").
			Where("waitlist_id = ?", waitlistId).
			Where("email = ?", email).
			Returning("*").
			Scan(ctx, &resp)

	return resp, err
}
