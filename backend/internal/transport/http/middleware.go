package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	oapi "ukapp/gen/api"
	"ukapp/internal/domain"
	"ukapp/internal/service"

	"ukapp/gen/db/ukapp/public/model"
)

type ctxKey struct{}

var publicOps = map[string]bool{"Health": true, "GetPhoto": true}

func authMiddleware(svc *service.Service, botToken string) oapi.StrictMiddlewareFunc {
	return func(next oapi.StrictHandlerFunc, opID string) oapi.StrictHandlerFunc {
		if publicOps[opID] {
			return next
		}
		return func(c *gin.Context, req any) (any, error) {
			raw, ok := strings.CutPrefix(c.GetHeader("Authorization"), "tma ")
			if !ok {
				writeError(c.Writer, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return nil, nil
			}
			iu, err := domain.ValidateInitData(raw, botToken)
			if err != nil {
				writeError(c.Writer, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return nil, nil //nolint:nilerr // 401 уже записан, наружу ошибка не нужна
			}
			u, err := svc.UpsertUser(c.Request.Context(), iu)
			if err != nil {
				return nil, err
			}

			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxKey{}, u))
			return next(c, req)
		}
	}
}

func userFromCtx(ctx context.Context) model.AppUser { return ctx.Value(ctxKey{}).(model.AppUser) }
