package ukauth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"ukapp/internal/domain"
)

func TestPasswordMatches(t *testing.T) {
	h, _ := bcrypt.GenerateFromPassword([]byte("admin2026"), bcrypt.MinCost)
	hash := string(h)
	if !PasswordMatches(&hash, "admin2026") || PasswordMatches(&hash, "admin") || PasswordMatches(nil, "admin2026") {
		t.Fatal("PasswordMatches")
	}
}

func TestTokenSigner(t *testing.T) {
	signer, err := NewTokenSigner("")
	if err != nil {
		t.Fatal(err)
	}
	tok, exp, err := signer.Issue(42, 7, "1655000003", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	got, err := signer.Parse(tok)
	if err != nil || got.Subject != "42" || got.UkID != 7 || got.ID == "" || !got.ExpiresAt.Equal(exp) {
		t.Fatalf("round trip: %+v %v", got, err)
	}

	other, _ := NewTokenSigner("")
	if _, err := other.Parse(tok); err == nil {
		t.Error("token of another key must be rejected")
	}
	parts := strings.Split(tok, ".")
	if _, err := signer.Parse(parts[0] + "." + parts[1] + "x." + parts[2]); err == nil {
		t.Error("tampered payload must be rejected")
	}
	expired, _, _ := signer.Issue(42, 7, "1655000003", time.Now().Add(-domain.UkTokenTTL-time.Minute))
	if _, err := signer.Parse(expired); err == nil {
		t.Error("expired token must be rejected")
	}
	resident := jwt.NewWithClaims(jwt.SigningMethodES256, Claims{Role: domain.RoleResident, UkID: 7,
		RegisteredClaims: jwt.RegisteredClaims{Subject: "42", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}})
	resident.Header["kid"] = signer.kid
	raw, _ := resident.SignedString(signer.key)
	if _, err := signer.Parse(raw); err == nil {
		t.Error("non-dispatcher token must be rejected")
	}
}
