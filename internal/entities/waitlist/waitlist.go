package waitlist

import (
	"waitq/api/internal/entities/shared"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Waitlist struct {
	bun.BaseModel `bun:"table:waitlists"`

	ID         uuid.UUID `json:"id"`
	AccountID  uuid.UUID `json:"accountId"`
	Name       string    `json:"name"`
	URLAlias   string    `json:"urlAlias"`
	JWTSecret  string    `json:"jwtSecret"`
	AnonKey    string    `json:"anonKey"`
	ServiceKey string    `json:"serviceKey"`
	shared.Timestamps
}

func (w *Waitlist) PublicWaitlist() PublicWaitlist {
	return PublicWaitlist{
		ID:         w.ID,
		URLAlias:   w.URLAlias,
		Name:       w.Name,
		AccountID:  w.AccountID,
		Timestamps: w.Timestamps,
	}
}

type PublicWaitlist struct {
	ID        uuid.UUID `json:"id"`
	URLAlias  string    `json:"urlAlias"`
	Name      string    `json:"name"`
	AccountID uuid.UUID `json:"accountId"`
	shared.Timestamps
}

type WaitlistAnalytics struct {
	WaitlistID         uuid.UUID `json:"waitlistId"`
	TotalEmails        int       `json:"totalEmails"`
	ActiveEmails       int       `json:"activeEmails"`
	UnsubscribedEmails int       `json:"unsubscribedEmails"`
	DeletedEmails      int       `json:"deletedEmails"`
}
