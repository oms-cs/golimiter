package algorithms

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	proto "github.com/omscs/golimiter/gen/go"
	"github.com/omscs/golimiter/internal/infrastructure"
)

type slidingWindowLog struct {
}

func (swl *slidingWindowLog) IsAllowed(req *proto.RateLimitRequest) bool {
	fmt.Printf("received req from remote address : %s \n", req.Path)

	redis := infrastructure.RedisClient()
	ctx := context.Background()

	//Read lua Script Path
	filePath := filepath.Join("scripts", "sliding_window_log.lua")
	keys := make([]string, len(req.Keys))

	for _, key := range req.Keys {
		keys = append(keys, key.Value)
	}

	luaScript, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
	}

	result, err := redis.Eval(ctx, string(luaScript), keys).Result()
	if err != nil {
		panic(err)
	}

	vals := result.([]interface{})

	fmt.Println("vals", vals[0].(string))

	return true
}
