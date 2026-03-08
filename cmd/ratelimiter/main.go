package main

import (
	"fmt"
	"log"
	"os"

	"github.com/omscs/golimiter/internal"
	"github.com/omscs/golimiter/internal/app"
)

func main() {
	fmt.Printf("Hello world! \n")

	configFile := "rate_limit_config.yml"

	//load configs from file
	configs, err := internal.LoadConfig(configFile)
	if err != nil {
		fmt.Printf("Failed to Load Rate Limiter Configs : %v \n", err.Error())
		os.Exit(0)
	}

	// instantiate path matcher
	pathMatcher := internal.NewPathMatcher()

	// load Rate Limiter config
	loadPathMatcher(configs, pathMatcher)

	// Get port from environment variable or use default
	port := getEnvOrDefault("PORT", "50051")
	serverConfig := app.Config{Port: port}

	// start server
	if err := app.Run(serverConfig, pathMatcher); err != nil {
		log.Printf("failed to start server on port %s , due to : %v \n", serverConfig.Port, err)
		os.Exit(0)
	}
	log.Printf("server started on port %s \n", serverConfig.Port)
}

func loadPathMatcher(configs *internal.Config, pathMatcher *internal.PathMatcher) {
	for _, config := range configs.Resources {
		paths := config.Paths
		for _, path := range paths {
			pathMatcher.Insert(&path, config.Service, config.Algorithm)
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
