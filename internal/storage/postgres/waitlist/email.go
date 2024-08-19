package waitlist

import (
	"context"
	"waitq/api/internal/entities/waitlist"
	"waitq/api/internal/storage/postgres/shared"

	"github.com/google/uuid"
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
			ExcludeColumn("parsed_email").
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

func (r *WaitlistRepository) GetEmailsByWaitlistID(ctx context.Context, waitlistId uuid.UUID, input shared.EmailPaginationRequest) ([]waitlist.Email, error) {
	resp := []waitlist.Email{}

	query := r.db.NewSelect().
		Model(&resp).
		Where("waitlist_id = ?", waitlistId).
		OrderExpr("created_at DESC")

	if input.PageSize > 0 && input.Page > 0 {
		query = query.
			Limit(input.PageSize).
			Offset((input.Page - 1) * input.PageSize)
	}

	if !input.IncludeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	if !input.IncludeUnsubscribed {
		query = query.Where("unsubscribed_at IS NULL")
	}

	return resp, query.Scan(ctx)
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

func (r *WaitlistRepository) GetEmailCountByWaitlistID(ctx context.Context, waitlistId uuid.UUID, includeDeleted bool, includeUnsubscribed bool) (int, error) {
	query :=
		r.db.
			NewSelect().
			Model(&waitlist.Email{}).
			Where("waitlist_id = ?", waitlistId)

	if !includeDeleted {
		query = query.Where("deleted_at IS NULL")
	}

	if !includeUnsubscribed {
		query = query.Where("unsubscribed_at IS NULL")
	}

	return query.Count(ctx)
}


