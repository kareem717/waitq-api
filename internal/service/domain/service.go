package domain

import (
	"waitq/api/internal/service"
	"waitq/api/internal/service/domain/account"
	"waitq/api/internal/service/domain/waitlist"
	"waitq/api/internal/storage"
	"waitq/api/pkg/mailer"

	supabase "github.com/supabase-community/supabase-go"
	"go.uber.org/zap"
)

// NewService implementation for storage of all services.
func NewService(
	repositories *storage.Repository,
	logger *zap.Logger,
	sb *supabase.Client,
	mailer *mailer.Mailer,
) *service.Service {
	return &service.Service{
		AccountService:  account.NewAccountService(repositories.Account, sb),
		WaitlistService: waitlist.NewWaitlistService(repositories.Waitlist, sb, mailer),
		Logger:          logger,
		SupabaseClient:  sb,
	}
}
