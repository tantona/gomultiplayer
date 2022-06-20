package server

import "tantona/gomultiplayer/server/api"

func New() Server {
	return &WebhookServer{
		MessageChan: make(chan *api.Message),
		ClientChan:  make(chan *Client),
	}
}
