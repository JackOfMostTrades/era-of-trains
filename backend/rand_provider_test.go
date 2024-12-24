package main

import "fmt"

type fixedRandProvider struct {
	idx    int
	values []int
}

func (frp *fixedRandProvider) RandN(n int) (int, error) {
	idx := frp.idx
	frp.idx += 1

	if idx >= len(frp.values) {
		return 0, fmt.Errorf("random values from fixed rand provider have been exhausted")
	}

	val := frp.values[idx]
	if val >= n {
		return 0, fmt.Errorf("fixed rand provider value %d greater than n=%d", val, n)
	}
	return val, nil
}
