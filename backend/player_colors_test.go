package main

import (
	"database/sql"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssignPlayerColors(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	bootstrapSql, err := os.ReadFile("bootstrap.sql")
	require.NoError(t, err)
	_, err = db.Exec(string(bootstrapSql))
	require.NoError(t, err)

	_, err = db.Exec("INSERT INTO users (id, color_preferences) VALUES ('player1','[4,3,2]'), ('player2','[4,2,3]'), ('player3','[5]'), ('player4',NULL), ('player5',NULL)")
	require.NoError(t, err)

	playerColors, err := assignPlayerColors(db, []string{"player1", "player2", "player3", "player4"})
	require.NoError(t, err)
	assert.Equal(t, 4, len(playerColors))
	if playerColors["player1"] == 4 {
		assert.Equal(t, 4, playerColors["player1"])
		assert.Equal(t, 2, playerColors["player2"])
		assert.Equal(t, 5, playerColors["player3"])
	} else {
		assert.Equal(t, 3, playerColors["player1"])
		assert.Equal(t, 4, playerColors["player2"])
		assert.Equal(t, 5, playerColors["player3"])
	}

	// Assert player4's color is in the valid range and not equal to any other player's color
	assert.GreaterOrEqual(t, playerColors["player4"], 0)
	assert.LessOrEqual(t, playerColors["player4"], 5)
	for _, playerId := range []string{"player1", "player2", "player3"} {
		assert.NotEqual(t, playerColors[playerId], playerColors["player4"])
	}
}
