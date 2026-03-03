package app

import (
	"context"
	"log"
	"net"

	proto "github.com/omscs/golimiter/gen/go"
	"google.golang.org/grpc"
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

type Config struct {
	Port string
}

func Run(cfg Config) error {
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	proto.RegisterGoLimiterServer(s, &server{})
	log.Printf("Server is running on port : %s... \n", cfg.Port)
	return s.Serve(lis)
}
