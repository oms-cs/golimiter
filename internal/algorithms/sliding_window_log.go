package algorithms

import (
	"log"

	proto "github.com/omscs/golimiter/gen/pb"
)

// SlidingWindowLog implements sliding window log rate limiting algorithm
type SlidingWindowLog struct {
	*BaseAlgorithm
}

// NewSlidingWindowLog creates a new sliding window log rate limiter
func NewSlidingWindowLog() RateLimiter {
	return &SlidingWindowLog{
		BaseAlgorithm: NewBaseAlgorithm("sliding_window_log"),
	}
}

// IsAllowed checks if the request is allowed based on sliding window log algorithm
func (swl *SlidingWindowLog) IsAllowed(req *proto.RateLimitRequest, limits []byte, weight int) *proto.RateLimitResponse {
	res, err := swl.ExecuteScript(req, limits, weight)
	if err != nil {
		// Log error but default to allowing the request to avoid service disruption
		log.Printf("failed to execute script due to %v \n", err)
		return &proto.RateLimitResponse{
			IsAllowed: true,
		}
	}
	return res
}
