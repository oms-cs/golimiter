package main

import (
	"log/slog"
	"os"

	"github.com/omscs/golimiter/internal"
	"github.com/omscs/golimiter/internal/app"
)

func main() {
	configFile := "rate_limit_config.yml"

	//log handler
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Defaults to Info if not set
	})

	// 2. Make it the global logger
	logger := slog.New(handler)
	slog.SetDefault(logger)

	//load configs from file
	configs, err := internal.LoadConfig(configFile)
	if err != nil {
		slog.Error("Failed to Load Rate Limiter Configs: ", "error", err)
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
		slog.Error("failed to start server", slog.String("port", serverConfig.Port), "error", err)
		os.Exit(0)
	}
	slog.Info("server started", slog.String("port", serverConfig.Port))
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
