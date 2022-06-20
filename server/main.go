package main

import (
	"os"
	"os/signal"
	"tantona/gomultiplayer/server/logging"
	"tantona/gomultiplayer/server/server"
	"tantona/gomultiplayer/server/state"
)

var logger = logging.New("main")

func main() {
	server := server.New()
	s := state.New(server)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	go server.Run()

	for {
		select {
		case msg := <-server.GetMessageChan():
			s.MessageHandler(msg)
		case <-interrupt:
			logger.Info("quit!")
			os.Exit(0)
		}
	}
}
