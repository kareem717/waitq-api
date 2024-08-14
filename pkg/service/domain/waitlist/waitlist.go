package waitlist

import (
	"context"

	"waitq/api/pkg/entities/waitlist"
	"waitq/api/pkg/storage"
	"waitq/api/pkg/storage/postgres/shared"

	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
)

type WaitlistService struct {
	waitlistRepository storage.WaitlistRepository // interface of the repository, not implementation
	sb                 *supabase.Client
}

func NewWaitlistService(waitlistRepository storage.WaitlistRepository, sb *supabase.Client) *WaitlistService {
	return &WaitlistService{
		waitlistRepository: waitlistRepository,
		sb:                 sb,
	}
}

func (s *WaitlistService) GetById(ctx context.Context, id uuid.UUID) (waitlist.Waitlist, error) {
	return s.waitlistRepository.GetById(ctx, id)
}

func (s *WaitlistService) GetByAccountId(ctx context.Context, accountId uuid.UUID, input shared.PaginationRequest) ([]waitlist.Waitlist, error) {
	return s.waitlistRepository.GetByAccountId(ctx, accountId, input)
}

func (s *WaitlistService) Create(ctx context.Context, input waitlist.Waitlist) (waitlist.Waitlist, error) {
	return s.waitlistRepository.Create(ctx, input)
}

func (s *WaitlistService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.waitlistRepository.Delete(ctx, id)
}

func (s *WaitlistService) Update(ctx context.Context, id uuid.UUID, input waitlist.Waitlist) (waitlist.Waitlist, error) {
	return s.waitlistRepository.Update(ctx, id, input)
}

func (s *WaitlistService) GetAnalytics(ctx context.Context, waitlistId uuid.UUID) (waitlist.WaitlistAnalytics, error) {
	return s.waitlistRepository.GetAnalytics(ctx, waitlistId)
}
