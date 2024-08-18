package subscription

import (
	"waitq/api/internal/entities/shared"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AccountSubscription struct {
	bun.BaseModel `bun:"table:account_subscriptions"`

	AccountID            uuid.UUID `json:"accountId"`
	SubscriptionID       uuid.UUID `json:"subscriptionId"`
	StripeSubscriptionID string    `json:"stripeSubscriptionID"`
	StripeCustomerID     string    `json:"stripeCustomerID"`
	shared.Timestamps
}
