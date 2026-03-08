package algorithms

import (
	"log"

	proto "github.com/omscs/golimiter/gen/pb"
)

// SlidingWindowCounter implements sliding window counter rate limiting algorithm
type SlidingWindowSet struct {
	*BaseAlgorithm
}

// NewSlidingWindowCounter creates a new sliding window counter rate limiter
func NewSlidingWindowSet() RateLimiter {
	return &SlidingWindowCounter{
		BaseAlgorithm: NewBaseAlgorithm("sliding_window_set"),
	}
}

// IsAllowed checks if the request is allowed based on sliding window counter algorithm
func (sws *SlidingWindowSet) IsAllowed(req *proto.RateLimitRequest, limits []byte) *proto.RateLimitResponse {
	res, err := sws.ExecuteScript(req, limits)
	if err != nil {
		// Log error but default to allowing the request to avoid service disruption
		log.Printf("failed to execute script due to %v \n", err)
		return &proto.RateLimitResponse{
			IsAllowed: true,
		}
	}
	return res
}
