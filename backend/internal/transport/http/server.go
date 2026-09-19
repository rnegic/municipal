// Package http — реализация oapi.StrictServerInterface на gin: разбор запроса, вызов service,
// маппинг результата в JSON-контракт из openapi.yaml. Бизнес-решений здесь нет.
package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	oapi "ukapp/gen/api"
	"ukapp/internal/service"
)

// Middlewares are applied in reverse order: auth wraps dispatcher wraps handler.
// server implements oapi.StrictServerInterface. Handlers live in per-feature files
// (users.go, incidents.go, events.go, uk.go).
type server struct {
	svc *service.Service
}

func NewServer(svc *service.Service, botToken string) http.Handler {
	srv := &server{svc: svc}
	h := oapi.NewStrictHandlerWithOptions(srv, []oapi.StrictMiddlewareFunc{dispatcherMiddleware, authMiddleware(svc, botToken)}, oapi.StrictGinServerOptions{
		RequestErrorHandlerFunc: func(c *gin.Context, err error) {
			writeError(c.Writer, http.StatusBadRequest, "validation_failed", err.Error())
		},
		HandlerErrorFunc: func(c *gin.Context, err error) {
			slog.Error("handler failed", "method", c.Request.Method, "path", c.Request.URL.Path, "err", err)
			writeError(c.Writer, http.StatusInternalServerError, "internal", "internal error")
		},
		ResponseErrorHandlerFunc: func(c *gin.Context, err error) {
			slog.Error("response encode failed", "method", c.Request.Method, "path", c.Request.URL.Path, "err", err)
		},
	})
	router := gin.New()
	// StrictServerInterface handlers only see *gin.Context through the context.Context
	// interface (Value/Deadline/Done/Err); without this it doesn't fall back to the
	// request's own context, so context.WithValue(c.Request.Context(), ...) in
	// authMiddleware would be invisible to userFromCtx.
	router.ContextWithFallback = true
	oapi.RegisterHandlers(router, h)
	return withCORS(router)
}

// withCORS: фронт живёт в вебвью MAX на другом origin — без этих заголовков вебвью
// заблокирует запрос (см. docs/frontend-api-contract.md §8).
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, httpCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	_ = json.NewEncoder(w).Encode(oapi.ApiErrorResponse{Code: code, Message: message})
}

func apiErr(code, message string) oapi.ApiErrorResponse {
	return oapi.ApiErrorResponse{Code: code, Message: message}
}

func (s *server) Health(context.Context, oapi.HealthRequestObject) (oapi.HealthResponseObject, error) {
	return oapi.Health200TextResponse("ok"), nil
}
