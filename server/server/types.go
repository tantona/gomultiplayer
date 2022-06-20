package server

import (
	"tantona/gomultiplayer/server/api"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Server interface {
	Broadcast(message *api.Message)
	AddClient(*websocket.Conn) uuid.UUID
	GetMessageChan() chan *api.Message
	Run()
}

type Client struct {
	Id          uuid.UUID
	conn        *websocket.Conn
	isListening bool
}
