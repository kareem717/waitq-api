package billing

import (
	"context"
	"waitq/api/internal/entities/billing"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type BillingRepository struct {
	db  bun.IDB
	ctx context.Context
}

func NewBillingRepository(db bun.IDB, ctx context.Context) *BillingRepository {
	return &BillingRepository{
		db:  db,
		ctx: ctx,
	}
}

func (q *BillingRepository) WithTransaction(tx bun.Tx) *BillingRepository {
	return &BillingRepository{db: tx, ctx: q.ctx}
}

func (q *BillingRepository) BeginTx() (bun.Tx, error) {
	return q.db.BeginTx(q.ctx, nil)
}

func (q *BillingRepository) RunInTx(fn func(ctx context.Context, tx bun.Tx) error) error {
	return q.db.RunInTx(q.ctx, nil, fn)
}

func (r *BillingRepository) CreateAccountSubscription(ctx context.Context, sub billing.AccountSubscription) (billing.AccountSubscription, error) {
	resp := billing.AccountSubscription{}

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

func (r *BillingRepository) UpdateAccountSubscription(ctx context.Context, accountId uuid.UUID, newSubscriptionId uuid.UUID, newPriceId string) (billing.AccountSubscription, error) {
	resp := billing.AccountSubscription{}

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
func (r *BillingRepository) GetByAccountId(ctx context.Context, accountId uuid.UUID) (billing.Subscription, error) {
	resp := billing.Subscription{}

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

func (r *BillingRepository) GetRelationshipByAccountId(ctx context.Context, accountId uuid.UUID) (billing.AccountSubscription, error) {
	resp := billing.AccountSubscription{}

	err := r.db.
		NewSelect().
		Model(&resp).
		Where("account_id = ?", accountId).
		Where("deleted_at IS NULL").
		Scan(ctx, &resp)

	return resp, err
}

func (r *BillingRepository) GetByStripeProductId(ctx context.Context, stripeProductId string) (billing.Subscription, error) {
	resp := billing.Subscription{}

	err := r.db.
		NewSelect().
		Model(&resp).
		Where("stripe_product_id = ?", stripeProductId).
		Where("deleted_at IS NULL").
		Scan(ctx, &resp)

	return resp, err
}

func (r *BillingRepository) DeleteRelationship(ctx context.Context, accountId uuid.UUID) error {
	_, err := r.db.
		NewUpdate().
		Model(&billing.AccountSubscription{}).
		Set("deleted_at = CLOCK_TIMESTAMP()").
		Where("account_id = ?", accountId).
		Exec(ctx)

	return err
}

func (r *BillingRepository) GetRelationshipByCustomerId(ctx context.Context, customerId string) (billing.AccountSubscription, error) {
	resp := billing.AccountSubscription{}

	err := r.db.
		NewSelect().
		Model(&resp).
		Where("stripe_customer_id = ?", customerId).
		Where("deleted_at IS NULL").
		Scan(ctx, &resp)

	return resp, err
}
