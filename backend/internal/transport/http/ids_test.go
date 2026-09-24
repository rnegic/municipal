package http

import "testing"

func TestParseIDStrict(t *testing.T) {
	for _, s := range []string{"1", "ph_+1", "ph_-1", "ph_0", "ph_ 1", "ph_1x", "inc_1"} {
		if _, ok := parseID("ph_", s); ok {
			t.Errorf("%q must be rejected", s)
		}
	}
	if id, ok := parseID("ph_", "ph_42"); !ok || id != 42 {
		t.Fatalf("ph_42: %d %v", id, ok)
	}
}
