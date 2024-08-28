package subscription

import (
	"context"
	"waitq/api/internal/entities/subscription"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type SubscriptionRepository struct {
	db  bun.IDB
	ctx context.Context
}

func NewSubscriptionRepository(db bun.IDB, ctx context.Context) *SubscriptionRepository {
	return &SubscriptionRepository{
		db:  db,
		ctx: ctx,
	}
}

func (q *SubscriptionRepository) WithTransaction(tx bun.Tx) *SubscriptionRepository {
	return &SubscriptionRepository{db: tx, ctx: q.ctx}
}

func (q *SubscriptionRepository) BeginTx() (bun.Tx, error) {
	return q.db.BeginTx(q.ctx, nil)
}

func (q *SubscriptionRepository) RunInTx(fn func(ctx context.Context, tx bun.Tx) error) error {
	return q.db.RunInTx(q.ctx, nil, fn)
}

func (r *SubscriptionRepository) CreateAccountSubscription(ctx context.Context, sub subscription.AccountSubscription) (subscription.AccountSubscription, error) {
	resp := subscription.AccountSubscription{}

	err :=
		r.db.
			NewInsert().
			Model(&sub).
			ExcludeColumn("created_at", "updated_at", "deleted_at", "id").
			Set("subscription_id = ?", sub.SubscriptionID).
			Set("stripe_price_id = ?", sub.StripePriceID).
			Returning("*").
			Scan(ctx, &resp)

	return resp, err
}

func (r *SubscriptionRepository) UpdateAccountSubscription(ctx context.Context, accountId uuid.UUID, newSubscriptionId uuid.UUID, newPriceId string) (subscription.AccountSubscription, error) {
	resp := subscription.AccountSubscription{}

	err :=
		r.db.
			NewUpdate().
			Model(&resp).
			ExcludeColumn("created_at", "updated_at").
			Set("subscription_id = ?", newSubscriptionId).
			Set("stripe_price_id = ?", newPriceId).
			Where("account_id = ?", accountId).
			Where("deleted_at IS NULL").
			Returning("*").
			Scan(ctx, &resp)

	return resp, err
}
func (r *SubscriptionRepository) GetByAccountId(ctx context.Context, accountId uuid.UUID) (subscription.Subscription, error) {
	resp := subscription.Subscription{}

	err := r.db.
		NewSelect().
		Model(&resp).
		Join("JOIN account_subscriptions AS acc_sub").
		JoinOn("acc_sub.subscription_id = subscription.id").
		Where("acc_sub.account_id = ?", accountId).
		Where("acc_sub.deleted_at IS NULL").
		Scan(ctx, &resp)

	return resp, err
}

func (r *SubscriptionRepository) GetRelationshipByAccountId(ctx context.Context, accountId uuid.UUID) (subscription.AccountSubscription, error) {
	resp := subscription.AccountSubscription{}

	err := r.db.
		NewSelect().
		Model(&resp).
		Where("account_id = ?", accountId).
		Where("deleted_at IS NULL").
		Scan(ctx, &resp)

	return resp, err
}

func (r *SubscriptionRepository) GetByStripeProductId(ctx context.Context, stripeProductId string) (subscription.Subscription, error) {
	resp := subscription.Subscription{}

	err := r.db.
		NewSelect().
		Model(&resp).
		Where("stripe_product_id = ?", stripeProductId).
		Where("deleted_at IS NULL").
		Scan(ctx, &resp)

	return resp, err
}

func (r *SubscriptionRepository) DeleteRelationship(ctx context.Context, accountId uuid.UUID) error {
	_, err := r.db.
		NewUpdate().
		Model(&subscription.AccountSubscription{}).
		Set("deleted_at = CLOCK_TIMESTAMP()").
		Where("account_id = ?", accountId).
		Exec(ctx)

	return err
}

func (r *SubscriptionRepository) GetRelationshipByCustomerId(ctx context.Context, customerId string) (subscription.AccountSubscription, error) {
	resp := subscription.AccountSubscription{}

	err := r.db.
		NewSelect().
		Model(&resp).
		Where("stripe_customer_id = ?", customerId).
		Where("deleted_at IS NULL").
		Scan(ctx, &resp)

	return resp, err
}
