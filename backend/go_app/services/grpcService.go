package services

import (
	"context"
	"net"

	"chatbot/logger"
	pb "chatbot/utils/proto"

	"google.golang.org/grpc"
)

const (
	port = ":50052"
)

type server struct {
	pb.UnimplementedWhatsAppServiceServer
}

func (s *server) CreateThread(ctx context.Context, in *pb.CreateThreadRequest) (*pb.CreateThreadResponse, error) {
	// Implementación de CreateThread
	return &pb.CreateThreadResponse{ThreadId: "nuevo_thread_id"}, nil
}

func (s *server) CreateThreadAnalizer(ctx context.Context, in *pb.CreateThreadAnalizerRequest) (*pb.CreateThreadAnalizerResponse, error) {
	// Implementación de CreateThreadAnalizer
	return &pb.CreateThreadAnalizerResponse{ThreadIdAnalizer: "nuevo_thread_analizer_id"}, nil
}

func (s *server) GenerateResponse(ctx context.Context, in *pb.GenerateResponseRequest) (*pb.GenerateResponseResponse, error) {
	// Implementación de GenerateResponse
	return &pb.GenerateResponseResponse{Response: "respuesta_generada"}, nil
}

func (s *server) GenerateResponseAnalizer(ctx context.Context, in *pb.GenerateResponseAnalizerRequest) (*pb.GenerateResponseAnalizerResponse, error) {
	// Implementación de GenerateResponseAnalizer
	return &pb.GenerateResponseAnalizerResponse{Response: "respuesta_analizer_generada"}, nil
}

func StartGRPCServer() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterWhatsAppServiceServer(s, &server{})
	logger.Log.Infof("gRPC server is running on port %v", port)
	if err := s.Serve(lis); err != nil {
		logger.Log.Fatalf("failed to serve: %v", err)
	}
}
