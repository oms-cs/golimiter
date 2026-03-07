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
	port := getEnvOrDefault("PORT", "510001")
	serverConfig := app.Config{Port: port}

	// start server
	app.Run(serverConfig, pathMatcher)
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
