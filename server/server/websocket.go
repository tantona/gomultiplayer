package server

import (
	"encoding/json"
	"log"
	"net/http"

	"tantona/gomultiplayer/server/api"
	"tantona/gomultiplayer/server/logging"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var logger = logging.New("server")
var upgrader = websocket.Upgrader{} // use default options

type WebhookServer struct {
	MessageChan chan *api.Message
	ClientChan  chan *Client
	Clients     []*Client
}

func (s *WebhookServer) setupClientListener(c *Client) {
	c.isListening = true
	logger.Debug("listening for messages from client", zap.Stringer("clientId", c.Id))
	for {
		mt, b, err := c.conn.ReadMessage()
		if err != nil {
			if cE, ok := err.(*websocket.CloseError); ok {
				if err := s.removeClient(c.Id); err != nil {
					logger.Error("unable to remove client", zap.Stringer("id", c.Id))

				}

				if cE.Code != websocket.CloseNormalClosure {
					logger.Error("abnormal websocket closure", zap.Error(err))

				}
				return
			} else {
				logger.Warn("error reading message", zap.Int("messageType", mt), zap.Error(err))
			}

			continue
		}

		logger.Debug("recieved message",
			zap.Stringer("clientId", c.Id),
			zap.Int("type", mt),
			zap.ByteString("message", b),
		)

		switch mt {
		case websocket.TextMessage:
			msg := &api.Message{}
			if err := json.Unmarshal(b, msg); err != nil {
				logger.Error("unable to unmarshal message", zap.Stringer("Type", msg.Type))
				return
			}

			msg.ClientId = c.Id
			s.MessageChan <- msg

		default:
			logger.Error("handler for message type not implemented", zap.Any("websocket.MessageType", mt))
		}
	}

}

func (s *WebhookServer) listen() {
	for _, c := range s.Clients {
		if c.isListening {
			logger.Info("already listening to client", zap.Stringer("clientId", c.Id))
			continue
		}

		// TODO:
		go s.setupClientListener(c)
	}
}

func (s *WebhookServer) ListenForClients() {
	for {
		logger.Debug("waiting for clients")

		client := <-s.ClientChan
		s.Clients = append(s.Clients, client)
		logger.Info("clients changed! set up listeners", zap.Stringer("clientId", client.Id))

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

	logger.Debug("removed client", zap.String("id", id.String()))
	return nil
}

func (s *WebhookServer) Broadcast(message *api.Message) {
	for _, c := range s.Clients {
		if err := c.conn.WriteJSON(message); err != nil {
			logger.Error("error sending message", zap.Error(err))
		}
	}
}

func (s *WebhookServer) startHttpServer() {
	http.HandleFunc("/ws", s.addClientHandler)
	logger.Info("running websocket server on port :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Fatal("http server quit", zap.Error(err))
	}
}

func (s *WebhookServer) AddClient(conn *websocket.Conn) uuid.UUID {
	id := uuid.New()
	logger.Debug("ADDED CLIENT", zap.String("id", id.String()))
	s.ClientChan <- &Client{
		Id:   id,
		conn: conn,
	}

	return id
}

func (s *WebhookServer) addClientHandler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	clientId := s.AddClient(c)

	c.WriteJSON(&api.Message{Type: api.SetClientId, Data: clientId.String()})
}

func (s *WebhookServer) GetMessageChan() chan *api.Message {
	return s.MessageChan
}

func (s *WebhookServer) Run() {
	go s.startHttpServer()
	go s.ListenForClients()
}
