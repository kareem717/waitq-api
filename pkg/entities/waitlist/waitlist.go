package waitlist

import (
	"waitq/api/pkg/entities/shared"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Waitlist struct {
	bun.BaseModel `bun:"table:waitlists"`

	ID         uuid.UUID `json:"id"`
	AccountID  uuid.UUID `json:"accountId"`
	Name       string    `json:"name"`
	JWTSecret  string    `json:"jwtSecret"`
	AnonKey    string    `json:"anonKey"`
	ServiceKey string    `json:"serviceKey"`
	shared.Timestamps
}

type WaitlistAnalytics struct {
	WaitlistID         uuid.UUID `json:"waitlistId"`
	TotalEmails        int       `json:"totalEmails"`
	ActiveEmails       int       `json:"activeEmails"`
	UnsubscribedEmails int       `json:"unsubscribedEmails"`
	DeletedEmails      int       `json:"deletedEmails"`
}
