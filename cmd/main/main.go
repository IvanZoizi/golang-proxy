package main

import (
	"context"
	"go.uber.org/zap"
	http2 "net/http"
	"os"
	"os/signal"
	"proxy/configGolang"
	_ "proxy/docs"
	metrics2 "proxy/internal/metrics"
	"proxy/internal/repository"
	"proxy/internal/transport/http"
	"proxy/internal/usecase"
	"proxy/pkg/logger"
	"syscall"
	"time"
)

// @title Proxy API Documentation
// @version 1.0
// @description API документация для прокси-сервера
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@proxy.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: Bearer {token}
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

	cacheConfig := configGolang.GetDefaultCacheConfig()
	cacheRepo, err := repository.NewCacheRepository(
		"./data/cache.json",
		cacheConfig.MaxCacheSize,
		cacheConfig.MaxEntries,
	)
	if err != nil {
		panic("Failed to create cache repository: " + err.Error())
	}

	cacheUC := usecase.NewCacheUseCase(cacheRepo, cacheConfig)
	cacheMiddleware := http.NewCacheMiddleware(cacheUC)
	cacheHandler := http.NewCacheHandler(cacheUC, cacheConfig, cacheMiddleware)

	router := http.SetupRoute(listHandler, cacheHandler, cacheConfig)

	logger.Info("Server starting",
		zap.String("port", cfg.ServerPort))

	srv := &http2.Server{
		Addr:    cfg.ServerPort,
		Handler: router,
	}

	go func() {
		logger.Info("Server starting", zap.String("port", cfg.ServerPort))
		if err := srv.ListenAndServe(); err != nil && err != http2.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	if closer, ok := metrics.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			logger.Error("Failed to close metrics", zap.Error(err))
		}
	}

	logger.Info("Server exited gracefully")
}
