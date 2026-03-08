package algorithms

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	proto "github.com/omscs/golimiter/gen/pb"
	"github.com/omscs/golimiter/internal/infrastructure"
	"github.com/redis/go-redis/v9"
)

// RedisClient interface for better testability
type RedisClient interface {
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
}

// ScriptCache caches Lua scripts in memory
type ScriptCache struct {
	scripts map[string]string
	mutex   sync.RWMutex
}

var scriptCache = &ScriptCache{
	scripts: make(map[string]string),
}

// GetScript retrieves a script from cache or loads it from disk
func (sc *ScriptCache) GetScript(scriptName string) (string, error) {
	sc.mutex.RLock()
	script, exists := sc.scripts[scriptName]
	sc.mutex.RUnlock()

	if exists {
		return script, nil
	}

	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	// Double-check after acquiring write lock
	if script, exists := sc.scripts[scriptName]; exists {
		return script, nil
	}

	// Load script from disk
	scriptPath := filepath.Join("scripts", scriptName+".lua")
	luaScript, err := os.ReadFile(scriptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read Lua script %s: %w", scriptName, err)
	}

	scriptStr := string(luaScript)
	sc.scripts[scriptName] = scriptStr
	return scriptStr, nil
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
func (ba *BaseAlgorithm) ExecuteScript(req *proto.RateLimitRequest, limits []byte) (*proto.RateLimitResponse, error) {
	fmt.Printf("Processing rate limit request for path: %s using algorithm: %s\n", req.Path, ba.scriptName)

	limitsStr := string(limits)

	// Get script from cache
	luaScript, err := scriptCache.GetScript(ba.scriptName)
	if err != nil {
		return &proto.RateLimitResponse{
			IsAllowed: false,
		}, fmt.Errorf("failed to get Lua script %s: %w", ba.scriptName, err)
	}

	// Extract keys from request
	keys := make([]string, 0, len(req.Keys))
	for _, key := range req.Keys {
		keys = append(keys, key.Value)
	}

	// Execute script
	ctx := context.Background()
	nowMs := time.Now().UnixMilli()
	result, err := ba.redisClient.Eval(ctx, luaScript, keys, limitsStr, nowMs).Result()
	if err != nil {
		return &proto.RateLimitResponse{
			IsAllowed: false,
		}, fmt.Errorf("failed to execute Lua script: %w", err)
	}

	// Parse result
	vals, ok := result.([]interface{})
	if !ok || len(vals) == 0 {
		log.Printf("unexpected result format from Lua script %v", vals)
		return &proto.RateLimitResponse{
			IsAllowed: false,
		}, fmt.Errorf("unexpected result format from Lua script")
	}

	log.Printf("vals : %v \n ", vals)
	isAllowedVal, ok := vals[0].(int64)
	remaining, ok := vals[1].(int64)
	tryAgainDuration, ok := vals[2].(int64)

	if !ok {
		return &proto.RateLimitResponse{
			IsAllowed: false,
		}, fmt.Errorf("failed to convert result to int")
	}

	// Parse the actual result (assuming "1" means allowed, "0" means denied)
	return &proto.RateLimitResponse{
		IsAllowed:        isAllowedVal == 1,
		RemainingTokens:  int32(remaining),
		TryAgainDuration: tryAgainDuration,
	}, nil
}
