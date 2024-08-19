package middleware

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"waitq/api/internal/server/http/handler/shared"
	"waitq/api/internal/service"

	"github.com/danielgtaylor/huma/v2"
	"go.uber.org/zap"
)

func WithWaitlistOwnerSubscription(api huma.API) func(ctx huma.Context, next func(huma.Context), sv *service.Service) {
	return func(ctx huma.Context, next func(huma.Context), sv *service.Service) {
		reqCtx := ctx.Context()

		waitlistKey := shared.GetWaitlistKey(reqCtx)

		waitlist, err := sv.WaitlistService.GetById(reqCtx, waitlistKey.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				sv.Logger.Error("waitlist not found", zap.Error(err))
				huma.WriteErr(api, ctx, http.StatusNotFound,
					"Waitlist not found",
				)
				return
			}
			sv.Logger.Error("failed to get waitlist", zap.Error(err))
			huma.WriteErr(api, ctx, http.StatusInternalServerError,
				"Failed to get waitlist",
			)
			return
		}

		sub, err := sv.SubscriptionService.GetAccountSubscription(reqCtx, waitlist.AccountID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				sv.Logger.Error("account subscription not found", zap.Error(err))
			}
			sv.Logger.Error("failed to get account subscription", zap.Error(err))
			next(huma.WithValue(ctx, shared.WaitlistSubscriptionContextKey, nil))
			return
		}

		log.Println("sub", sub)

		next(huma.WithValue(ctx, shared.WaitlistSubscriptionContextKey, &sub))
	}
}
