package server

import (
	"log"
	"net/http"

	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"
	"tantona/gomultiplayer/server/logging"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

var wsLogger = logging.New("wsLogger")

type WebhookServer struct {
	upgrader    websocket.Upgrader
	MessageChan chan *multiplayer_v1.Message
	ClientChan  chan *Client
	Clients     []*Client
}

func (s *WebhookServer) setupClientListener(c *Client) {

	c.isListening = true
	wsLogger.Debug("listening for messages from client", zap.Stringer("clientId", c.Id))
	for {
		mt, b, err := c.conn.ReadMessage()
		if err != nil {
			if cE, ok := err.(*websocket.CloseError); ok {
				if err := s.removeClient(c.Id); err != nil {
					wsLogger.Error("unable to remove client", zap.Stringer("id", c.Id))

				}

				if cE.Code != websocket.CloseNormalClosure {
					wsLogger.Error("abnormal websocket closure", zap.Error(err))

				}
				return
			} else {
				wsLogger.Warn("error reading message", zap.Int("messageType", mt), zap.Error(err))
			}

			continue
		}

		wsLogger.Debug("recieved message",
			zap.Stringer("clientId", c.Id),
			zap.Int("type", mt),
			zap.ByteString("message", b),
		)

		switch mt {
		case websocket.TextMessage:
			wsLogger.Debug("UNMARSHAL TEXT", zap.ByteString("msg", b))
			msg := &multiplayer_v1.Message{}
			if err := protojson.Unmarshal(b, msg); err != nil {
				wsLogger.Error("unable to unmarshal text message", zap.Stringer("Type", msg.Type))
			}

			msg.ClientId = c.Id.String()
			s.MessageChan <- msg
		case websocket.BinaryMessage:
			msg := &multiplayer_v1.Message{}
			if err := protojson.Unmarshal(b, msg); err != nil {
				wsLogger.Error("unable to unmarshal binary message", zap.Stringer("Type", msg.Type))
			}

			msg.ClientId = c.Id.String()
			s.MessageChan <- msg

		default:
			wsLogger.Error("handler for message type not implemented", zap.Any("websocket.MessageType", mt))
		}
	}

}

func (s *WebhookServer) listen() {
	for _, c := range s.Clients {
		if c.isListening {
			wsLogger.Info("already listening to client", zap.Stringer("clientId", c.Id))
			continue
		}

		// TODO:
		go s.setupClientListener(c)
	}
}

func (s *WebhookServer) ListenForClients() {
	for {
		wsLogger.Debug("waiting for clients")

		client := <-s.ClientChan
		s.Clients = append(s.Clients, client)
		wsLogger.Info("clients changed! set up listeners", zap.Stringer("clientId", client.Id))

		s.listen()
	}
}

func (s *WebhookServer) removeClient(id uuid.UUID) error {
	clients := []*Client{}
	for _, c := range s.Clients {
		if c.Id != id {
			clients = append(clients, c)
		} else {
			c.conn.Close()
		}
	}

	s.Clients = clients

	wsLogger.Debug("removed client", zap.String("id", id.String()))
	return nil
}

func (s *WebhookServer) Broadcast(message *multiplayer_v1.Message) {
	for _, c := range s.Clients {
		if err := c.conn.WriteJSON(message); err != nil {
			wsLogger.Error("error sending message", zap.Error(err))
		}
	}
}

func (s *WebhookServer) BroadcastBinary(message *multiplayer_v1.Message) {
	b, err := protojson.Marshal(message)
	if err != nil {
		wsLogger.Error("error marshalling message", zap.Error(err))
	}
	for _, c := range s.Clients {
		if err := c.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
			wsLogger.Error("error sending message", zap.Error(err))
		}
	}
}

func (s *WebhookServer) startHttpServer() {

	http.HandleFunc("/ws", s.addClientHandler)
	http.HandleFunc("/wsb", s.addClientHandlerBinary)

	wsLogger.Info("running websocket server on port :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		wsLogger.Fatal("http server quit", zap.Error(err))
	}
}

func (s *WebhookServer) addClient(conn *websocket.Conn) uuid.UUID {
	id := uuid.New()
	wsLogger.Debug("ADDED CLIENT", zap.String("id", id.String()))
	s.ClientChan <- &Client{
		Id:   id,
		conn: conn,
	}

	return id
}

func (s *WebhookServer) addClientHandler(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	clientId := s.addClient(c)

	c.WriteJSON(&multiplayer_v1.Message{Type: multiplayer_v1.MessageType_SET_CLIENT_ID, Data: clientId.String()})
}

func (s *WebhookServer) addClientHandlerBinary(w http.ResponseWriter, r *http.Request) {
	c, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wsLogger.Error("upgrade:", zap.Error(err))
		return
	}

	clientId := s.addClient(c)
	msg := &multiplayer_v1.Message{Type: multiplayer_v1.MessageType_SET_CLIENT_ID, Data: clientId.String()}
	b, err := protojson.Marshal(msg)
	if err != nil {
		wsLogger.Error("unable to marshal message:", zap.Error(err))
		return
	}
	c.WriteMessage(websocket.BinaryMessage, b)
}

func (s *WebhookServer) GetMessageChan() chan *multiplayer_v1.Message {
	return s.MessageChan
}

func (s *WebhookServer) Run() {
	go s.startHttpServer()
	go s.ListenForClients()
}

func newWebsocketServer() *WebhookServer {
	return &WebhookServer{
		upgrader:    websocket.Upgrader{}, // use default options
		MessageChan: make(chan *multiplayer_v1.Message),
		ClientChan:  make(chan *Client),
	}
}
