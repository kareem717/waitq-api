package token

import (
	"context"
	"time"
	"waitq/api/internal/entities/token"

	"github.com/uptrace/bun"
)

type TokenRepository struct {
	db  bun.IDB
	ctx context.Context
}

// NewAccountRepository returns a new instance of the repository.
func NewTokenRepository(db bun.IDB, ctx context.Context) *TokenRepository {
	return &TokenRepository{
		db:  db,
		ctx: ctx,
	}
}

func (r *TokenRepository) CreateRandom(ctx context.Context) (token.Token, error) {
	resp := token.Token{}

	resp.ExpiresAt = time.Now().Add(time.Minute * 15)

	err :=
		r.db.
			NewInsert().
			Model(&resp).
			ExcludeColumn("id").
			ExcludeColumn("created_at").
			Returning("*").
			Scan(ctx, &resp)

	return resp, err
}

func (r *TokenRepository) GetByToken(ctx context.Context, input int) (token.Token, error) {
	resp := token.Token{}

	err := r.db.
		NewSelect().
		Model(&resp).
		Where("token = ?", input).
		Scan(ctx)

	return resp, err
}
