package state

import (
	"encoding/json"
	"sync"
	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"
	"tantona/gomultiplayer/server/logging"
	"tantona/gomultiplayer/server/server"

	"go.uber.org/zap"
)

var logger = logging.New("state")

type PlayerData struct {
	Name  string `json:"name"`
	Score int    `json:"score"`

	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Color string  `json:"color"`
}

func messageToPlayerData(data string) (*PlayerData, error) {
	pd := &PlayerData{}
	if err := json.Unmarshal([]byte(data), pd); err != nil {
		return nil, err
	}

	return pd, nil
}

type GameState struct {
	mut     sync.Mutex             `json:"-"`
	Server  server.Server          `json:"-"`
	Players map[string]*PlayerData `json:"players"`
}

func (gs *GameState) UpdatePlayerData(id string, data *PlayerData) {
	gs.mut.Lock()
	defer gs.mut.Unlock()

	gs.Players[id] = data
}

func (gs *GameState) RemovePlayer(id string) {
	gs.mut.Lock()
	defer gs.mut.Unlock()

	delete(gs.Players, id)
}

func (gs *GameState) UpdatePlayerDataHandler(msg *multiplayer_v1.Message) {
	logger.Debug("GameState.UpdatePlayerDataHandler", zap.Any("msg", msg))
	data, err := messageToPlayerData(msg.Data)
	if err != nil {
		logger.Error("unable to parse game state", zap.String("msg.Data", msg.Data))
		return
	}
	logger.Debug("update player data", zap.Any("player", data))

	gs.UpdatePlayerData(msg.ClientId, data)

	gs.Broadcast()
}

func (gs *GameState) Broadcast() {
	b, err := json.Marshal(gs)
	if err != nil {
		logger.Error("unable to marshal game state", zap.Error(err))
		return
	}
	logger.Info("broadcast game state", zap.String("state", string(b)))
	gs.Server.BroadcastBinary(&multiplayer_v1.Message{Type: multiplayer_v1.MessageType_UPDATE_GAME_STATE, Data: string(b)})
}

func (gs *GameState) MessageHandler(msg *multiplayer_v1.Message) {
	logger.Info("GameState.MessageHandler", zap.Any("msg", msg))

	if msg.Type == multiplayer_v1.MessageType_UPDATE_PLAYER_DATA {
		gs.UpdatePlayerDataHandler(msg)
	}
}

func New(server server.Server) *GameState {
	return &GameState{
		Server:  server,
		Players: make(map[string]*PlayerData),
	}
}
