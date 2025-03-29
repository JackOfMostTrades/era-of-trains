package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSpoofSession(t *testing.T) {
	t.SkipNow()
	server := &GameServer{
		config: &Config{
			MacKey: "mac_key_here",
		},
	}
	session, err := server.createSession(&Session{
		UserId: "user_id_here",
	})
	require.NoError(t, err)
	t.Log(session)
}
