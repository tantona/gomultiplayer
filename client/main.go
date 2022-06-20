package main

import (
	"encoding/json"
	"tantona/gomultiplayer/server/api"
	"tantona/gomultiplayer/server/logging"
	"tantona/gomultiplayer/server/state"

	"os"
	"os/signal"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"go.uber.org/zap"
)

var logger = logging.New("main")

var clientState = &ClientState{
	Id: uuid.Nil,
	GameState: &GameState{
		Players: make(map[uuid.UUID]*state.PlayerData),
	},
}

func createUpdatePlayerMsg(data *state.PlayerData) (*api.Message, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &api.Message{Type: api.UpdatePlayerData, Data: string(b)}, nil
}

func main() {
	fake := faker.New()
	username := fake.Internet().User()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	done := make(chan struct{})

	client := NewClient("ws://localhost:8080/ws", done)

	if err := client.Connect(); err != nil {
		logger.Fatal("unable to connect to websocket host", zap.Error(err))
	}
	defer client.Disconnect()

	go client.Listen()

	go initKeyboard(client, signalChan)

	for {
		select {
		case <-done:
			return
		case msg := <-client.Messages:
			logger.Debug("received msg", zap.Any("msg", msg))
			switch msg.Type {
			case api.SetClientId:
				clientId, err := uuid.Parse(msg.Data)
				logger.Debug("set client id", zap.Stringer("clientId", clientId))
				if err != nil {
					logger.Debug("error", zap.Error(err))
				} else {
					logger.Debug("updated client id", zap.Stringer("clientId", clientId))
					clientState.SetClientId(clientId)
				}
				msg, _ := createUpdatePlayerMsg(&state.PlayerData{Name: username, Score: 0})
				if err := client.Send(msg); err != nil {
					logger.Error("error sending message", zap.Error(err))
					return
				}

			case api.UpdateGameState:
				state, err := ToGameState(msg.Data)
				if err != nil {
					logger.Debug("error", zap.Error(err))
				} else {
					logger.Debug("updated game state:", zap.Any("state", state))
					clientState.SetGameState(state)
				}
				clientState.Print()
			}
		case <-signalChan:
			logger.Debug("received interrupt signal")
			client.CloseConnection()
			os.Exit(0)
			return
		}
	}
}
