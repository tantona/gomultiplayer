package helpers

import (
	"encoding/json"
	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"
	"tantona/gomultiplayer/server/state"

	"github.com/google/uuid"
)

func CreateUpdatePlayerMsg(clientId uuid.UUID, data *state.PlayerData) (*multiplayer_v1.Message, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &multiplayer_v1.Message{
		Type:     multiplayer_v1.MessageType_UPDATE_PLAYER_DATA,
		Data:     string(b),
		ClientId: clientId.String(),
	}, nil
}
