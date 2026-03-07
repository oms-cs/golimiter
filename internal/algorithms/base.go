package algorithms

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	proto "github.com/omscs/golimiter/gen/go"
	"github.com/omscs/golimiter/internal/infrastructure"
	"github.com/redis/go-redis/v9"
)

// RedisClient interface for better testability
type RedisClient interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
}

// BaseAlgorithm provides common functionality for all rate limiting algorithms
type BaseAlgorithm struct {
	redisClient RedisClient
	scriptName  string
}

// NewBaseAlgorithm creates a new base algorithm instance
func NewBaseAlgorithm(scriptName string) *BaseAlgorithm {
	return &BaseAlgorithm{
		redisClient: infrastructure.RedisClient(),
		scriptName:  scriptName,
	}
}

// ExecuteScript executes the Lua script and returns the result
func (ba *BaseAlgorithm) ExecuteScript(req *proto.RateLimitRequest) (*proto.RateLimitResponse, error) {
	fmt.Printf("Processing rate limit request for path: %s using algorithm: %s\n", req.Path, ba.scriptName)

	// Load Lua script
	scriptPath := filepath.Join("scripts", ba.scriptName+".lua")
	luaScript, err := os.ReadFile(scriptPath)
	if err != nil {
		return &proto.RateLimitResponse{
			IsAllowed: false,
		}, fmt.Errorf("failed to read Lua script %s: %w", ba.scriptName, err)
	}

	// Extract keys from request
	keys := make([]string, 0, len(req.Keys))
	for _, key := range req.Keys {
		keys = append(keys, key.Value)
	}

	// Execute script
	ctx := context.Background()
	result, err := ba.redisClient.Eval(ctx, string(luaScript), keys).Result()
	if err != nil {
		return &proto.RateLimitResponse{
			IsAllowed: false,
		}, fmt.Errorf("failed to execute Lua script: %w", err)
	}

	// Parse result
	vals, ok := result.([]interface{})
	if !ok || len(vals) == 0 {
		return &proto.RateLimitResponse{
			IsAllowed: false,
		}, fmt.Errorf("unexpected result format from Lua script")
	}

	isAllowedVal, ok := vals[0].(int)
	remaining, ok := vals[1].(int32)
	tryAgainDuration, ok := vals[2].(int64)
	if !ok {
		return &proto.RateLimitResponse{
			IsAllowed: false,
		}, fmt.Errorf("failed to convert result to int")
	}

	fmt.Printf("Rate limit result: %s\n", isAllowedVal)

	// Parse the actual result (assuming "1" means allowed, "0" means denied)
	return &proto.RateLimitResponse{
		IsAllowed:        isAllowedVal == 1,
		RemainingTokens:  remaining,
		TryAgainDuration: tryAgainDuration,
	}, nil
}
