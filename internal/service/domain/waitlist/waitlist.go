package waitlist

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"waitq/api/internal/entities/waitlist"
	utils "waitq/api/internal/service/domain/shared"
	"waitq/api/internal/storage"
	pgShared "waitq/api/internal/storage/postgres/shared"

	"github.com/golang-jwt/jwt/v5"

	"github.com/google/uuid"
)

type WaitlistService struct {
	repositories storage.Repository
}

func NewWaitlistService(repositories storage.Repository) *WaitlistService {
	return &WaitlistService{
		repositories: repositories,
	}
}

func (s *WaitlistService) GetById(ctx context.Context, id uuid.UUID) (waitlist.Waitlist, error) {
	return s.repositories.Waitlist().GetById(ctx, id)
}

func (s *WaitlistService) GetByAccountId(ctx context.Context, accountId uuid.UUID, input pgShared.PaginationRequest) ([]waitlist.Waitlist, error) {
	return s.repositories.Waitlist().GetByAccountId(ctx, accountId, input)
}

func (s *WaitlistService) Create(ctx context.Context, input waitlist.Waitlist) (waitlist.Waitlist, error) {
	var err error

	input.JWTSecret, err = utils.GenerateRandomString(64)
	if err != nil {
		return waitlist.Waitlist{}, err
	}

	input.ID = uuid.New()
	keys, err := generateKeys(input.JWTSecret, input.ID)
	if err != nil {
		return waitlist.Waitlist{}, err
	}

	input.AnonKey = keys.AnonKey
	input.ServiceKey = keys.ServiceKey

	return s.repositories.Waitlist().Create(ctx, input)
}

func (s *WaitlistService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repositories.Waitlist().Delete(ctx, id)
}

func (s *WaitlistService) Update(ctx context.Context, id uuid.UUID, input waitlist.Waitlist) (waitlist.Waitlist, error) {
	curr, err := s.repositories.Waitlist().GetById(ctx, id)
	if err != nil {
		return waitlist.Waitlist{}, err
	}

	// Disable jwt secret update
	if input.JWTSecret != "" {
		input.JWTSecret = curr.JWTSecret
		input.AnonKey = curr.AnonKey
		input.ServiceKey = curr.ServiceKey
	}

	// Update already omits zero values, thus if input.JWTSecret is empty, it will not be updated
	return s.repositories.Waitlist().Update(ctx, id, input)
}

func (s *WaitlistService) UpdateJWTSecret(ctx context.Context, id uuid.UUID, secret string) (waitlist.Waitlist, error) {
	var vals waitlist.Waitlist

	log.Printf("vals_1: %+v", vals)

	// Disable jwt secret update
	if secret == "" {
		generatedSecret, err := utils.GenerateRandomString(64)
		if err != nil {
			return waitlist.Waitlist{}, err
		}

		vals.JWTSecret = generatedSecret
	} else {
		vals.JWTSecret = secret
	}
	log.Printf("vals_2: %+v", vals)

	// Generate new api keys
	keys, err := generateKeys(vals.JWTSecret, id)
	if err != nil {
		return waitlist.Waitlist{}, err
	}

	vals.AnonKey = keys.AnonKey
	vals.ServiceKey = keys.ServiceKey

	log.Printf("vals_3: %+v", vals)

	// Update already omits zero values, thus if input.JWTSecret is empty, it will not be updated
	return s.repositories.Waitlist().Update(ctx, id, vals)
}

func (s *WaitlistService) GetActiveWaitlistCountByAccountId(ctx context.Context, accountId uuid.UUID) (int, error) {
	return s.repositories.Waitlist().GetActiveWaitlistCountByAccountId(ctx, accountId)
}

func (s *WaitlistService) GetAnalytics(ctx context.Context, waitlistId uuid.UUID) (waitlist.WaitlistAnalytics, error) {
	analytics, err := s.repositories.Waitlist().GetAnalytics(ctx, waitlistId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return waitlist.WaitlistAnalytics{}, nil
		}
		
		return waitlist.WaitlistAnalytics{}, err
	}

	return analytics, nil
}

type Keys struct {
	AnonKey    string
	ServiceKey string
}

type KeyRole string

const (
	AnonKey    KeyRole = "anon"
	ServiceKey KeyRole = "service"
)

func generateToken(secret string, id uuid.UUID, role KeyRole) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":  "waitq",
		"ref":  id,
		"role": role,
		"exp":  time.Now().Add(time.Hour * 24 * 365 * 10).Unix(),
		"iat":  time.Now().Unix(),
	})

	// Ensure the secret is a byte slice for HMAC
	hmacSecret := []byte(secret)

	// Sign and get the complete encoded token as a string using the HMAC secret
	return token.SignedString(hmacSecret)
}

func generateKeys(secret string, id uuid.UUID) (Keys, error) {
	keys := Keys{}

	var err error

	keys.AnonKey, err = generateToken(secret, id, AnonKey)
	if err != nil {
		return keys, err
	}

	keys.ServiceKey, err = generateToken(secret, id, ServiceKey)
	if err != nil {
		return keys, err
	}

	return keys, nil
}

func (s *WaitlistService) IsURLAliasAvailable(ctx context.Context, urlAlias string) (bool, error) {
	return s.repositories.Waitlist().IsURLAliasAvailable(ctx, urlAlias)
}

func (s *WaitlistService) GetByURLAlias(ctx context.Context, urlAlias string) (waitlist.Waitlist, error) {
	return s.repositories.Waitlist().GetByURLAlias(ctx, urlAlias)
}

func (s *WaitlistService) GetPublicMany(ctx context.Context, input pgShared.CursorPaginationRequest) ([]waitlist.PublicWaitlist, error) {
	return s.repositories.Waitlist().GetPublicMany(ctx, input)
}
