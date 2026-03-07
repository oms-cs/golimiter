package algorithms

import (
	"fmt"
)

// AlgorithmType represents supported rate limiting algorithms
type AlgorithmType string

const (
	TokenBucketAlgorithm          AlgorithmType = "token_bucket"
	SlidingWindowCounterAlgorithm AlgorithmType = "sliding_window_counter"
	SlidingWindowLogAlgorithm     AlgorithmType = "sliding_window_log"
)

// RateLimiterFactory defines the interface for creating rate limiters
type RateLimiterFactory interface {
	CreateRateLimiter(algorithm AlgorithmType) (RateLimiter, error)
	CreateRateLimiterWithConfig(algorithm AlgorithmType, config map[string]interface{}) (RateLimiter, error)
}

// DefaultRateLimiterFactory implements the factory pattern
type DefaultRateLimiterFactory struct{}

// NewRateLimiterFactory creates a new rate limiter factory
func NewRateLimiterFactory() RateLimiterFactory {
	return &DefaultRateLimiterFactory{}
}

// CreateRateLimiter creates a rate limiter for the specified algorithm
func (f *DefaultRateLimiterFactory) CreateRateLimiter(algorithm AlgorithmType) (RateLimiter, error) {
	return f.CreateRateLimiterWithConfig(algorithm, nil)
}

// CreateRateLimiterWithConfig creates a rate limiter with additional configuration
func (f *DefaultRateLimiterFactory) CreateRateLimiterWithConfig(algorithm AlgorithmType, config map[string]interface{}) (RateLimiter, error) {
	switch algorithm {
	case TokenBucketAlgorithm:
		return NewTokenBucket(), nil
	case SlidingWindowCounterAlgorithm:
		return NewSlidingWindowCounter(), nil
	case SlidingWindowLogAlgorithm:
		return NewSlidingWindowLog(), nil
	default:
		return nil, fmt.Errorf("unsupported rate limiting algorithm: %s", algorithm)
	}
}

// GetSupportedAlgorithms returns a list of all supported algorithms
func GetSupportedAlgorithms() []AlgorithmType {
	return []AlgorithmType{
		TokenBucketAlgorithm,
		SlidingWindowCounterAlgorithm,
		SlidingWindowLogAlgorithm,
	}
}

// IsValidAlgorithm checks if the algorithm type is supported
func IsValidAlgorithm(algorithm AlgorithmType) bool {
	for _, supported := range GetSupportedAlgorithms() {
		if algorithm == supported {
			return true
		}
	}
	return false
}

// Convenience function for direct creation
func CreateRateLimiter(algorithm AlgorithmType) (RateLimiter, error) {
	factory := NewRateLimiterFactory()
	return factory.CreateRateLimiter(algorithm)
}
