package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const MaxInitDataAge = 24 * time.Hour

type InitUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	PhotoURL  string `json:"photo_url"`
}

func ValidateInitData(raw, botToken string) (InitUser, error) {
	q, err := url.ParseQuery(raw)
	if err != nil {
		return InitUser{}, err
	}
	gotHash := q.Get("hash")
	if gotHash == "" {
		return InitUser{}, errors.New("initData: no hash")
	}
	q.Del("hash")
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + "=" + q.Get(k)
	}
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(botToken))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	mac.Write([]byte(strings.Join(parts, "\n")))
	if !hmac.Equal([]byte(hex.EncodeToString(mac.Sum(nil))), []byte(gotHash)) {
		return InitUser{}, errors.New("initData: bad hash")
	}
	authDate, err := strconv.ParseInt(q.Get("auth_date"), 10, 64)
	if err != nil {
		return InitUser{}, errors.New("initData: bad auth_date")
	}
	if time.Since(time.Unix(authDate, 0)) > MaxInitDataAge {
		return InitUser{}, errors.New("initData: expired")
	}
	var u InitUser
	if err := json.Unmarshal([]byte(q.Get("user")), &u); err != nil || u.ID == 0 {
		return InitUser{}, errors.New("initData: bad user")
	}
	return u, nil
}
