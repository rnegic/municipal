package ukauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const APIKeyPrefix = "uk_live_"

func NewAPIKey() (key, prefix string, hash []byte, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", nil, err
	}
	body := base64.RawURLEncoding.EncodeToString(b)
	key = APIKeyPrefix + body
	return key, body[:8], HashAPIKey(key), nil
}

func HashAPIKey(key string) []byte {
	sum := sha256.Sum256([]byte(key))
	return sum[:]
}
