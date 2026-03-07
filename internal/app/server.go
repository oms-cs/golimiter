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
	// Validate input parameters
	if req == nil {
		log.Printf("Received nil request")
		return newDeniedResponse(), nil
	}

	if req.Path == "" {
		log.Printf("Received request with empty path")
		return newDeniedResponse(), nil
	}

	if req.Method == "" {
		log.Printf("Received request with empty method for path %s", req.Path)
		return newDeniedResponse(), nil
	}

	if len(req.Keys) == 0 {
		log.Printf("Received request with no keys for path %s, method %s", req.Path, req.Method)
		return newDeniedResponse(), nil
	}

	// Validate keys
	for i, key := range req.Keys {
		if key.Value == "" {
			log.Printf("Received request with empty key at index %d for path %s, method %s", i, req.Path, req.Method)
			return newDeniedResponse(), nil
		}
	}

	log.Printf("Received: %v", req.GetPath())
	rules, algorithm, err := s.matcher.Search(req.Path, req.Service, req.Method)
	if err != nil {
		log.Printf("No rules found for path %s, service %s, method %s: %v", req.Path, req.Service, req.Method, err)
		return newAllowedResponse(), nil
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
