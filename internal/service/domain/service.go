package domain

import (
	"waitq/api/internal/service"
	"waitq/api/internal/service/domain/account"
	"waitq/api/internal/service/domain/billing"
	"waitq/api/internal/service/domain/waitlist"
	"waitq/api/internal/storage"
	"waitq/api/pkg/stripe"
)

// NewService implementation for storage of all services.
func NewService(
	repositories storage.Repository,
	stripeClient *stripe.Client,
) *service.Service {
	return &service.Service{
		AccountService:  account.NewAccountService(repositories),
		WaitlistService: waitlist.NewWaitlistService(repositories),
		BillingService:  billing.NewBillingService(repositories, stripeClient),
	}
}
