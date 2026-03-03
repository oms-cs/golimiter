package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/omscs/golimiter/proto"
	"google.golang.org/grpc"
)

const (
	defaultExpirationTime = 2 * time.Minute
	sleepTime             = 2 * time.Second
)

type server struct {
	proto.UnimplementedGoLimiterServer
}

func (s *server) CheckLimit(ctx context.Context, req *proto.RateLimitRequest) (*proto.RateLimitResponse, error) {
	log.Printf("Received: %v", req.GetPath())
	return &proto.RateLimitResponse{
		IsAllowed: true,
	}, nil
}

func main() {
	fmt.Printf("Hello world! \n")
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	proto.RegisterGoLimiterServer(s, &server{})

	log.Println("Server is running on port :50051...")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
