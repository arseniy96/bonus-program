package converter

const multiplier = 100

func ConvertToCent(amount float64) int {
	return int(amount * multiplier)
}

func ConvertFromCent(amount int) float64 {
	return float64(amount) / multiplier
}
