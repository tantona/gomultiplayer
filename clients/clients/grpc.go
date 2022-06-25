package clients

import (
	"context"
	"io"
	"os"
	"tantona/gomultiplayer/clients/helpers"
	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"
	"tantona/gomultiplayer/server/state"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	id          uuid.UUID
	client      multiplayer_v1.MessageServiceClient
	signalChan  chan os.Signal
	messageChan chan *multiplayer_v1.Message
}

func (c *GRPCClient) GetClientId() uuid.UUID {
	return c.id
}

func (c *GRPCClient) SetClientId(id uuid.UUID) {
	c.id = id
}

func (c *GRPCClient) RegisterKeypressHandlers(ctx context.Context) {
	f := faker.New()
	username := f.Internet().User()
	score := 0
	keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.CtrlC, keys.Escape:
			c.signalChan <- os.Interrupt

		case keys.Up:
			score = score + 10
		case keys.Down:
			score = score - 10

		}

		msg, _ := helpers.CreateUpdatePlayerMsg(
			c.GetClientId(),
			&state.PlayerData{
				Name:  username,
				Score: score,
				X:     0,
				Y:     0,
			})

		logger.Debug("sent message", zap.Any("msg", msg))
		if _, err := c.client.SendMessage(ctx, &multiplayer_v1.SendMessageRequest{
			Message: msg,
		}); err != nil {
			logger.Error("unable to send message", zap.Error(err))
		}

		return false, nil // Return false to continue listening
	})
}
func (c *GRPCClient) Listen() {
	req := &multiplayer_v1.GetMessageStreamRequest{}
	ctx := context.Background()
	stream, err := c.client.GetMessageStream(ctx, req)
	if err != nil {
		logger.Fatal("open stream error", zap.Error(err))
	}

	logger.Info("LISTENING")
	go func() {
		for {
			resp, err := stream.Recv() // replace with RecvMsg on a stream?
			logger.Debug("read msg from stream", zap.Any("resp", resp))
			if err == io.EOF {
				logger.Error("end of stream")
				// done <- true //means stream is finished
				break
			}
			if err != nil {
				logger.Error("cannot receive", zap.Error(err))
				break
			}

			c.messageChan <- resp

		}
	}()
}

func (c *GRPCClient) Run() {
	ctx := context.Background()

	go c.RegisterKeypressHandlers(ctx)
	c.Listen()

}

func (c *GRPCClient) Send(ctx context.Context, msg *multiplayer_v1.Message) error {
	logger.Debug("Sending message", zap.Any("msg", msg))

	if _, err := c.client.SendMessage(ctx, &multiplayer_v1.SendMessageRequest{
		ClientId: c.GetClientId().String(),
		Message:  msg}); err != nil {
		return err
	}

	return nil
}

func (c *GRPCClient) GetMessageChan() chan *multiplayer_v1.Message {
	return c.messageChan
}

func (c *GRPCClient) GetSignalChan() chan os.Signal {
	return c.signalChan
}

func (c *GRPCClient) CloseConnection() {}

func NewGRPCClient(url string) Client {
	conn, err := grpc.Dial(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("can not connect to server", zap.Error(err))
	}

	client := multiplayer_v1.NewMessageServiceClient(conn)
	return &GRPCClient{
		client:      client,
		signalChan:  make(chan os.Signal),
		messageChan: make(chan *multiplayer_v1.Message),
	}
}
