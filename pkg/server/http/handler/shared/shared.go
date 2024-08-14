package shared

import (
	"context"

	"github.com/supabase-community/gotrue-go/types"
	"yakubu-llc/waitlist/pkg/entities/account"
)

const (
	UserContextKey    = "user"
	AccountContextKey = "account"
)

func GetAuthenticatedUser(ctx context.Context) types.User {
	ctxValue, ok := ctx.Value(UserContextKey).(types.User)
	if !ok {
		return types.User{}
	}

	return ctxValue
}

func GetAuthenticatedAccount(ctx context.Context) account.Account {
	ctxValue, ok := ctx.Value(AccountContextKey).(account.Account)
	if !ok {
		return account.Account{}
	}

	return ctxValue
}
