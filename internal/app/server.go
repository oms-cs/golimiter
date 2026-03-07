package app

import (
	"context"
	"log"
	"net"

	proto "github.com/omscs/golimiter/gen/pb"
	"github.com/omscs/golimiter/internal"
	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedGoLimiterServer
	matcher *internal.PathMatcher
}

func (s *server) CheckLimit(ctx context.Context, req *proto.RateLimitRequest) (*proto.RateLimitResponse, error) {
	log.Printf("Received: %v", req.GetPath())
	rules, algorithm, err := s.matcher.Search(req.Path, req.Service, req.Method)
	if err != nil {
		log.Fatalf("no rules found, due to %v", err.Error())
	}

	return HandleRateLimiter(rules, algorithm, ctx, req)
}

type Config struct {
	Port string
}

func Run(cfg Config, matcher *internal.PathMatcher) error {
	lis, err := net.Listen("tcp", ":"+cfg.Port)
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	proto.RegisterGoLimiterServer(s, &server{matcher: matcher})
	log.Printf("Server is running on port : %s... \n", cfg.Port)
	return s.Serve(lis)
}
