package domain

import (
	"github.com/gorilla/securecookie"
	supabase "github.com/supabase-community/supabase-go"
	"go.uber.org/zap"
	"yakubu-llc/waitlist/pkg/service"
	"yakubu-llc/waitlist/pkg/service/domain/account"
	"yakubu-llc/waitlist/pkg/service/domain/waitlist"
	"yakubu-llc/waitlist/pkg/storage"
)

// NewService implementation for storage of all services.
func NewService(
	repositories *storage.Repository,
	logger *zap.Logger,
	sb *supabase.Client,
	cookieStore *securecookie.SecureCookie,
) *service.Service {
	return &service.Service{
		AccountService: account.NewAccountService(repositories.Account, sb),
		WaitlistService: waitlist.NewWaitlistService(repositories.Waitlist, sb),
		CookieStore:    cookieStore,
		Logger:         logger,
		SupabaseClient: sb,
	}
}
