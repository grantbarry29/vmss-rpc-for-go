package rpc

import (
	"context"
	"flag"
	"log"
	"time"

	pb "vmss-rpc-for-go/rpc/stubs"

	"google.golang.org/grpc"
)

const (
	defaultName   = "world"
	serverAddress = "localhost:50051"
)

var (
	name = flag.String("name", defaultName, "Name to greet")
)

func Client() {

	// Set up a connection to the server.
	conn, err := grpc.Dial(serverAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
