package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LalatinaHub/LatinaApi/config"
	"github.com/LalatinaHub/LatinaApi/internal/api"
	"github.com/LalatinaHub/LatinaApi/internal/api/handler"
	"github.com/LalatinaHub/LatinaApi/internal/api/middleware"
	"github.com/LalatinaHub/LatinaApi/internal/repository"
	"github.com/LalatinaHub/LatinaApi/internal/service/converter"
	"github.com/LalatinaHub/LatinaApi/internal/service/dbadmin"
	"github.com/LalatinaHub/LatinaApi/internal/service/subscription"
	"github.com/LalatinaHub/LatinaApi/internal/service/user"
	"github.com/LalatinaHub/LatinaApi/pkg/logger"
	"github.com/LalatinaHub/common/database"
)

func main() {
	// 1. Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load application configuration")
	}

	// 2. Setup structured zerolog logger
	logger.SetupLogger(cfg.AppEnv, cfg.IsProduction())
	logger.Info().
		Str("app_env", cfg.AppEnv).
		Str("port", cfg.Port).
		Msg("Starting LatinaApi service...")

	// 3. Initialize database connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.GetDB()
	if err != nil {
		logger.Warn().Err(err).Msg("Turso database connection not established, running in degraded mode")
	} else {
		logger.Info().Msg("Connected to Turso LibSQL database pool")
		if err := database.InitIndexes(ctx); err != nil {
			logger.Warn().Err(err).Msg("Failed to verify/create database indexes")
		} else {
			logger.Info().Msg("Database indexes verified successfully")
		}
	}

	// 4. Wire repository layer
	var (
		userRepo   repository.UserRepository
		serverRepo repository.ServerRepository
		proxyRepo  repository.ProxyRepository
		kvRepo     repository.KVRepository
	)

	if db != nil {
		userRepo = repository.NewUserRepository(db)
		serverRepo = repository.NewServerRepository(db)
		proxyRepo = repository.NewProxyRepository(db)
		kvRepo = repository.NewKVRepository(db)
	}

	// 5. Wire domain services
	convService := converter.NewConverterService()
	subService := subscription.NewSubscriptionService(userRepo, serverRepo, proxyRepo, convService, cfg.DefaultSubscriptionTitle)
	userService := user.NewUserService(userRepo, kvRepo)
	adminService := dbadmin.NewDBAdminService(db, kvRepo)

	// 6. Wire handlers & middleware
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	router := api.SetupRouter(api.RouterConfig{
		SubHandler:     handler.NewSubHandler(subService),
		UserHandler:    handler.NewUserHandler(userService),
		AdminHandler:   handler.NewAdminHandler(adminService),
		HealthHandler:  handler.NewHealthHandler(),
		InfoHandler:    handler.NewInfoHandler(),
		ConvertHandler: handler.NewConvertHandler(convService),
		RateLimiter:    rateLimiter,
		IsProduction:   cfg.IsProduction(),
	})

	// 7. Start HTTP server
	srv := &http.Server{
		Addr:              cfg.Address(),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info().Str("addr", srv.Addr).Msg("HTTP server is listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Err(err).Msg("Server failed to listen and serve")
		}
	}()

	// 8. Graceful shutdown on SIGINT or SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info().Msg("Shutting down LatinaApi server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("HTTP server forced to shutdown")
	}

	if err := database.Close(); err != nil {
		logger.Warn().Err(err).Msg("Failed to close database pool cleanly")
	}

	logger.Info().Msg("LatinaApi service exited cleanly")
}
