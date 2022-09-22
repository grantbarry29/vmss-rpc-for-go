package rpc

import (
	"context"
	"log"
	"net"

	pb "vmss-rpc-for-go/rpc/stubs"

	"google.golang.org/grpc"
)

const (
	serverPort        = "50051"
	serverProtocol    = "tcp"
	StatusCodeSuccess = 200
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func Server() {
	listen, err := net.Listen(serverProtocol, ":"+serverPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", listen.Addr())
	if err := s.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
