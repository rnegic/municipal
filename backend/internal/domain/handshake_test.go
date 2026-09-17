package domain

import "testing"

func TestShouldClose(t *testing.T) {
	cases := []struct {
		name          string
		subscribers   int
		confirmations int
		want          bool
	}{
		{"1 of 2 = 50% but under min count is fine (2 confirmations, 2 subs)", 2, 2, true},
		{"1 of 4 = 25%, under threshold", 4, 1, false},
		{"1 confirmation only, under min count even if 100%", 1, 1, false},
		{"2 of 4 = 50% and min count met", 4, 2, true},
		{"3 of 4 = 75%", 4, 3, true},
		{"zero subscribers never closes", 0, 5, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ShouldClose(c.subscribers, c.confirmations); got != c.want {
				t.Fatalf("got %v want %v", got, c.want)
			}
		})
	}
}
