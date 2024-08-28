package waitlist

import (
	"context"
	"errors"
	"math"
	"waitq/api/internal/entities/waitlist"
	"waitq/api/internal/storage/postgres/shared"

	"github.com/google/uuid"
)

func (s *WaitlistService) AddEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error) {
	emails, err := s.repositories.Waitlist().AddEmails(ctx, []waitlist.Email{{WaitlistID: waitlistId, Email: email}})
	if err != nil {
		return waitlist.Email{}, err
	}

	return emails[0], nil
}

func (s *WaitlistService) GetEmailsByWaitlistID(ctx context.Context, waitlistId uuid.UUID, input shared.EmailPaginationRequest) ([]waitlist.Email, error) {
	return s.repositories.Waitlist().GetEmailsByWaitlistID(ctx, waitlistId, input)
}

func (s *WaitlistService) DeleteEmail(ctx context.Context, waitlistId uuid.UUID, email string) error {
	return s.repositories.Waitlist().DeleteEmail(ctx, waitlistId, email)
}

func (s *WaitlistService) UpdateEmail(ctx context.Context, email waitlist.Email) (waitlist.Email, error) {
	return s.repositories.Waitlist().UpdateEmail(ctx, email)
}

func (s *WaitlistService) UnsubscribeEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error) {
	return s.repositories.Waitlist().UnsubscribeEmail(ctx, waitlistId, email)
}

func (s *WaitlistService) GetEmailsByWaitlistIDAndEmail(ctx context.Context, waitlistId uuid.UUID, email string) (waitlist.Email, error) {
	return s.repositories.Waitlist().GetEmailsByWaitlistIDAndEmail(ctx, waitlistId, email)
}

func (s *WaitlistService) ExportEmails(ctx context.Context, waitlistId uuid.UUID) (chan string, chan error) {
	emailsChan := make(chan string, 100) // Buffered channel with size 1
	errChan := make(chan error, 1)       // Buffered channel with size 1

	emailCount, err := s.repositories.Waitlist().GetEmailCountByWaitlistID(ctx, waitlistId, false, false)
	if err != nil {
		errChan <- err
		close(errChan)
		close(emailsChan)
		return emailsChan, errChan
	}

	if emailCount == 0 {
		errChan <- errors.New("no emails found")
		close(emailsChan)
		close(errChan)
		return emailsChan, errChan
	}

	totalPages := math.Ceil(float64(emailCount) / float64(100))

	go func() {
		defer close(emailsChan)
		defer close(errChan)
		for i := 0; i < int(totalPages); i++ {
			emails, err := s.GetEmailsByWaitlistID(ctx, waitlistId, shared.EmailPaginationRequest{Page: i, PageSize: 100})
			if err != nil {
				errChan <- err
				return
			}

			for _, email := range emails {
				emailsChan <- email.Email
			}

		}
	}()

	return emailsChan, errChan
}

func (s *WaitlistService) GetActiveEmailCountByWaitlistId(ctx context.Context, waitlistId uuid.UUID) (int, error) {
	return s.repositories.Waitlist().GetEmailCountByWaitlistID(ctx, waitlistId, false, false)
}
