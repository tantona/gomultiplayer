package state

import (
	"encoding/json"
	"sync"
	"tantona/gomultiplayer/server/api"
	"tantona/gomultiplayer/server/logging"
	"tantona/gomultiplayer/server/server"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var logger = logging.New("state")

type PlayerData struct {
	Name  string `json:"name"`
	Score int    `json:"score"`

	X     int    `json:"x"`
	Y     int    `json:"y"`
	Color string `json:"color"`
}

func messageToPlayerData(data string) (*PlayerData, error) {
	pd := &PlayerData{}
	if err := json.Unmarshal([]byte(data), pd); err != nil {
		return nil, err
	}

	return pd, nil
}

type GameState struct {
	mut     sync.Mutex                `json:"-"`
	Server  server.Server             `json:"-"`
	Players map[uuid.UUID]*PlayerData `json:"players"`
}

func (gs *GameState) UpdatePlayerData(id uuid.UUID, data *PlayerData) {
	gs.mut.Lock()
	defer gs.mut.Unlock()

	gs.Players[id] = data
}

func (gs *GameState) RemovePlayer(id uuid.UUID) {
	gs.mut.Lock()
	defer gs.mut.Unlock()

	delete(gs.Players, id)
}

func (gs *GameState) UpdatePlayerDataHandler(msg *api.Message) {
	data, err := messageToPlayerData(msg.Data)
	if err != nil {
		logger.Error("unable to parse game state")
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
	gs.Server.Broadcast(&api.Message{Type: api.UpdateGameState, Data: string(b)})
}

func (gs *GameState) MessageHandler(msg *api.Message) {
	logger.Info("received message", zap.Any("msg", msg))

	if msg.Type == api.UpdatePlayerData {
		gs.UpdatePlayerDataHandler(msg)
	}
}

func New(server server.Server) *GameState {
	return &GameState{
		Server:  server,
		Players: make(map[uuid.UUID]*PlayerData),
	}
}
