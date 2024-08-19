package subscription

import (
	"waitq/api/internal/entities/shared"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Subscription struct {
	bun.BaseModel `bun:"table:subscriptions"`

	ID                   uuid.UUID `json:"id"`
	Name                 string    `json:"name"`
	StripeProductID      string    `json:"stripeProductID"`
	PermissionTier       int       `json:"permissionTier"`
	MaxPeoplePerWaitlist int       `json:"maxPeoplePerWaitlist"`
	MaxWaitlists         int       `json:"maxWaitlists"`
	shared.Timestamps
}
