package clients

import (
	"context"
	"os"
	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"

	"github.com/google/uuid"
)

type Client interface {
	GetMessageChan() chan *multiplayer_v1.Message
	GetSignalChan() chan os.Signal
	Run()
	Send(context.Context, *multiplayer_v1.Message) error
	CloseConnection()
	GetClientId() uuid.UUID
	SetClientId(id uuid.UUID)
}
