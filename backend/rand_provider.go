package main

import (
	crand "crypto/rand"
	"fmt"
	"math/big"
)

type randProvider interface {
	RandN(n int) (int, error)
}

type cryptoRandProvider struct{}

func (*cryptoRandProvider) RandN(n int) (int, error) {
	n64 := big.NewInt(int64(n))
	val, err := crand.Int(crand.Reader, n64)
	if err != nil {
		return 0, err
	}
	if !val.IsInt64() {
		return 0, fmt.Errorf("unable to generate random number")
	}
	return int(val.Int64()), nil
}
