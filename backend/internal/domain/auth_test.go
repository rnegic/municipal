package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"testing"
	"time"
)

const testBotToken = "test-bot-token"

// signInitData builds a valid initData string the way MAX does (see docs/webapps/validation).
func signInitData(t *testing.T, pairs map[string]string) string {
	t.Helper()
	keys := make([]string, 0, len(pairs))
	for k := range pairs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	check := ""
	for i, k := range keys {
		if i > 0 {
			check += "\n"
		}
		check += k + "=" + pairs[k]
	}
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(testBotToken))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	mac.Write([]byte(check))
	hash := hex.EncodeToString(mac.Sum(nil))
	q := url.Values{}
	for k, v := range pairs {
		q.Set(k, v)
	}
	q.Set("hash", hash)
	return q.Encode()
}

func nowAuthDate() string { return strconv.FormatInt(time.Now().Unix(), 10) }

func TestValidateInitData_OK(t *testing.T) {
	raw := signInitData(t, map[string]string{
		"auth_date": nowAuthDate(),
		"query_id":  "4c0ab423-342b-4e45-aea4-2747dbc500cd",
		"user":      `{"id":67890,"first_name":"Max","last_name":"Ivanov"}`,
	})
	u, err := ValidateInitData(raw, testBotToken)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID != 67890 || u.FirstName != "Max" || u.LastName != "Ivanov" {
		t.Fatalf("bad user: %+v", u)
	}
}

func TestValidateInitData_BadHash(t *testing.T) {
	raw := signInitData(t, map[string]string{"auth_date": nowAuthDate(), "user": `{"id":1}`})
	if _, err := ValidateInitData(raw, "other-token"); err == nil {
		t.Fatal("expected error for wrong token")
	}
}

func TestValidateInitData_NoHash(t *testing.T) {
	if _, err := ValidateInitData("auth_date="+nowAuthDate()+"&user=%7B%22id%22%3A1%7D", testBotToken); err == nil {
		t.Fatal("expected error when hash missing")
	}
}

func TestValidateInitData_Expired(t *testing.T) {
	old := strconv.FormatInt(time.Now().Add(-25*time.Hour).Unix(), 10)
	raw := signInitData(t, map[string]string{"auth_date": old, "user": `{"id":1}`})
	if _, err := ValidateInitData(raw, testBotToken); err == nil {
		t.Fatal("expected error for stale auth_date")
	}
}
