package main

import (
	"encoding/json"
	"log"
	"time"

	"tantona/gomultiplayer/server/api"
	"tantona/gomultiplayer/server/state"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type GameState struct {
	Players map[uuid.UUID]*state.PlayerData
}

func ToGameState(data string) (*GameState, error) {

	gs := &GameState{}
	if err := json.Unmarshal([]byte(data), gs); err != nil {
		return nil, err
	}

	return gs, nil
}

type Client struct {
	url      string
	conn     *websocket.Conn
	done     chan struct{}
	Messages chan *api.Message
}

func (c *Client) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}
	c.conn = conn

	return nil
}

func (c *Client) Disconnect() error {
	return c.conn.Close()
}

func (c *Client) Listen() {
	defer close(c.done)
	for {

		msg := &api.Message{}
		if err := c.conn.ReadJSON(msg); err != nil {
			log.Println("read json err:", msg, err)
			return
		}

		c.Messages <- msg
	}
}

func (c *Client) Send(message *api.Message) error {
	b, err := json.Marshal(message)
	if err != nil {
		return err
	}

	logger.Debug("sent message", zap.Any("msg", message))
	return c.conn.WriteMessage(websocket.TextMessage, b)
}

func (c *Client) CloseConnection() {
	// Cleanly close the connection by sending a close message and then
	// waiting (with timeout) for the server to close the connection.
	err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}
	select {
	case <-c.done:
	case <-time.After(time.Second):
	}

}

func NewClient(url string, done chan struct{}) *Client {
	return &Client{
		url:      url,
		done:     make(chan struct{}),
		Messages: make(chan *api.Message),
	}
}
