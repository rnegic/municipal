package domain

import "time"

const (
	RoleResident     = "resident"
	RoleUkDispatcher = "uk_dispatcher"

	AuthMethodEsiaMock = "esia_mock"

	UkTokenTTL         = 8 * time.Hour
	UkLoginWindow      = 15 * time.Minute
	UkLoginMaxFailures = 5

	UkMaxApiKeys          = 10
	UkApiKeyRPS           = 10
	UkIntegrationUserName = "1С-интеграция"
)

var innOrgWeights = [9]int{2, 4, 10, 3, 5, 9, 4, 6, 8}

func ValidOrgINN(inn string) bool {
	if len(inn) != 10 {
		return false
	}
	sum := 0
	for i, r := range inn {
		if r < '0' || r > '9' {
			return false
		}
		if i < len(innOrgWeights) {
			sum += int(r-'0') * innOrgWeights[i]
		}
	}
	return sum%11%10 == int(inn[9]-'0')
}

func LicenseActive(number *string, validUntil *time.Time, now time.Time) bool {
	return number != nil && validUntil != nil && !validUntil.Before(now.Truncate(24*time.Hour))
}
