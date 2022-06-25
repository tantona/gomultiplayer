package server

import (
	"net"
	"sync"
	multiplayer_v1 "tantona/gomultiplayer/gen/proto/go/multiplayer/v1"
	"tantona/gomultiplayer/server/logging"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var grpcLogger = logging.New("grpcLogger")

type GRPCServer struct {
	MessageChan   chan *multiplayer_v1.Message
	mut           sync.RWMutex
	streams       map[uuid.UUID]multiplayer_v1.MessageService_GetMessageStreamServer
	endStreamChan chan uuid.UUID
}

func (s *GRPCServer) Broadcast(message *multiplayer_v1.Message) {
	grpcLogger.Info("GRPCServer.Broadcast", zap.Any("req", message))

	for clientID, stream := range s.streams {
		if err := stream.Send(message); err != nil {
			grpcLogger.Error("Broadcast: unable to send message to client", zap.Stringer("clientId", clientID), zap.Any("msg", message))
		}
	}
}

func (s *GRPCServer) BroadcastBinary(message *multiplayer_v1.Message) {
	grpcLogger.Info("GRPCServer.BroadcastBinary", zap.Any("req", message))
	for clientId, stream := range s.streams {
		if err := stream.Send(message); err != nil {
			grpcLogger.Error("BroadcastBinary: unable to send message to client", zap.Stringer("clientId", clientId), zap.Any("msg", message))
		}
		grpcLogger.Info("GRPCServer.BroadcastBinary: sent message to client", zap.Stringer("clientId", clientId))
	}
}

func (s *GRPCServer) addClient(stream multiplayer_v1.MessageService_GetMessageStreamServer) error {
	s.mut.Lock()
	defer s.mut.Unlock()
	clientId := uuid.New()
	s.streams[clientId] = stream
	wsLogger.Debug("ADDED CLIENT", zap.Stringer("id", clientId))

	msg := &multiplayer_v1.Message{Type: multiplayer_v1.MessageType_SET_CLIENT_ID, Data: clientId.String()}
	if err := stream.Send(msg); err != nil {
		grpcLogger.Error("unable to send message", zap.Error(err))
	}

	return nil
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

func (s *GRPCServer) GetMessageStream(req *multiplayer_v1.GetMessageStreamRequest, stream multiplayer_v1.MessageService_GetMessageStreamServer) error {
	grpcLogger.Info("GRPCServer.GetMessageStream", zap.Any("req", req))

	if err := s.addClient(stream); err != nil {
		grpcLogger.Error("unable to add client", zap.Error(err))
		return err
	}

	for {
		select {
		case clientId := <-s.endStreamChan:
			grpcLogger.Info("GRPCServer.endStreamChan.read", zap.Stringer("clientId", clientId))
			return nil
			// if err := stream.Send(msg); err != nil {
			// 	grpcLogger.Error("unable to send message", zap.Error(err))
			// 	return err
			// }
		}
	}

}

func (s *GRPCServer) Run() {
	go s.startGrpcServer()
}

func newGRPCServer() *GRPCServer {
	return &GRPCServer{
		MessageChan:   make(chan *multiplayer_v1.Message),
		streams:       make(map[uuid.UUID]multiplayer_v1.MessageService_GetMessageStreamServer),
		endStreamChan: make(chan uuid.UUID),
	}
}
