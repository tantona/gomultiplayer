package api

import (
	"github.com/google/uuid"
)

type MessageType string

func (mt MessageType) String() string {
	return string(mt)
}

const (
	UpdatePlayerData MessageType = "UPDATE_PLAYER_DATA"
	UpdateGameState  MessageType = "UPDATE_GAME_STATE"
	SetClientId      MessageType = "SET_CLIENT_ID"
	DisconnectClient MessageType = "DISCONNECT_CLIENT"
	ClientAdded      MessageType = "CLIENT_ADDED"
)

type Message struct {
	Type     MessageType `json:"type"`
	Data     string      `json:"data"`
	ClientId uuid.UUID
}
