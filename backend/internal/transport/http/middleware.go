package http

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	oapi "ukapp/gen/api"
	"ukapp/internal/domain"
	"ukapp/internal/service"
	"ukapp/internal/ukauth"

	"ukapp/gen/db/ukapp/public/model"
)

type ctxKey struct{}

type maxUserKey struct{}

var publicOps = map[string]bool{"Health": true, "GetPhoto": true, "HouseSticker": true}

var dispatcherOps = map[string]bool{"ListUkQueue": true, "CreateUkEvent": true, "SetIncidentStatus": true, "MergeIncidents": true,
	"CreateUkApiKey": true, "ListUkApiKeys": true, "RevokeUkApiKey": true, "ListUkIncidentChanges": true}

var jwtOnlyOps = map[string]bool{"CreateUkApiKey": true, "ListUkApiKeys": true, "RevokeUkApiKey": true}

var residentOps = map[string]bool{
	"BindHouse": true, "UnbindHouse": true, "CreateIncident": true, "JoinIncident": true,
	"ConfirmIncident": true, "UploadIncidentPhoto": true,
}

func authMiddleware(svc *service.Service, botToken string) oapi.StrictMiddlewareFunc {
	return func(next oapi.StrictHandlerFunc, opID string) oapi.StrictHandlerFunc {
		if publicOps[opID] {
			return next
		}
		if opID == "EsiaMockLogin" {
			return func(c *gin.Context, req any) (any, error) {
				if raw, ok := strings.CutPrefix(c.GetHeader("Authorization"), "tma "); ok {
					if iu, err := domain.ValidateInitData(raw, botToken); err == nil {
						c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), maxUserKey{}, iu.ID))
					}
				}
				return next(c, req)
			}
		}
		return func(c *gin.Context, req any) (any, error) {
			u, viaKey, err := authenticate(c, svc, botToken)
			if errors.Is(err, errUnauthorized) {
				writeError(c.Writer, http.StatusUnauthorized, "unauthorized", "unauthorized")
				return nil, nil
			}
			if errors.Is(err, service.ErrRateLimited) {
				writeError(c.Writer, http.StatusTooManyRequests, "rate_limited", "не больше 10 запросов в секунду на ключ")
				return nil, nil
			}
			if err != nil {
				return nil, err
			}
			if viaKey && jwtOnlyOps[opID] {
				writeError(c.Writer, http.StatusForbidden, "forbidden", "ключи управляются только из кабинета")
				return nil, nil
			}
			if dispatcherOps[opID] && u.Role != domain.RoleUkDispatcher || residentOps[opID] && u.Role != domain.RoleResident {
				writeError(c.Writer, http.StatusForbidden, "forbidden", "недостаточно прав")
				return nil, nil
			}
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxKey{}, u))
			return next(c, req)
		}
	}
}

var errUnauthorized = errors.New("unauthorized")

func authenticate(c *gin.Context, svc *service.Service, botToken string) (model.AppUser, bool, error) {
	h := c.GetHeader("Authorization")
	if raw, ok := strings.CutPrefix(h, "Bearer "); ok {
		viaKey := strings.HasPrefix(raw, ukauth.APIKeyPrefix)
		var u model.AppUser
		var err error
		if viaKey {
			u, err = svc.AuthenticateApiKey(c.Request.Context(), raw)
		} else {
			u, err = svc.AuthenticateUkToken(c.Request.Context(), raw)
		}
		if errors.Is(err, service.ErrForbidden) {
			return u, viaKey, errUnauthorized
		}
		return u, viaKey, err
	}
	raw, ok := strings.CutPrefix(h, "tma ")
	if !ok {
		return model.AppUser{}, false, errUnauthorized
	}
	iu, err := domain.ValidateInitData(raw, botToken)
	if err != nil {
		return model.AppUser{}, false, errUnauthorized
	}
	u, err := svc.UpsertUser(c.Request.Context(), iu)
	if err != nil {
		return u, false, err
	}
	if houseID, ok := parseID("h_", iu.StartParam); ok {
		u, err = svc.BindHouseFromSticker(c.Request.Context(), u, houseID)
	}
	return u, false, err
}

func userFromCtx(ctx context.Context) model.AppUser { return ctx.Value(ctxKey{}).(model.AppUser) }

func maxUserFromCtx(ctx context.Context) *int64 {
	if id, ok := ctx.Value(maxUserKey{}).(int64); ok {
		return &id
	}
	return nil
}
