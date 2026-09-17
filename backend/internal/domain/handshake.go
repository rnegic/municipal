package domain

const (
	ConfirmThreshold = 0.5 // доля подписчиков дома, подтвердивших «починили»
	MinConfirmations = 2   // и не меньше стольки людей
)

// ShouldClose: достаточно жителей подтвердили, что проблема решена —
// >= ConfirmThreshold от подписчиков и не меньше MinConfirmations.
func ShouldClose(subscribers, confirmations int) bool {
	if confirmations < MinConfirmations || subscribers <= 0 {
		return false
	}
	return float64(confirmations)/float64(subscribers) >= ConfirmThreshold
}
