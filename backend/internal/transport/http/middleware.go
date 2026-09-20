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

// publicOps are operations served without initData (see openapi.yaml `security: []`).
var publicOps = map[string]bool{"Health": true, "GetPhoto": true}

// authMiddleware validates `Authorization: tma <initData>`, upserts the user and puts
// model.AppUser into the request context. Writes 401 itself and returns (nil, nil) on failure.
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
			// StrictServerInterface handlers receive *gin.Context as context.Context (it
			// implements the interface); its Value() falls through to the request's context
			// for non-string keys, so the user must live there, not in gin's own key/value store.
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxKey{}, u))
			return next(c, req)
		}
	}
}

func userFromCtx(ctx context.Context) model.AppUser { return ctx.Value(ctxKey{}).(model.AppUser) }

// dispatcherMiddleware: uk-cabinet operations require role uk_dispatcher.
var dispatcherOps = map[string]bool{"SetIncidentStatus": true, "UkQueue": true, "UkHouses": true, "UkCreateEvent": true}

func dispatcherMiddleware(next oapi.StrictHandlerFunc, opID string) oapi.StrictHandlerFunc {
	if !dispatcherOps[opID] {
		return next
	}
	return func(c *gin.Context, req any) (any, error) {
		if userFromCtx(c).Role != "uk_dispatcher" {
			writeError(c.Writer, http.StatusForbidden, "forbidden", "dispatcher only")
			return nil, nil
		}
		return next(c, req)
	}
}
