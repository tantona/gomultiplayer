package server

type ServerType = int64

const (
	SERVER_GRPC = iota
	SERVER_WEBSOCKET
)

func New(st ServerType) Server {
	switch st {
	case SERVER_GRPC:
		return newGRPCServer()
	case SERVER_WEBSOCKET:
		return newWebsocketServer()
	}

	return nil
}
