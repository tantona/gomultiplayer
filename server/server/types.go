package server

import (
	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Server interface {
	Broadcast(message *multiplayer_v1.Message)
	AddClient(*websocket.Conn) uuid.UUID
	GetMessageChan() chan *multiplayer_v1.Message
	Run()
}

type Client struct {
	Id          uuid.UUID
	conn        *websocket.Conn
	isListening bool
}
