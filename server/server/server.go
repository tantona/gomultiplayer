package server

import multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"

func New() Server {
	return &WebhookServer{
		MessageChan: make(chan *multiplayer_v1.Message),
		ClientChan:  make(chan *Client),
	}
}
