package ukauth

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewAPIKey(t *testing.T) {
	key, prefix, hash, err := NewAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(key, APIKeyPrefix) || len(key) != len(APIKeyPrefix)+43 {
		t.Fatalf("key format: %q", key)
	}
	if prefix != key[len(APIKeyPrefix):len(APIKeyPrefix)+8] {
		t.Fatalf("prefix %q for %q", prefix, key)
	}
	if !bytes.Equal(hash, HashAPIKey(key)) || len(hash) != 32 {
		t.Fatalf("hash mismatch")
	}
	other, _, _, _ := NewAPIKey()
	if other == key {
		t.Fatal("keys must differ")
	}
}
