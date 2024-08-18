package middleware

import (
	"net/http"
	"waitq/api/internal/server/http/handler/shared"
	postgres "waitq/api/internal/storage/postgres/shared"

	"waitq/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/supabase-community/supabase-go"
	"go.uber.org/zap"
)

func WithUser(api huma.API) func(ctx huma.Context, next func(huma.Context), sb *supabase.Client, logger *zap.Logger) {
	return func(ctx huma.Context, next func(huma.Context), sb *supabase.Client, logger *zap.Logger) {
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

		authedClient := sb.Auth.WithToken(accessToken)

		resp, err := authedClient.GetUser()
		if err != nil {
			logger.Error("Error getting user", zap.Error(err))
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				"An invalid access token was provided",
			)
			return
		}

		next(huma.WithValue(ctx, shared.UserContextKey, resp.User))
	}
}

func WithAccount(api huma.API) func(ctx huma.Context, next func(huma.Context), as service.AccountService, logger *zap.Logger) {
	return func(ctx huma.Context, next func(huma.Context), as service.AccountService, logger *zap.Logger) {
		user := shared.GetAuthenticatedUser(ctx.Context())
		if user.ID == uuid.Nil {
			huma.WriteErr(api, ctx, http.StatusUnauthorized,
				"User not authenticated",
			)
			return
		}

		queryResp, err := as.GetByUserId(ctx.Context(), user.ID, postgres.GetManyRequest{
			IncludeDeleted: false,
		})
		if err != nil {
			logger.Error("Error getting account", zap.Error(err))
			huma.WriteErr(api, ctx, http.StatusInternalServerError,
				"Something went wrong",
			)
			return
		}

		// There should only be one non-deleted account per user
		if len(queryResp) == 0 {
			logger.Error("User has no accounts", zap.Any("user", user))
			huma.WriteErr(api, ctx, http.StatusForbidden,
				"User does not have an account",
			)
			return
		} else if len(queryResp) > 1 {
			logger.Error("User has multiple accounts", zap.Any("user", user), zap.Any("accounts", queryResp))
			huma.WriteErr(api, ctx, http.StatusInternalServerError,
				"Something went wrong",
			)
			return
		}

		next(huma.WithValue(ctx, shared.AccountContextKey, queryResp[0]))
	}
}
