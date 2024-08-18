package token

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Token struct {
	bun.BaseModel `bun:"table:verification_tokens"`

	ID        uuid.UUID `json:"id"`
	Token     int       `json:"token"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}
