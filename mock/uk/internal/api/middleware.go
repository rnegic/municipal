package api

import (
	"crypto/subtle"
	"math/rand/v2"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	ukapi "mockuk/gen/api"
)

func writeErr(c *gin.Context, code int, kind, msg string) {
	c.AbortWithStatusJSON(code, ukapi.Error{Code: kind, Message: msg})
}

func bearer(token string) ukapi.MiddlewareFunc {
	return func(c *gin.Context) {
		got, ok := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !ok || subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
			writeErr(c, http.StatusUnauthorized, "unauthorized", "invalid bearer token")
			return
		}
		c.Next()
	}
}

// faults: MOCK_FAILURE_RATE — доля запросов, падающих 500, чтобы показать устойчивость
// бэкенда к недоступной УК. ponytail: только 500; таймауты/504 — добавить sleep здесь же.
func faults(rate float64) ukapi.MiddlewareFunc {
	return func(c *gin.Context) {
		if rate > 0 && rand.Float64() < rate { //nolint:gosec // не криптография
			writeErr(c, http.StatusInternalServerError, "internal", "injected failure")
			return
		}
		c.Next()
	}
}
