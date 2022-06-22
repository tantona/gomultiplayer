package main

import (
	"encoding/json"
	"fmt"
	"sync"
	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"
	"tantona/gomultiplayer/server/state"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ClientState struct {
	mut       sync.Mutex
	Id        uuid.UUID
	GameState *GameState
}

func (c *ClientState) Print() {
	b, _ := json.MarshalIndent(c, "", "  ")
	fmt.Println(string(b))
}

func (c *ClientState) GetPlayerData() *state.PlayerData {
	logger.Info("GetPlayerData", zap.Any("d", c.GameState))
	if c.GameState == nil {
		return nil
	}
	return c.GameState.Players[c.Id]
}

func (c *ClientState) SetClientId(id uuid.UUID) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.Id = id
}

func (c *ClientState) SetGameState(gs *GameState) {
	c.mut.Lock()
	defer c.mut.Unlock()
	for id, player := range gs.Players {
		c.GameState.Players[id] = player
	}

}

func (c *ClientState) MessageHandler(msg *multiplayer_v1.Message) {
	logger.Debug("received msg", zap.Any("msg", msg))

	switch msg.Type {

	case multiplayer_v1.MessageType_SET_CLIENT_ID:
		clientId, err := uuid.Parse(msg.Data)
		logger.Debug("set client id", zap.Stringer("clientId", clientId))
		if err != nil {
			logger.Debug("error", zap.Error(err))
		} else {
			logger.Debug("updated client id", zap.Stringer("clientId", clientId))
			c.SetClientId(clientId)
		}

	case multiplayer_v1.MessageType_UPDATE_GAME_STATE:
		state, err := ToGameState(msg.Data)
		if err != nil {
			logger.Debug("error", zap.Error(err))
		} else {
			logger.Debug("updated game state:", zap.Any("state", state))
			c.SetGameState(state)
		}

	default:
		logger.Debug("api message type not implemented")
	}
}
