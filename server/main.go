package main

import (
	"os"
	"os/signal"
	"tantona/gomultiplayer/server/logging"
	"tantona/gomultiplayer/server/server"
	"tantona/gomultiplayer/server/state"

	"go.uber.org/zap"
)

var logger = logging.New("main")

func main() {
	server := server.New(server.SERVER_GRPC)
	s := state.New(server)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go server.Run()

	logger.Info("Waiting for Clients")
	for {
		select {
		case msg := <-server.GetMessageChan():
			logger.Info("main thread: read message", zap.Any("msg", msg))
			s.MessageHandler(msg)
		case <-interrupt:
			logger.Info("quit!")
			os.Exit(0)
		}
	}
}
