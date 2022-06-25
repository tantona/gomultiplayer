package main

import (
	"context"
	"os"
	"tantona/gomultiplayer/clients/clients"
	"tantona/gomultiplayer/clients/gamestate"
	"tantona/gomultiplayer/clients/helpers"
	"tantona/gomultiplayer/clients/logging"
	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"
	"tantona/gomultiplayer/server/state"

	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"go.uber.org/zap"
)

var logger = logging.New("main")

var clientState = &ClientState{
	Id: uuid.Nil,
	GameState: &gamestate.GameState{
		Players: make(map[uuid.UUID]*state.PlayerData),
	},
}

func main() {
	f := faker.New()
	username := f.Internet().User()
	ctx := context.Background()
	client := clients.NewGRPCClient(":50005")

	go client.Run()

	for {
		select {

		case msg := <-client.GetMessageChan():
			logger.Debug("received msg", zap.Any("msg", msg))
			switch msg.Type {
			case multiplayer_v1.MessageType_SET_CLIENT_ID:
				clientId, err := uuid.Parse(msg.Data)
				if err != nil {
					logger.Debug("error", zap.Error(err))
				}

				clientState.SetClientId(clientId)
				client.SetClientId(clientId)

				msg, _ := helpers.CreateUpdatePlayerMsg(clientId, &state.PlayerData{Name: username, Score: 0})
				if err := client.Send(ctx, msg); err != nil {
					logger.Error("error sending message", zap.Error(err))
					return
				}

			case multiplayer_v1.MessageType_UPDATE_GAME_STATE:
				state, err := ToGameState(msg.Data)
				if err != nil {
					logger.Debug("error", zap.Error(err))
				}

				logger.Debug("updated game state:", zap.Any("state", state))
				clientState.SetGameState(state)

			}
		case <-client.GetSignalChan():
			logger.Debug("received interrupt signal")
			client.CloseConnection()
			os.Exit(0)
			return
		}

		clientState.Print()
	}
}
