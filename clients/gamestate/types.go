package gamestate

import (
	"tantona/gomultiplayer/server/state"

	"github.com/google/uuid"
)

type GameState struct {
	Players map[uuid.UUID]*state.PlayerData
}
