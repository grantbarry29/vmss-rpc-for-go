package rpc

import (
	"context"
	"fmt"
	"time"

	pb "vmss-rpc-for-go/rpc/stubs"

	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type RPCClient struct {
}

func NewRPCClient() *RPCClient {
	return &RPCClient{}
}

func (*RPCClient) CallGetSendServerName(ip string, name string) error {
	// Set up a connection to the server.
	conn, err := grpc.Dial(ip+port, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	// Create client
	c := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Call RPC
	fmt.Print("Sending RPC... My unique ID is: ", name, "\n")
	_, err = c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		return err
	}

	return nil
}
