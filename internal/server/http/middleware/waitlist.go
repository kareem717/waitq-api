package middleware

import (
	"errors"
	"net/http"
	"strings"
	"waitq/api/internal/server/http/handler/shared"
	"waitq/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func WithWaitlistServiceKey(api huma.API) func(ctx huma.Context, next func(huma.Context), sv *service.Service) {
	return func(ctx huma.Context, next func(huma.Context), sv *service.Service) {
		authHeader := ctx.Header("Authorization")
		if authHeader == "" {
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				"No authorization header was provided",
			)
			return
		}

		accessToken, err := parseBearerToken(authHeader)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				err.Error(),
			)
			return
		}

		token, _ := jwt.Parse(accessToken, nil)
		if token == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				"An invalid access token was provided",
			)
			return
		}

		waitlistKey, err := parseTokenClaims(token, sv.Logger)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				err.Error(),
			)
			return

		}

		if waitlistKey.Role != shared.WaitlistServiceKey {
			sv.Logger.Error("Invalid access token", zap.Any("waitlistKey", waitlistKey))
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				"An invalid access token was provided",
			)
			return
		}

		next(huma.WithValue(ctx, shared.WaitlistKeyContextKey, waitlistKey))
	}
}

func WithWaitlistAnonKey(api huma.API) func(ctx huma.Context, next func(huma.Context), sv *service.Service) {
	return func(ctx huma.Context, next func(huma.Context), sv *service.Service) {
		authHeader := ctx.Header("Authorization")
		if authHeader == "" {
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				"No authorization header was provided",
			)
			return
		}

		accessToken, err := parseBearerToken(authHeader)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				err.Error(),
			)
			return
		}

		token, _ := jwt.Parse(accessToken, nil)
		if token == nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				"An invalid access token was provided",
			)
			return
		}

		waitlistKey, err := parseTokenClaims(token, sv.Logger)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				err.Error(),
			)
			return
		}

		if waitlistKey.Role != shared.WaitlistAnonKey {
			sv.Logger.Error("Invalid access token", zap.Any("waitlistKey", waitlistKey))
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				"An invalid access token was provided",
			)
			return
		}

		next(huma.WithValue(ctx, shared.WaitlistKeyContextKey, waitlistKey))
	}
}

// parseTokenClaims parses the JWT token and returns the waitlist ID
func parseTokenClaims(token *jwt.Token, logger *zap.Logger) (shared.WaitlistKey, error) {
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		logger.Debug("claims", zap.Any("claims", claims))

		waitlistID, ok := claims["ref"].(string)
		if !ok {
			logger.Error("waitlistID claim is missing or not a string", zap.Any("claims", claims))
			return shared.WaitlistKey{}, errors.New("Token does not contain a valid waitlist ID")
		}

		roleClaim, ok := claims["role"].(string)
		if !ok {
			logger.Error("role claim is missing or not a string", zap.Any("claims", claims))
			return shared.WaitlistKey{}, errors.New("Token does not contain a valid role")
		}

		parsedId, err := uuid.Parse(waitlistID)
		if err != nil {
			logger.Error("waitlistID claim is not a valid UUID", zap.Any("claims", claims))
			return shared.WaitlistKey{}, errors.New("Token does not contain a valid waitlist ID")
		}

		return shared.WaitlistKey{
			ID:   parsedId,
			Role: shared.KeyRole(roleClaim),
		}, nil
	} else {
		logger.Error("Error parsing waitlist service key JWT claims", zap.Any("token", token))
		return shared.WaitlistKey{}, errors.New("An invalid access token was provided")
	}
}

func parseBearerToken(token string) (string, error) {
	var accessToken string
	accessToken = strings.Replace(token, "Bearer ", "", 1)
	if accessToken == "" {
		return "", errors.New("An invalid access token was provided")
	}

	return accessToken, nil
}
