package shared

import (
	"context"

	"waitq/api/pkg/entities/account"

	"github.com/supabase-community/gotrue-go/types"
)

const (
	UserContextKey    = "user"
	AccountContextKey = "account"

	WaitlistServiceKeyContextKey = "waitlistServiceKey"
	WaitlistAnonKeyContextKey    = "waitlistAnonKey"
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


func GetWaitlistAnonKey(ctx context.Context) string {
	ctxValue, ok := ctx.Value(WaitlistAnonKeyContextKey).(string)
	if !ok {
		return ""
	}

	return ctxValue
}

func GetWaitlistServiceKey(ctx context.Context) string {
	ctxValue, ok := ctx.Value(WaitlistServiceKeyContextKey).(string)
	if !ok {
		return ""
	}

	return ctxValue
}
