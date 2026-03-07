package algorithms

import (
	proto "github.com/omscs/golimiter/gen/pb"
)

// RateLimiter defines the interface for all rate limiting algorithms
type RateLimiter interface {
	IsAllowed(req *proto.RateLimitRequest, limits []byte) *proto.RateLimitResponse
}
