package middleware

import (
	"errors"
	"net/http"
	"strings"
	"waitq/api/internal/server/http/handler/shared"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type KeyRole string

const (
	WaitlistServiceKey KeyRole = "service"
	WaitlistAnonKey    KeyRole = "anon"
)

func WithWaitlistServiceKey(api huma.API) func(ctx huma.Context, next func(huma.Context), logger *zap.Logger) {
	return func(ctx huma.Context, next func(huma.Context), logger *zap.Logger) {
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

		waitlistID, err := parseTokenClaims(token, WaitlistServiceKey, logger)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				err.Error(),
			)
			return
		}

		next(huma.WithValue(ctx, shared.WaitlistServiceKeyContextKey, waitlistID))
	}
}

func WithWaitlistAnonKey(api huma.API) func(ctx huma.Context, next func(huma.Context), logger *zap.Logger) {
	return func(ctx huma.Context, next func(huma.Context), logger *zap.Logger) {
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

		waitlistID, err := parseTokenClaims(token, WaitlistAnonKey, logger)
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				err.Error(),
			)
			return
		}

		next(huma.WithValue(ctx, shared.WaitlistAnonKeyContextKey, waitlistID))
	}
}

// parseTokenClaims parses the JWT token and returns the waitlist ID
func parseTokenClaims(token *jwt.Token, role KeyRole, logger *zap.Logger) (string, error) {
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		logger.Debug("claims", zap.Any("claims", claims))

		waitlistID, ok := claims["ref"].(string)
		if !ok {
			logger.Error("waitlistID claim is missing or not a string", zap.Any("claims", claims))
			return "", errors.New("Token does not contain a valid waitlist ID")
		}

		roleClaim, ok := claims["role"].(string)
		if !ok {
			logger.Error("role claim is missing or not a string", zap.Any("claims", claims))
			return "", errors.New("Token does not contain a valid role")
		}

		logger.Error("role claim", zap.Any("role", roleClaim), zap.Any("role", role))
		if roleClaim != string(role) {
			logger.Error("waitlist api key JWT does not have the correct role", zap.Any("token", token))
			return "", errors.New("Token does not have the correct role")
		}

		return waitlistID, nil
	} else {
		logger.Error("Error parsing waitlist service key JWT claims", zap.Any("token", token))
		return "", errors.New("An invalid access token was provided")
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
