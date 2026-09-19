// Package lk — минимальный личный кабинет диспетчера УК: вход по логину/паролю из сида,
// список обращений организации, смена статуса. html/template, cookie-сессии в памяти.
package lk

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/hex"
	"errors"
	"html/template"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"

	"mockuk/internal/store"
)

//go:embed templates/*.html
var tmplFS embed.FS

var tmpl = template.Must(template.New("").Funcs(template.FuncMap{"str": func(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}}).ParseFS(tmplFS, "templates/*.html"))

const cookieName = "lk_session"

type lk struct {
	st       *store.Store
	mu       sync.Mutex
	sessions map[string]store.Dispatcher // ponytail: сессии в памяти, рестарт = разлогин
}

func Mount(r gin.IRouter, st *store.Store) {
	l := &lk{st: st, sessions: map[string]store.Dispatcher{}}
	r.GET("/lk/login", func(c *gin.Context) { l.render(c, http.StatusOK, "login.html", gin.H{}) })
	r.POST("/lk/login", l.login)
	r.POST("/lk/logout", l.logout)
	r.GET("/lk", l.index)
	r.POST("/lk/incidents/:id/status", l.setStatus)
}

func (l *lk) render(c *gin.Context, code int, name string, data any) {
	c.Status(code)
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(c.Writer, name, data); err != nil {
		_ = c.Error(err)
	}
}

func (l *lk) login(c *gin.Context) {
	d, err := l.st.Dispatcher(c.Request.Context(), c.PostForm("login"))
	sum := sha256.Sum256([]byte(c.PostForm("password")))
	if err != nil || subtle.ConstantTimeCompare([]byte(hex.EncodeToString(sum[:])), []byte(d.PasswordSHA256)) != 1 {
		l.render(c, http.StatusUnauthorized, "login.html", gin.H{"Error": "Неверный логин или пароль"})
		return
	}
	var raw [16]byte
	_, _ = rand.Read(raw[:])
	tok := hex.EncodeToString(raw[:])
	l.mu.Lock()
	l.sessions[tok] = d
	l.mu.Unlock()
	c.SetCookie(cookieName, tok, 86400, "/", "", false, true)
	c.Redirect(http.StatusFound, "/lk")
}

func (l *lk) logout(c *gin.Context) {
	if tok, err := c.Cookie(cookieName); err == nil {
		l.mu.Lock()
		delete(l.sessions, tok)
		l.mu.Unlock()
	}
	c.SetCookie(cookieName, "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/lk/login")
}

// current returns the dispatcher for the session cookie or redirects to login (ok=false).
func (l *lk) current(c *gin.Context) (store.Dispatcher, bool) {
	tok, err := c.Cookie(cookieName)
	if err == nil {
		l.mu.Lock()
		d, ok := l.sessions[tok]
		l.mu.Unlock()
		if ok {
			return d, true
		}
	}
	c.Redirect(http.StatusFound, "/lk/login")
	return store.Dispatcher{}, false
}

func (l *lk) index(c *gin.Context) {
	d, ok := l.current(c)
	if !ok {
		return
	}
	items, err := l.st.ListByOrg(c.Request.Context(), d.OrgID)
	if err != nil {
		_ = c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	l.render(c, http.StatusOK, "index.html", gin.H{"Login": d.Login, "OrgName": d.OrgID, "Incidents": items})
}

func (l *lk) setStatus(c *gin.Context) {
	if _, ok := l.current(c); !ok {
		return
	}
	switch c.PostForm("status") {
	case "accepted", "in_progress", "verifying", "done":
	default:
		c.String(http.StatusBadRequest, "неизвестный статус")
		return
	}
	_, err := l.st.SetStatus(c.Request.Context(), c.Param("id"), c.PostForm("status"))
	switch {
	case errors.Is(err, store.ErrNotFound):
		c.String(http.StatusNotFound, "нет такого обращения")
	case errors.Is(err, store.ErrConflict):
		c.String(http.StatusConflict, "обращение уже закрыто")
	case err != nil:
		_ = c.AbortWithError(http.StatusInternalServerError, err)
	default:
		c.Redirect(http.StatusFound, "/lk")
	}
}
