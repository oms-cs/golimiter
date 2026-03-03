package algorithms

import (
	proto "github.com/omscs/golimiter/gen/go"
)

type RateLimiterAlgorithm interface {
	IsAllowed(req *proto.RateLimitRequest) bool
}
