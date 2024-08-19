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

func (r *SubscriptionRepository) UpdateAccountSubscription(ctx context.Context, sub subscription.AccountSubscription) (subscription.AccountSubscription, error) {
	resp := subscription.AccountSubscription{}

	err :=
		r.db.
			NewInsert().
			Model(&sub).
			ExcludeColumn("created_at", "updated_at", "deleted_at").
			On("CONFLICT (account_id) DO UPDATE").
			Set("subscription_id = ?", sub.SubscriptionID).
			Set("stripe_customer_id = ?", sub.StripeCustomerID).
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
		Scan(ctx, &resp)

	return resp, err
}

func (r *SubscriptionRepository) GetByStripeProductId(ctx context.Context, stripeProductId string) (subscription.Subscription, error) {
	resp := subscription.Subscription{}

	err := r.db.
		NewSelect().
		Model(&resp).
		Where("stripe_product_id = ?", stripeProductId).
		Scan(ctx, &resp)

	return resp, err
}
