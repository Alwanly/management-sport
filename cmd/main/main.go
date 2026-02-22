package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/pkg/authentication"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg/logger"
	"github.com/Alwanly/management-sport/pkg/middleware"
	"github.com/Alwanly/management-sport/pkg/redis"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// @title Codebase API Example documentation
// @version 1.0
// @description This is a sample server.

// @host 127.0.0.1:9000
// @BasePath /

// @securityDefinitions.basic BasicAuth
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// load config
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Setup dependencies
	globalLogger := logger.NewLogger(cfg.ServiceName, cfg.LogLevel,
		logger.WithPrettyPrint(),
	)
	l := logger.WithID(globalLogger, "server", "main")
	l.Info("Starting application",
		zap.String("service", cfg.ServiceName),
		zap.String("version", cfg.ServiceVersion),
		zap.String("environment", cfg.Environment),
	)

	// Create upload directories
	uploadDir := filepath.Join(cfg.UploadDirectory, "logo_teams")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		l.Error("Failed to create upload directory", zap.Error(err))
		os.Exit(1)
	}
	l.Info("Upload directory ready", zap.String("path", uploadDir))

	// Setup database
	dbConfig := database.DBServiceOpts{
		Debug:                      cfg.Debug,
		Logger:                     globalLogger,
		PostgresURI:                &cfg.PostgresURI,
		PostgresMaxOpenConnections: cfg.PostgresMaxOpenConnections,
		PostgresMaxIdleConnections: cfg.PostgresMaxIdleConnections,
	}
	db, err := database.NewPostgres(&dbConfig)
	if err != nil {
		l.Error("Failed to initialize database", zap.Error(err))
		os.Exit(1)
	}

	// Setup redis
	redisConfig := redis.Opts{
		Logger:   globalLogger,
		RedisURI: &cfg.RedisURI,
	}
	redisClient, err := redis.NewRedis(&redisConfig)
	if err != nil {
		l.Error("Failed to initialize Redis", zap.Error(err))
		os.Exit(1)
	}

	// Setup JWT middleware
	jwtService := authentication.NewJWTService(&authentication.JWTConfig{
		SecretKey:      cfg.JwtSecret,
		Audience:       cfg.JwtAudience,
		Issuer:         cfg.JwtIssuer,
		ExpirationTime: cfg.JwtExpirationTime,
		RefreshTime:    cfg.JwtRefreshTime,
	})

	// Setup Basic Auth middleware
	basicAuthService := authentication.NewBasicAuthService(&authentication.BasicAuthTConfig{
		Username: cfg.BasicAuthUsername,
		Password: cfg.BasicAuthPassword,
	})

	authMiddleware := middleware.NewAuthMiddleware(jwtService, basicAuthService)

	// Bootstrap application
	app := Bootstrap(&AppDeps{
		Config: &cfg,
		Logger: globalLogger,
		DB:     db,
		Redis:  redisClient,
		Auth:   authMiddleware,
	})

	// Create HTTP server for graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: app.Gin,
	}

	ctx, cancel := context.WithCancel(context.Background())
	g, gCtx := errgroup.WithContext(ctx)

	// Start server
	g.Go(func() error {
		l.Info("Starting server...", zap.Int("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("Server error", zap.Error(err))
			return err
		}
		return nil
	})

	// Graceful shutdown
	g.Go(func() error {
		<-gCtx.Done()
		l.Info("Gracefully shutting down...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			l.Error("Cannot shutdown server", zap.Error(err))
			return err
		}

		if err := app.DB.Close(); err != nil {
			l.Error("Cannot close database connection", zap.Error(err))
			return err
		}

		return nil
	})

	// Signal handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-c
		l.Info("Received shutdown signal")
		cancel()
	}()

	if err := g.Wait(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	l.Info("Server stopped")
}
