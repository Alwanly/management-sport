package main

import (
	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg/deps"
	"github.com/Alwanly/management-sport/pkg/health"
	"github.com/Alwanly/management-sport/pkg/middleware"
	"github.com/Alwanly/management-sport/pkg/redis"
	"github.com/Alwanly/management-sport/pkg/validator"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"net/http"

	_ "github.com/Alwanly/management-sport/api"
	audit_handler "github.com/Alwanly/management-sport/internal/audit/handler"
	auth_handler "github.com/Alwanly/management-sport/internal/auth/handler"
	book_handler "github.com/Alwanly/management-sport/internal/example/handler"
	goal_handler "github.com/Alwanly/management-sport/internal/goal/handler"
	match_handler "github.com/Alwanly/management-sport/internal/match/handler"
	player_handler "github.com/Alwanly/management-sport/internal/player/handler"
	report_handler "github.com/Alwanly/management-sport/internal/report/handler"
	team_handler "github.com/Alwanly/management-sport/internal/team/handler"
)

type AppDeps struct {
	Config *config.GlobalConfig
	Logger *zap.Logger
	DB     *database.DBService
	Redis  *redis.Service
	Auth   *middleware.AuthMiddleware
}

var inst *deps.App

// @title Management Sport API
// @version 1.0
// @description This is a Management Sport API server with Gin framework
// @host localhost:9000
// @BasePath /
// @securityDefinitions.basic BasicAuth
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func Bootstrap(d *AppDeps) *deps.App {
	// Set Gin mode
	if d.Config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin
	e := gin.New()

	// Add middleware
	e.Use(gin.Logger())
	e.Use(middleware.Recover(d.Logger))
	e.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	v, _ := validator.NewValidator()

	// Swagger

	e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	inst = &deps.App{
		Config:    d.Config,
		Logger:    d.Logger,
		DB:        d.DB,
		Redis:     d.Redis,
		Auth:      d.Auth,
		Gin:       e,
		Validator: v,
	}

	// Run migrations
	database.MigrateIfNeed(inst.DB.Gorm)

	// Health endpoints
	healthHandler := health.NewHandler(d.Config.ServiceName, d.Config.ServiceVersion)
	e.GET("/health", healthHandler.Check)
	e.GET("/ready", healthHandler.Readiness)
	e.GET("/live", healthHandler.Liveness)

	// Register handlers
	book_handler.NewHandler(inst)
	team_handler.NewHandler(inst)
	player_handler.NewHandler(inst)
	match_handler.NewHandler(inst)
	goal_handler.NewHandler(inst)
	audit_handler.NewHandler(inst)
	report_handler.NewHandler(inst)
	auth_handler.NewHandler(inst)

	// Admin-only test endpoint (for RBAC verification)
	e.GET("/admin/test", inst.Auth.AdminAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access ok"})
	})

	return inst
}
