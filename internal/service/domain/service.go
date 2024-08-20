package domain

import (
	"waitq/api/internal/service"
	"waitq/api/internal/service/domain/account"
	"waitq/api/internal/service/domain/subscription"
	"waitq/api/internal/service/domain/waitlist"
	"waitq/api/internal/storage"
	"waitq/api/pkg/mailer"
	"waitq/api/pkg/stripe"

	supabase "github.com/supabase-community/supabase-go"
	"go.uber.org/zap"
)

// NewService implementation for storage of all services.
func NewService(
	repositories *storage.Repository,
	logger *zap.Logger,
	sb *supabase.Client,
	mailer *mailer.Mailer,
	stripeClient *stripe.Client,
) *service.Service {
	return &service.Service{
		AccountService:      account.NewAccountService(repositories.Account, sb),
		WaitlistService:     waitlist.NewWaitlistService(repositories.Waitlist, sb, mailer),
		SubscriptionService: subscription.NewSubscriptionService(repositories.Subscription, repositories.Waitlist, stripeClient, logger),
		Logger:              logger,
		SupabaseClient:      sb,
	}
}
