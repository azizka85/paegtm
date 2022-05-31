package helpers

import (
	"math"
)

func NumTruncate(num float64, digits int) float64 {
	base := math.Pow(10, float64(digits))

	return math.Round(num*base) / base
}
