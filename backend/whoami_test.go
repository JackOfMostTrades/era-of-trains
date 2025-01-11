package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWhoAmI(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	res := h.whoami(t, player1, &WhoAmIRequest{})
	assert.Equal(t, player1, res.User.Id)
}
