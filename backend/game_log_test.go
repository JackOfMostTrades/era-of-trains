package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/JackOfMostTrades/eot/backend/common"
	"github.com/JackOfMostTrades/eot/backend/maps"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

type GameStep struct {
	Action               *ConfirmMoveRequest `json:"action"`
	ExpectedGameState    *common.GameState   `json:"expectedGameState"`
	ExpectedActivePlayer string              `json:"expectedActivePlayer"`
}

type GameLogTestCaseDefinition struct {
	MapName      string      `json:"mapName"`
	Steps        []*GameStep `json:"steps"`
	RandomValues []int       `json:"randomValues"`
}

func gameLogTestCase(definition string) func(t *testing.T) {
	return func(t *testing.T) {
		caseDefinition := new(GameLogTestCaseDefinition)
		defBytes, err := os.ReadFile("testlogs/" + definition + ".json")
		require.NoError(t, err)
		err = json.Unmarshal(defBytes, caseDefinition)
		require.NoError(t, err)

		gameMaps, err := maps.LoadMaps()
		require.NoError(t, err)
		gameMap := gameMaps[caseDefinition.MapName]
		gameState := caseDefinition.Steps[0].ExpectedGameState
		activePlayer := caseDefinition.Steps[0].ExpectedActivePlayer

		handler := &confirmMoveHandler{
			gameMap:      gameMap,
			gameState:    gameState,
			activePlayer: activePlayer,
			randProvider: &fixedRandProvider{values: caseDefinition.RandomValues},
			gameFinished: false,
		}

		for idx, step := range caseDefinition.Steps[1:] {
			assert.Equal(t, false, handler.gameFinished)

			err = handler.handleAction(step.Action)
			require.NoError(t, err)

			assert.Equal(t, step.ExpectedActivePlayer, handler.activePlayer, "Unexpected active player on step %d", idx+1)
			assert.Equal(t, step.ExpectedGameState, handler.gameState, "Unexpected game state on step %d", idx+1)
		}

		assert.Equal(t, true, handler.gameFinished)
	}
}

func TestGameLogCases(t *testing.T) {
	t.Run("beta test, 4p, rust belt", gameLogTestCase("5f3b56c3-ae82-495e-b153-39245820a5ad"))
	t.Run("beta test, 5p, rust belt", gameLogTestCase("62915bcc-04ed-48a2-a3c5-4b54782c5efb"))
}

func TestExportGameLogTestCase(t *testing.T) {
	gameId := ""
	if gameId == "" {
		t.SkipNow()
	}

	config := new(Config)
	f, err := os.Open("../deployment/config-prod.json")
	require.NoError(t, err)
	defer f.Close()
	err = json.NewDecoder(f).Decode(config)
	require.NoError(t, err)

	definition := &GameLogTestCaseDefinition{}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s",
		config.Database.MysqlUsername, config.Database.MysqlPassword,
		config.Database.MysqlHostname, config.Database.MysqlDatabase))
	require.NoError(t, err)

	stmt, err := db.Prepare("SELECT map_name FROM games WHERE id=?")
	require.NoError(t, err)
	defer stmt.Close()
	row := stmt.QueryRow(gameId)
	err = row.Scan(&definition.MapName)
	require.NoError(t, err)

	stmt, err = db.Prepare("SELECT action,new_active_player,new_game_state FROM game_log WHERE game_id=? ORDER BY timestamp ASC")
	require.NoError(t, err)
	defer stmt.Close()
	rows, err := stmt.Query(gameId)
	require.NoError(t, err)
	for rows.Next() {
		var actionStr sql.NullString
		var newActivePlayer sql.NullString
		var newGameStateStr sql.NullString
		err = rows.Scan(&actionStr, &newActivePlayer, &newGameStateStr)
		require.NoError(t, err)

		var action *ConfirmMoveRequest
		if actionStr.Valid {
			action = new(ConfirmMoveRequest)
			err = json.Unmarshal([]byte(actionStr.String), action)
			require.NoError(t, err)
		}
		var gameState *common.GameState
		if newGameStateStr.Valid {
			gameState = new(common.GameState)
			err = json.Unmarshal([]byte(newGameStateStr.String), gameState)
			require.NoError(t, err)
		}

		definition.Steps = append(definition.Steps, &GameStep{
			Action:               action,
			ExpectedGameState:    gameState,
			ExpectedActivePlayer: newActivePlayer.String,
		})
	}

	f, err = os.OpenFile("testlogs/"+gameId+".json", os.O_WRONLY|os.O_CREATE, 0644)
	require.NoError(t, err)
	defer f.Close()
	err = json.NewEncoder(f).Encode(definition)
	require.NoError(t, err)
}
