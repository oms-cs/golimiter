package app

import (
	"context"
	"encoding/json"
	"log"

	proto "github.com/omscs/golimiter/gen/pb"
	"github.com/omscs/golimiter/internal"
	"github.com/omscs/golimiter/internal/algorithms"
)

// Helper functions for common responses
func newErrorResponse(isAllowed bool) *proto.RateLimitResponse {
	return &proto.RateLimitResponse{
		IsAllowed:        isAllowed,
		RemainingTokens:  0,
		TryAgainDuration: 0,
	}
}

func newAllowedResponse() *proto.RateLimitResponse {
	return newErrorResponse(true)
}

func newDeniedResponse() *proto.RateLimitResponse {
	return newErrorResponse(false)
}

func HandleRateLimiter(rules *internal.RuleSet, algorithm string, ctx context.Context, req *proto.RateLimitRequest) (*proto.RateLimitResponse, error) {

	// initiate factory
	factory := algorithms.NewRateLimiterFactory()

	// check algorithm support
	alg, err1 := algorithms.GetAlgorithmType(algorithm)
	if err1 != nil {
		log.Printf("Unsupported algorithm %s: %v", algorithm, err1)
		return newAllowedResponse(), nil
	}

	//initiate rate limiter from factory
	rateLimiter, err2 := factory.CreateRateLimiterWithConfig(alg, nil)
	if err2 != nil {
		log.Printf("Failed to create rate limiter for algorithm %s: %v", algorithm, err2)
		return newAllowedResponse(), nil
	}

	//convert to lua adjustable format
	limitValues := make([][]int, 0, len(rules.Limits))
	weight := 1

	for _, rule := range rules.Limits {
		limitValues = append(limitValues, []int{rule.WindowSeconds, rule.Limit, rule.Precision})
	}

	//convert limits to json
	jsonData, err := json.Marshal(limitValues)
	if err != nil {
		log.Printf("Failed to marshal limits: %v", err)
		return newAllowedResponse(), nil
	}

	//return response
	return rateLimiter.IsAllowed(req, jsonData, weight), nil
}
