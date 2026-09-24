package domain

import (
	"testing"
	"time"
)

func TestValidOrgINN(t *testing.T) {
	for inn, want := range map[string]bool{
		"1655000003": true, "7707083893": true, "1655000000": false,
		"165500000": false, "165500000a": false, "500100732259": false,
	} {
		if got := ValidOrgINN(inn); got != want {
			t.Errorf("ValidOrgINN(%q) = %v, want %v", inn, got, want)
		}
	}
}

func TestLicenseActive(t *testing.T) {
	now := time.Date(2026, 9, 23, 15, 0, 0, 0, time.UTC)
	num := "16-000123"
	today := time.Date(2026, 9, 23, 0, 0, 0, 0, time.UTC)
	yesterday := today.AddDate(0, 0, -1)
	if !LicenseActive(&num, &today, now) {
		t.Error("license valid until today must be active")
	}
	if LicenseActive(&num, &yesterday, now) || LicenseActive(nil, &today, now) || LicenseActive(&num, nil, now) {
		t.Error("expired or incomplete license must be inactive")
	}
}
