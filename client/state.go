package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"tantona/gomultiplayer/server/api"
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

func (c *ClientState) MessageHandler(msg *api.Message) {
	logger.Debug("received msg", zap.Any("msg", msg))

	switch msg.Type {

	case api.SetClientId:
		clientId, err := uuid.Parse(msg.Data)
		logger.Debug("set client id", zap.Stringer("clientId", clientId))
		if err != nil {
			logger.Debug("error", zap.Error(err))
		} else {
			logger.Debug("updated client id", zap.Stringer("clientId", clientId))
			c.SetClientId(clientId)
		}

	case api.UpdateGameState:
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
