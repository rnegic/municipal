package domain

const (
	ConfirmThreshold = 0.5
	MinConfirmations = 2
)

func ShouldClose(subscribers, confirmations int) bool {
	if confirmations < MinConfirmations || subscribers <= 0 {
		return false
	}
	return float64(confirmations)/float64(subscribers) >= ConfirmThreshold
}
