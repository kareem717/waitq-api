package domain

import (
	"waitq/api/pkg/service"
	"waitq/api/pkg/service/domain/account"
	"waitq/api/pkg/service/domain/waitlist"
	"waitq/api/pkg/storage"

	"github.com/gorilla/securecookie"
	supabase "github.com/supabase-community/supabase-go"
	"go.uber.org/zap"
)

// NewService implementation for storage of all services.
func NewService(
	repositories *storage.Repository,
	logger *zap.Logger,
	sb *supabase.Client,
	cookieStore *securecookie.SecureCookie,
) *service.Service {
	return &service.Service{
		AccountService:  account.NewAccountService(repositories.Account, sb),
		WaitlistService: waitlist.NewWaitlistService(repositories.Waitlist, sb),
		CookieStore:     cookieStore,
		Logger:          logger,
		SupabaseClient:  sb,
	}
}
