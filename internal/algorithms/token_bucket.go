package algorithms

import (
	proto "github.com/omscs/golimiter/gen/go"
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
func (tb *TokenBucket) IsAllowed(req *proto.RateLimitRequest) bool {
	allowed, err := tb.ExecuteScript(req)
	if err != nil {
		// Log error but default to allowing the request to avoid service disruption
		// In production, you might want to use a circuit breaker or other fallback
		return true
	}
	return allowed
}
