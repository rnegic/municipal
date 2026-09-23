package ukauth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"ukapp/internal/domain"
)

var dummyPasswordHash, _ = bcrypt.GenerateFromPassword([]byte("dummy"), bcrypt.DefaultCost)

func PasswordMatches(hash *string, password string) bool {
	if hash == nil {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(*hash), []byte(password)) == nil
}

type Claims struct {
	Role    string  `json:"role"`
	UkID    int64   `json:"uk_id"`
	OrgINN  string  `json:"org_inn"`
	EsiaOID *string `json:"esia_oid"`
	jwt.RegisteredClaims
}

type TokenSigner struct {
	key *ecdsa.PrivateKey
	kid string
}

func NewTokenSigner(pemKey string) (*TokenSigner, error) {
	var key *ecdsa.PrivateKey
	var err error
	if pemKey == "" {
		key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	} else {
		key, err = parseECKey(pemKey)
	}
	if err != nil {
		return nil, err
	}
	pub, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(pub)
	return &TokenSigner{key: key, kid: hex.EncodeToString(sum[:8])}, nil
}

func parseECKey(pemKey string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, errors.New("uk token key: no PEM block")
	}
	if k, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return k, nil
	}
	k, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	ec, ok := k.(*ecdsa.PrivateKey)
	if !ok || ec.Curve != elliptic.P256() {
		return nil, errors.New("uk token key: want ECDSA P-256")
	}
	return ec, nil
}

func (t *TokenSigner) Issue(userID, ukID int64, orgINN string, now time.Time) (string, time.Time, error) {
	c := Claims{Role: domain.RoleUkDispatcher, UkID: ukID, OrgINN: orgINN}
	c.Subject = strconv.FormatInt(userID, 10)
	jti := make([]byte, 16)
	if _, err := rand.Read(jti); err != nil {
		return "", time.Time{}, err
	}
	exp := now.Add(domain.UkTokenTTL).UTC().Truncate(time.Second)
	c.IssuedAt = jwt.NewNumericDate(now)
	c.ExpiresAt = jwt.NewNumericDate(exp)
	c.ID = hex.EncodeToString(jti)
	tok := jwt.NewWithClaims(jwt.SigningMethodES256, c)
	tok.Header["kid"] = t.kid
	s, err := tok.SignedString(t.key)
	return s, exp, err
}

func (t *TokenSigner) Parse(raw string) (Claims, error) {
	var c Claims
	_, err := jwt.ParseWithClaims(raw, &c, func(tok *jwt.Token) (any, error) {
		if tok.Header["kid"] != t.kid {
			return nil, errors.New("unknown kid")
		}
		return &t.key.PublicKey, nil
	}, jwt.WithValidMethods([]string{"ES256"}), jwt.WithExpirationRequired())
	if err != nil {
		return Claims{}, err
	}
	if c.Role != domain.RoleUkDispatcher || c.UkID == 0 || c.Subject == "" {
		return Claims{}, errors.New("uk token: bad claims")
	}
	return c, nil
}
