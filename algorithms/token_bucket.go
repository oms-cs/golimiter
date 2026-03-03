package algorithms

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/omscs/golimiter/infrastructure"
)

func IsAllowed(req *http.Request) bool {
	remoteAddr := req.RemoteAddr
	fmt.Printf("received req from remote address : %s", remoteAddr)

	// TODO
	// 1. get keys from request, initially only remoteIpAddress
	// 2. create a bucket for each key and increment if key already exists of initiate to 1
	// 3. define leak rate, every time a request is allowed increment the bucket
	// 4. every time request is hit check how much time has passed since last request
	// 5. and calculate how many requests can be allowed based on leak rate
	// 6. return if allowed or not

	redis := infrastructure.RedisClient()
	ctx := context.Background()

	//Read lua Script Path
	filePath := filepath.Join("scripts", "leaky_bucket.lua")
	key := remoteAddr
	luaScript, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
	}

	redis.Eval(ctx, string(luaScript), []string{key})

	// 7. any time bucket is filled return http.TooManyAttemptsException
	return true
}
