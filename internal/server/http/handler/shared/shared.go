package shared

import (
	"context"

	"waitq/api/internal/entities/account"
	"waitq/api/internal/entities/billing"

	"github.com/google/uuid"
	"github.com/supabase-community/gotrue-go/types"
)

const (
	UserContextKey                 = "user"
	AccountContextKey              = "account"
	WaitlistKeyContextKey          = "waitlistKey"
	WaitlistSubscriptionContextKey = "waitlistSubscription"
)

func GetAuthenticatedUser(ctx context.Context) types.User {
	if ctxValue, ok := ctx.Value(UserContextKey).(types.User); ok {
		return ctxValue
	}

	return types.User{}
}

func GetAuthenticatedAccount(ctx context.Context) account.Account {
	if ctxValue, ok := ctx.Value(AccountContextKey).(account.Account); ok {
		return ctxValue
	}

	return account.Account{}
}

type KeyRole string

const (
	WaitlistServiceKey KeyRole = "service"
	WaitlistAnonKey    KeyRole = "anon"
)

type WaitlistKey struct {
	ID   uuid.UUID
	Role KeyRole
}

func GetWaitlistKey(ctx context.Context) WaitlistKey {
	if ctxValue, ok := ctx.Value(WaitlistKeyContextKey).(WaitlistKey); ok {
		return ctxValue
	}

	return WaitlistKey{}
}

func GetWaitlistSubscription(ctx context.Context) *billing.Subscription {
	if ctxValue, ok := ctx.Value(WaitlistSubscriptionContextKey).(*billing.Subscription); ok {
		return ctxValue
	}

	return nil
}
