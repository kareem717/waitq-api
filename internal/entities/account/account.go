package account

import (
	"waitq/api/internal/entities/shared"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Account struct {
	bun.BaseModel `bun:"table:accounts"`

	ID               uuid.UUID          `json:"id"`
	UserID           uuid.UUID          `json:"userId"`
	Name             string             `json:"name"`
	Email            string             `json:"email"`
	ParsedEmail      shared.ParsedEmail `json:"parsedEmail"`
	StripeCustomerID string             `json:"stripeCustomerID"`
	shared.Timestamps
}
