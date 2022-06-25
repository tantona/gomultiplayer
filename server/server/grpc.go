package server

import (
	"net"
	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"
	"tantona/gomultiplayer/server/logging"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var grpcLogger = logging.New("grpcLogger")

type GRPCServer struct {
	MessageChan chan *multiplayer_v1.Message
	out         chan *multiplayer_v1.Message
}

func (s *GRPCServer) Broadcast(message *multiplayer_v1.Message) {
	grpcLogger.Info("GRPCServer.Broadcast", zap.Any("req", message))
	// s.MessageChan <- message
}

func (s *GRPCServer) BroadcastBinary(message *multiplayer_v1.Message) {
	grpcLogger.Info("GRPCServer.BroadcastBinary", zap.Any("req", message))
	s.out <- message
}

func (s *GRPCServer) AddClient(*websocket.Conn) uuid.UUID {
	return uuid.UUID{}
}

func (s *GRPCServer) GetMessageChan() chan *multiplayer_v1.Message {
	return s.MessageChan
}

func (s *GRPCServer) startGrpcServer() {
	lis, err := net.Listen("tcp", ":50005")
	if err != nil {
		grpcLogger.Fatal("grpc server failed to listen", zap.Error(err))
	}

	grpcServer := grpc.NewServer()
	multiplayer_v1.RegisterMessageServiceServer(grpcServer, s)

	grpcLogger.Info("start grpc server")
	if err := grpcServer.Serve(lis); err != nil {
		grpcLogger.Fatal("grpc server failed to serve", zap.Error(err))
	}
}

func (s *GRPCServer) SendMessage(req *multiplayer_v1.SendMessageRequest, srv multiplayer_v1.MessageService_SendMessageServer) error {
	grpcLogger.Info("GRPCServer.SendMessage", zap.Any("req", req))
	s.MessageChan <- req.Message
	return srv.Send(&multiplayer_v1.SendMessageResponse{})
}

func (s *GRPCServer) GetMessageStream(req *multiplayer_v1.GetMessageStreamRequest, srv multiplayer_v1.MessageService_GetMessageStreamServer) error {
	grpcLogger.Info("GRPCServer.GetMessageStream", zap.Any("req", req))

	for {
		select {
		case msg := <-s.out:
			grpcLogger.Info("GRPCServer.GetMessageStream.read", zap.Any("msg", msg))
			if err := srv.Send(msg); err != nil {
				grpcLogger.Error("unable to send message", zap.Error(err))
				return err
			}
		}
	}

}

func (s *GRPCServer) Run() {
	go s.startGrpcServer()
}

func newGRPCServer() *GRPCServer {
	return &GRPCServer{
		MessageChan: make(chan *multiplayer_v1.Message),
		out:         make(chan *multiplayer_v1.Message),
	}
}
