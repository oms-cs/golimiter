package algorithms

import (
	proto "github.com/omscs/golimiter/gen/pb"
)

// SlidingWindowCounter implements sliding window counter rate limiting algorithm
type SlidingWindowCounter struct {
	*BaseAlgorithm
}

// NewSlidingWindowCounter creates a new sliding window counter rate limiter
func NewSlidingWindowCounter() RateLimiter {
	return &SlidingWindowCounter{
		BaseAlgorithm: NewBaseAlgorithm("sliding_window_counter"),
	}
}

// IsAllowed checks if the request is allowed based on sliding window counter algorithm
func (swc *SlidingWindowCounter) IsAllowed(req *proto.RateLimitRequest, limits []byte) *proto.RateLimitResponse {
	res, err := swc.ExecuteScript(req, limits)
	if err != nil {
		// Log error but default to allowing the request to avoid service disruption
		return &proto.RateLimitResponse{
			IsAllowed: true,
		}
	}
	return res
}
