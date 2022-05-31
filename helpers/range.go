package helpers

func RangeMake(min, max int) []int {
	r := make([]int, max-min+1)

	for i := range r {
		r[i] = min + i
	}

	return r
}
