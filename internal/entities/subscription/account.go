package subscription

import (
	"waitq/api/internal/entities/shared"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AccountSubscription struct {
	bun.BaseModel `bun:"table:account_subscriptions"`

	ID                   uuid.UUID `json:"id"`
	AccountID            uuid.UUID `json:"accountId"`
	SubscriptionID       uuid.UUID `json:"subscriptionId"`
	StripeSubscriptionID string    `json:"stripeSubscriptionID"`
	StripePriceID        string    `json:"stripePriceID"`
	StripeCustomerID     string    `json:"stripeCustomerID"`
	shared.Timestamps
}
