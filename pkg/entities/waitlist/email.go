package waitlist

import (
	"yakubu-llc/waitlist/pkg/entities/shared"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"time"
)

type Email struct {
	bun.BaseModel  `bun:"table:waitlist_emails"`
	
	ID             uuid.UUID  `json:"id"`
	WaitlistID     uuid.UUID  `json:"waitlistId"`
	Email          string     `json:"email"`
	UnsubscribedAt *time.Time `json:"unsubscribedAt"`
	shared.Timestamps
}
