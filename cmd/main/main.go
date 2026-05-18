// cmd/main.go
package main

import (
	"go.uber.org/zap"
	"proxy/configGolang"
	metrics2 "proxy/iternal/metrics"
	"proxy/iternal/repository"
	"proxy/iternal/transport/http"
	"proxy/iternal/usecase"
	"proxy/pkg/logger"
)

func main() {
	cfg := configGolang.CreateConfig("./config.json")

	logger.InitLogger(cfg.Env)
	defer logger.Sync()

	logger.Info("Starting application",
		zap.String("env", cfg.Env),
		zap.String("port", cfg.ServerPort))

	whiteListRepo, err := repository.NewListRepository("./data/whitelist.json")
	if err != nil {
		logger.Fatal("Failed to create white repo", zap.Error(err))
	}

	blackListRepo, err := repository.NewListRepository("./data/blacklist.json")
	if err != nil {
		logger.Fatal("Failed to create black repo", zap.Error(err))
	}

	grayListRepo, err := repository.NewListRepository("./data/graylist.json")
	if err != nil {
		logger.Fatal("Failed to create gray repo", zap.Error(err))
	}

	rateLimiterRepo, err := repository.NewRateLimiterRepository(
		"./data/rateLimiter.json",
		"./data/rateLimiterData.json",
	)
	if err != nil {
		logger.Fatal("Failed to create rate limiter repo", zap.Error(err))
	}

	listUseCase := usecase.CreateListUseCase(whiteListRepo, grayListRepo, blackListRepo)
	rateLimiterUC := usecase.NewRateLimiterUseCase(rateLimiterRepo)
	listHandler := http.CreateIpHandler(listUseCase, rateLimiterUC, cfg)

	metrics := metrics2.CreateMetricsPuller(listUseCase, rateLimiterUC)
	metrics.Init()

	router := http.SetupRoute(listHandler)

	logger.Info("Server starting",
		zap.String("port", cfg.ServerPort))

	if err := router.Run(cfg.ServerPort); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
