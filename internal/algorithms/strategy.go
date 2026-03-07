package algorithms

import (
	proto "github.com/omscs/golimiter/gen/go"
)

type RateLimiter interface {
	IsAllowed(req *proto.RateLimitRequest) bool
}

type TokenBucket struct {
	*tokenBucket
}

type SlidingWindowCounter struct {
	*slidingWindowCounter
}

type SlidingWindowLog struct {
	*slidingWindowLog
}

func NewTokenBucket() RateLimiter {
	return &TokenBucket{tokenBucket: &tokenBucket{}}
}

func NewSlidingWindowCounter() RateLimiter {
	return &SlidingWindowCounter{slidingWindowCounter: &slidingWindowCounter{}}
}

func NewSlidingWindowLog() RateLimiter {
	return &SlidingWindowLog{slidingWindowLog: &slidingWindowLog{}}
}
