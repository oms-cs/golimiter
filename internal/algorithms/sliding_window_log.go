package algorithms

import (
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
func (swl *SlidingWindowLog) IsAllowed(req *proto.RateLimitRequest, limits []byte) *proto.RateLimitResponse {
	res, err := swl.ExecuteScript(req, limits)
	if err != nil {
		// Log error but default to allowing the request to avoid service disruption
		return &proto.RateLimitResponse{
			IsAllowed: true,
		}
	}
	return res
}
