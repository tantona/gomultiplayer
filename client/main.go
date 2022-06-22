package main

import (
	"encoding/json"
	"os"
	"os/signal"
	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"
	"tantona/gomultiplayer/server/logging"
	"tantona/gomultiplayer/server/state"

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

func createUpdatePlayerMsg(data *state.PlayerData) (*multiplayer_v1.Message, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &multiplayer_v1.Message{Type: multiplayer_v1.MessageType_UPDATE_PLAYER_DATA, Data: string(b)}, nil
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
			case multiplayer_v1.MessageType_SET_CLIENT_ID:
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

			case multiplayer_v1.MessageType_UPDATE_GAME_STATE:
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
