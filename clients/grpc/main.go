package grpc

import (
	"context"
	"io"
	"os"
	"os/signal"
	"tantona/gomultiplayer/clients/helpers"
	"tantona/gomultiplayer/clients/logging"
	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"
	"tantona/gomultiplayer/server/state"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/jaswdr/faker"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var logger = logging.New("grpcClient")

type Client struct {
	client      multiplayer_v1.MessageServiceClient
	signalChan  chan os.Signal
	messageChan chan *multiplayer_v1.Message
}

func (c *Client) RegisterKeypressHandlers(ctx context.Context) {
	f := faker.New()

	keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.CtrlC, keys.Escape:
			c.signalChan <- os.Interrupt

		case keys.Up:
			msg, _ := helpers.CreateUpdatePlayerMsg(&state.PlayerData{
				Name:  f.Company().Name(),
				Score: 10,
				X:     0,
				Y:     0,
			})

			logger.Debug("sent message", zap.Any("msg", msg))
			if _, err := c.client.SendMessage(ctx, &multiplayer_v1.SendMessageRequest{
				Message: msg,
			}); err != nil {
				logger.Error("unable to send message", zap.Error(err))
			}

		}

		return false, nil // Return false to continue listening
	})
}
func (c *Client) Listen(done chan bool) {
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
			logger.Info("GOT MESSAGE FROM STREAM")
			if err == io.EOF {
				done <- true //means stream is finished
				return
			}
			if err != nil {
				logger.Fatal("cannot receive", zap.Error(err))
			}

			c.messageChan <- resp
		}
	}()
}

func (c *Client) Run() {
	ctx := context.Background()
	done := make(chan bool)
	go c.RegisterKeypressHandlers(ctx)
	c.Listen(done)

	for {
		select {
		case msg := <-c.messageChan:
			logger.Info("received message", zap.Any("msg", msg))
		case <-done:
			//we will wait until all response is received
			logger.Info("finished")
			os.Exit(0)
		case <-c.signalChan:
			logger.Debug("received interrupt signal")
			// client.CloseConnection()
			os.Exit(0)
			return
		}
	}
}

func Run() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	conn, err := grpc.Dial(":50005", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("can not connect to server", zap.Error(err))
	}

	client := multiplayer_v1.NewMessageServiceClient(conn)

	c := &Client{
		client:     client,
		signalChan: signalChan,
	}

	c.Run()
}
