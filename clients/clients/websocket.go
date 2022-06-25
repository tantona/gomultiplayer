package clients

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"tantona/gomultiplayer/clients/helpers"
	"tantona/gomultiplayer/clients/logging"
	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"
	"tantona/gomultiplayer/server/state"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var logger = logging.New("client/websocket")

type WebsocketClient struct {
	id         uuid.UUID
	url        string
	conn       *websocket.Conn
	done       chan struct{}
	signalChan chan os.Signal
	Messages   chan *multiplayer_v1.Message
}

func (c *WebsocketClient) GetClientId() uuid.UUID {
	return c.id
}

func (c *WebsocketClient) SetClientId(id uuid.UUID) {
	c.id = id
}

func (c *WebsocketClient) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}
	c.conn = conn

	return nil
}

func (c *WebsocketClient) Disconnect() error {
	return c.conn.Close()
}

func (c *WebsocketClient) Listen() {
	defer close(c.done)
	for {

		msg := &multiplayer_v1.Message{}
		if err := c.conn.ReadJSON(msg); err != nil {
			log.Println("read json err:", msg, err)
			return
		}

		c.Messages <- msg
	}
}

func (c *WebsocketClient) Send(ctx context.Context, msg *multiplayer_v1.Message) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	logger.Debug("sent message", zap.Any("msg", msg))
	return c.conn.WriteMessage(websocket.TextMessage, b)
}

func (c *WebsocketClient) CloseConnection() {
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

func (c *WebsocketClient) initKeyboard() {
	ctx := context.Background()
	keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		// myData := clientState.GetPlayerData()
		// logger.Debug("GetPlayerData", zap.Any("myData", myData))
		switch key.Code {
		case keys.CtrlC, keys.Escape:
			c.signalChan <- os.Interrupt

			// return true, nil // Return true to stop listener
		case keys.Up:
			msg, _ := helpers.CreateUpdatePlayerMsg(c.GetClientId(), &state.PlayerData{Name: "foobar", Score: 10})
			logger.Debug("increase score", zap.Any("msg", msg))
			c.Send(ctx, msg)
		case keys.Down:
			msg, _ := helpers.CreateUpdatePlayerMsg(c.GetClientId(), &state.PlayerData{Name: "foobar", Score: 10})
			logger.Debug("decrease score", zap.Any("msg", msg))
			c.Send(ctx, msg)
		}

		return false, nil // Return false to continue listening
	})
}

func (c *WebsocketClient) Run() {
	if err := c.Connect(); err != nil {
		logger.Fatal("unable to connect to websocket host", zap.Error(err))
	}
	defer c.Disconnect()

	go c.Listen()
	go c.initKeyboard()
}

func (c *WebsocketClient) GetMessageChan() chan *multiplayer_v1.Message {
	return c.Messages
}

func (c *WebsocketClient) GetSignalChan() chan os.Signal {
	return c.signalChan
}

func NewWebsocketClient(url string) Client {
	return &WebsocketClient{
		url:        url,
		done:       make(chan struct{}),
		Messages:   make(chan *multiplayer_v1.Message),
		signalChan: make(chan os.Signal, 1),
	}
}
