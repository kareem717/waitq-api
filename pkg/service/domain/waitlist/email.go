package waitlist

import (
	"context"

	"yakubu-llc/waitlist/pkg/entities/waitlist"
	"yakubu-llc/waitlist/pkg/storage/postgres/shared"

	"github.com/google/uuid"
)

func (s *WaitlistService) AddEmails(ctx context.Context, emails []waitlist.Email) ([]waitlist.Email, error) {
	
	
	return s.waitlistRepository.AddEmails(ctx, emails)
}

func (s *WaitlistService) GetEmailsByWaitlistID(ctx context.Context, waitlistId uuid.UUID, input shared.PaginationRequest) ([]waitlist.Email, error) {
	return s.waitlistRepository.GetEmailsByWaitlistID(ctx, waitlistId, input)
}

func (s *WaitlistService) DeleteEmail(ctx context.Context, waitlistId uuid.UUID, email string) error {
	return s.waitlistRepository.DeleteEmail(ctx, waitlistId, email)
}

func (s *WaitlistService) UpdateEmail(ctx context.Context, email waitlist.Email) (waitlist.Email, error) {
	return s.waitlistRepository.UpdateEmail(ctx, email)
}

func (s *WaitlistService) UnsubscribeEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error) {
	return s.waitlistRepository.UnsubscribeEmail(ctx, waitlistId, email)
}

func (s *WaitlistService) GetEmailsByWaitlistIDAndEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error) {
	return s.waitlistRepository.GetEmailsByWaitlistIDAndEmail(ctx, waitlistId, email)
}
