package rpc

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "vmss-rpc-for-go/rpc/stubs"

	"google.golang.org/grpc"
)

const (
	serverPort     = "50051"
	serverProtocol = "tcp"
)

type RPCServer struct {
}

func NewRPCServer() *RPCServer {
	server := RPCServer{}
	return &server
}

func (rs *RPCServer) ListenAndRegister() {
	go func() {
		// Listen for RPC connections
		listen, err := net.Listen(serverProtocol, ":"+serverPort)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		// Register RPC server
		s := grpc.NewServer()
		pb.RegisterGreeterServer(s, &server{})
		log.Printf("server listening at %v", listen.Addr())
		if err := s.Serve(listen); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	fmt.Print("Received RPC from ID: ", in.GetName(), "\n")
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}
