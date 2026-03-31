package app

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	proto "github.com/omscs/golimiter/gen/pb"
	"github.com/omscs/golimiter/internal"
	"google.golang.org/grpc"

	grpc_prom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type server struct {
	proto.UnimplementedGoLimiterServer
	matcher *internal.PathMatcher
}

func (s *server) CheckLimit(ctx context.Context, req *proto.RateLimitRequest) (*proto.RateLimitResponse, error) {
	// Validate input parameters
	if req == nil {
		slog.Error("empty request")
		return newDeniedResponse(), nil
	}

	if req.Path == "" {
		slog.Error("received request with no path")
		return newDeniedResponse(), nil
	}

	if req.Method == "" {
		slog.Error("received request with no method",
			slog.String("path", req.Path),
		)
		return newDeniedResponse(), nil
	}

	if len(req.Keys) == 0 {
		slog.Error("received request with no keys",
			slog.String("path", req.Path),
			slog.String("method", req.Method),
		)
		return newDeniedResponse(), nil
	}

	// Validate keys
	for i, key := range req.Keys {
		if key.Value == "" {
			slog.Error("received request with empty key",
				slog.Int("index", i),
				slog.String("path", req.Path),
				slog.String("method", req.Method),
			)
			return newDeniedResponse(), nil
		}
	}

	rules, algorithm, err := s.matcher.Search(req.Path, req.Service, req.Method)
	if err != nil {
		slog.Error("no rules found", slog.String("path", req.Path), slog.String("service", req.Service), slog.String("method", req.Method), "error", err)
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
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prom.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prom.UnaryServerInterceptor),
	)

	grpc_prom.Register(s)
	//enable histograms
	grpc_prom.EnableHandlingTimeHistogram()

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":9092", nil)

	proto.RegisterGoLimiterServer(s, &server{matcher: matcher})
	slog.Debug("server is running on port", slog.String("port", cfg.Port))
	return s.Serve(lis)
}
