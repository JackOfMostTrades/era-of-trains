package main

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWhoAmI(t *testing.T) {
	h := NewTestHarness(t)
	defer h.Close()

	player1 := h.createUser(t)
	res, err := h.whoami(t, player1, &WhoAmIRequest{})
	require.NoError(t, err)
	assert.Equal(t, player1, res.User.Id)
}
