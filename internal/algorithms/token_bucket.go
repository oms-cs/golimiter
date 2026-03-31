package algorithms

import (
	"log/slog"

	proto "github.com/omscs/golimiter/gen/pb"
)

// TokenBucket implements token bucket rate limiting algorithm
type TokenBucket struct {
	*BaseAlgorithm
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket() RateLimiter {
	return &TokenBucket{
		BaseAlgorithm: NewBaseAlgorithm("token_bucket"),
	}
}

// IsAllowed checks if the request is allowed based on token bucket algorithm
func (tb *TokenBucket) IsAllowed(req *proto.RateLimitRequest, limits []byte, weight int) *proto.RateLimitResponse {
	res, err := tb.ExecuteScript(req, limits, weight)
	if err != nil {
		// Log error but default to allowing the request to avoid service disruption
		slog.Error("failed to execute script", "error", err)
		return &proto.RateLimitResponse{
			IsAllowed: true,
		}
	}
	return res
}
