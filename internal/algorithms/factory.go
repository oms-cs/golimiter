package algorithms

import (
	"fmt"
)

type RateLimiterFactory interface {
	CreateRateLimiter(algorithm string) (RateLimiter, error)
}

type DefaultRateLimiterFactory struct{}

func NewRateLimiterFactory() RateLimiterFactory {
	return &DefaultRateLimiterFactory{}
}

func (f *DefaultRateLimiterFactory) CreateRateLimiter(algorithm string) (RateLimiter, error) {
	switch algorithm {
	case "token_bucket":
		return NewTokenBucket(), nil
	case "sliding_window_counter":
		return NewSlidingWindowCounter(), nil
	case "sliding_window_log":
		return NewSlidingWindowLog(), nil
	default:
		return nil, fmt.Errorf("unsupported rate limiting algorithm: %s", algorithm)
	}
}

// Convenience function for direct creation
func CreateRateLimiter(algorithm string) (RateLimiter, error) {
	factory := NewRateLimiterFactory()
	return factory.CreateRateLimiter(algorithm)
}
