# Management Club Sport Application - Implementation Guide

## Goal
Migrate from Fiber to Gin framework and implement complete football team management system with teams, players, matches, goals, and comprehensive reporting.

## Prerequisites
- [ ] Ensure you are on the `feature/management-club-sport` branch
- [ ] If the branch doesn't exist, create it: `git checkout -b feature/management-club-sport`
- [ ] Ensure PostgreSQL is running and accessible
- [ ] Ensure Redis is running and accessible
- [ ] Ensure Go 1.23+ is installed

---

## Step 0: Migrate from Fiber to Gin Framework

### Step 0.1: Update Dependencies
- [x] Update `go.mod` to replace Fiber with Gin dependencies:

```go
module github.com/Alwanly/management-sport

go 1.23

replace (
	github.com/Alwanly/management-sport/config => ./config
	github.com/Alwanly/management-sport/pkg => ./pkg
)

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/go-playground/locales v0.14.1
	github.com/go-playground/universal-translator v0.18.1
	github.com/go-playground/validator/v10 v10.24.0
	github.com/go-redis/redis/v9 v9.0.0-rc.1
	github.com/goccy/go-json v0.10.4
	github.com/golang-jwt/jwt/v4 v4.5.1
	github.com/google/uuid v1.6.0
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.10.0
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/files v1.0.1
	github.com/swaggo/swag v1.16.4
	go.elastic.co/ecszap v1.0.3
	go.uber.org/zap v1.27.0
	golang.org/x/crypto v0.32.0
	golang.org/x/sync v0.10.0
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.25.10
)
```

- [x] Run `go mod tidy` to download new dependencies and remove old ones

- [x] Replace contents of `pkg/deps/App.go`:

```go
package deps

import (
	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg/middleware"
	"github.com/Alwanly/management-sport/pkg/redis"
	"github.com/Alwanly/management-sport/pkg/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type App struct {
	Config    *config.GlobalConfig
	Logger    *zap.Logger
	DB        *database.DBService
	Redis     *redis.Service
	Auth      *middleware.AuthMiddleware
	Validator validator.IValidatorService

	// APIs
	Gin *gin.Engine
}
```

- [x] Replace contents of `pkg/binding/binding.go`:

```go
package binding

import (
	"net/http"
	"reflect"

	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/logger"
	"github.com/Alwanly/management-sport/pkg/middleware"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ContextName = "Binding"

type Binder struct {
	l   *zap.Logger
	ctx *gin.Context
	m   interface{}
}

type Source func(*Binder) error

type ModelBindingError struct {
	Code         int
	ResponseBody wrapper.JSONResult
}

func (e *ModelBindingError) Error() string {
	return "Failed to bind request body"
}

func BindFromBody() Source {
	return func(b *Binder) error {
		if err := b.ctx.ShouldBindJSON(b.m); err != nil {
			b.l.Debug("Error when binding from body", zap.Error(err))
			return err
		}
		return nil
	}
}

func BindFromQuery() Source {
	return func(b *Binder) error {
		if err := b.ctx.ShouldBindQuery(b.m); err != nil {
			b.l.Debug("Error when binding from query string", zap.Error(err))
			return err
		}
		return nil
	}
}

func BindFromParams() Source {
	return func(b *Binder) error {
		if err := b.ctx.ShouldBindUri(b.m); err != nil {
			b.l.Debug("Error when binding from path params", zap.Error(err))
			return err
		}
		return nil
	}
}

func BindFromHeaders() Source {
	return func(b *Binder) error {
		if err := b.ctx.ShouldBindHeader(b.m); err != nil {
			b.l.Debug("Error when binding from request headers", zap.Error(err))
			return err
		}
		return nil
	}
}

func BindModel(log *zap.Logger, c *gin.Context, m interface{}, sources ...Source) error {
	l := logger.WithID(log, ContextName, "BindModel")

	binder := &Binder{l: l, ctx: c, m: m}

	for _, source := range sources {
		if err := source(binder); err != nil {
			result := wrapper.ResponseFailed(http.StatusBadRequest, contract.StatusCodeBindingFailed, contract.ErrorValidatePayload, nil)
			return &ModelBindingError{Code: result.Code, ResponseBody: result}
		}
	}

	// Set AuthUserData from context
	if authUserValue, exists := c.Get(middleware.LocalTokenKey); exists {
		if authUser, ok := authUserValue.(*middleware.AuthUserData); ok {
			dataField := reflect.Indirect(reflect.ValueOf(m)).FieldByName("AuthUserData")
			if dataField.IsValid() && dataField.CanSet() {
				dataField.Set(reflect.ValueOf(authUser))
			}
		}
	}

	return nil
}
```

- [x] Replace contents of `pkg/middleware/authentication.go`:

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/Alwanly/management-sport/pkg/authentication"
	"github.com/gin-gonic/gin"
)

type IAuthMiddleware interface {
	JwtAuth() gin.HandlerFunc
	BasicAuth() gin.HandlerFunc
}

type AuthMiddleware struct {
	Jwt   authentication.IJwtService
	Basic authentication.IBasicAuthService
}

type AuthUserData struct {
	UserID string `json:"userId"`
}

const LocalTokenKey = "user"

func NewAuthMiddleware(jwt authentication.IJwtService, basic authentication.IBasicAuthService) *AuthMiddleware {
	return &AuthMiddleware{
		Jwt:   jwt,
		Basic: basic,
	}
}

func (a *AuthMiddleware) JwtAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if !strings.Contains(token, "Bearer") {
			responseUnauthorized(c, "Bearer", "Invalid token")
			return
		}

		token = strings.Replace(token, "Bearer ", "", 1)
		if token == "" {
			responseUnauthorized(c, "Bearer", "Invalid token")
			return
		}

		auth, err := a.Jwt.ParseToken(token)
		if err != nil {
			responseUnauthorized(c, "Bearer", "Invalid token")
			return
		}

		c.Set(LocalTokenKey, decodeAuthToken(*auth))
		c.Next()
	}
}

func (a *AuthMiddleware) BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.Contains(auth, "Basic") {
			responseUnauthorized(c, "Basic", "Invalid auth")
			return
		}

		username, password := a.Basic.DecodeFromHeader(auth)
		if !a.Basic.Validate(username, password) {
			responseUnauthorized(c, "Basic", "Invalid auth")
			return
		}
		c.Next()
	}
}

func responseUnauthorized(c *gin.Context, _ string, message ...string) {
	c.Header("WWW-Authenticate", "Basic realm=Restricted")
	response := gin.H{"message": message[0]}
	if len(message) > 1 {
		response["statusCode"] = message[1]
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, response)
}

func decodeAuthToken(auth authentication.JWTClaims) *AuthUserData {
	return &AuthUserData{
		UserID: auth["userId"].(string),
	}
}
```

- [x] Replace contents of `pkg/middleware/recovery.go`:

```go
package middleware

import (
	"net/http"

	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/logger"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ContextNameRecovery = "Middleware.Recovery"

func Recover(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				l := logger.WithID(log, ContextNameRecovery, "Recover")
				l.Error("Panic recovered", zap.Any("error", err), zap.Stack("stack"))

				result := wrapper.ResponseFailed(
					http.StatusInternalServerError,
					contract.StatusCodeInternalServerError,
					"Internal server error",
					nil,
				)
				c.AbortWithStatusJSON(result.Code, result)
			}
		}()
		c.Next()
	}
}
```

- [x] Replace contents of `pkg/health/health.go`:

```go
package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Health struct {
	ServiceName    string `json:"service"`
	ServiceVersion string `json:"version"`
	Status         string `json:"status"`
}

type HealthHandler struct {
	ServiceName    string
	ServiceVersion string
}

func NewHandler(serviceName, serviceVersion string) *HealthHandler {
	return &HealthHandler{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	c.JSON(http.StatusOK, Health{
		ServiceName:    h.ServiceName,
		ServiceVersion: h.ServiceVersion,
		Status:         "ok",
	})
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	c.JSON(http.StatusOK, Health{
		ServiceName:    h.ServiceName,
		ServiceVersion: h.ServiceVersion,
		Status:         "ready",
	})
}

func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, Health{
		ServiceName:    h.ServiceName,
		ServiceVersion: h.ServiceVersion,
		Status:         "alive",
	})
}
```

- [x] Replace contents of `cmd/main/bootstrap.go`:

```go
package main

import (
	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/pkg/authentication"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg/deps"
	"github.com/Alwanly/management-sport/pkg/health"
	"github.com/Alwanly/management-sport/pkg/middleware"
	"github.com/Alwanly/management-sport/pkg/redis"
	"github.com/Alwanly/management-sport/pkg/validator"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"go.uber.org/zap"

	_ "github.com/Alwanly/management-sport/api"
	book_handler "github.com/Alwanly/management-sport/internal/example/handler"
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
	if d.Config.Environment == "development" {
		e.Static("/swagger.yaml", "./api/swagger.yaml")
		e.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

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

	return inst
}
```

- [x] Replace contents of `cmd/main/main.go`:

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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

// @host localhost:9000
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
	jwtService := authentication.NewJWT(&authentication.JWTConfig{
		PrivateKey:     cfg.PrivateKey,
		PublicKey:      cfg.PublicKey,
		Audience:       cfg.JwtAudience,
		Issuer:         cfg.JwtIssuer,
		ExpirationTime: cfg.JwtExpirationTime,
		RefreshTime:    cfg.JwtRefreshTime,
	})

	// Setup Basic Auth middleware
	basicAuthService := authentication.NewBasicAuth(&authentication.BasicAuthTConfig{
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
```

- [x] Replace contents of `internal/example/handler/handler.go`:

```go
package handler

import (
	"github.com/Alwanly/management-sport/internal/example/repository"
	"github.com/Alwanly/management-sport/internal/example/schema"
	"github.com/Alwanly/management-sport/internal/example/usecase"
	"github.com/Alwanly/management-sport/pkg/binding"
	"github.com/Alwanly/management-sport/pkg/deps"
	"github.com/Alwanly/management-sport/pkg/logger"
	"github.com/Alwanly/management-sport/pkg/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ContextName = "Internal.Book.Handler"

type (
	Handler struct {
		Logger    *zap.Logger
		Validator validator.IValidatorService
		UseCase   usecase.IUseCase
	}
)

func NewHandler(d *deps.App) *Handler {
	repository := repository.NewRepository(repository.Repository{
		DB:    d.DB,
		Redis: d.Redis,
	})
	usecase := usecase.NewUseCase(usecase.UseCase{
		Config:     d.Config,
		Logger:     d.Logger,
		Repository: repository,
	})
	handler := &Handler{
		Logger:    d.Logger,
		Validator: d.Validator,
		UseCase:   usecase,
	}

	e := d.Gin.Group("/books/v1")
	e.Use(d.Auth.JwtAuth())
	e.POST("/", handler.Create)
	e.GET("/", handler.List)
	e.GET("/:id", handler.Get)
	e.PUT("/:id", handler.Update)
	e.DELETE("/:id", handler.Delete)
	return handler
}

// Create creates a new book.
func (h *Handler) Create(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Create")

	// bind model
	model := &schema.RequestBookCreate{}
	if err := binding.BindModel(l, c, model, binding.BindFromBody()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// validate model
	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// create a new book
	response := h.UseCase.Create(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

// List returns a list of books.
func (h *Handler) List(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "List")

	// bind model
	model := &schema.RequestBookList{
		Page:      1,
		PageSize:  10,
		SortBy:    "title",
		SortOrder: "desc",
	}
	if err := binding.BindModel(l, c, model, binding.BindFromQuery()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// validate model
	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// get list of books
	response := h.UseCase.List(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

// Get returns a book by ID.
func (h *Handler) Get(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Get")

	// bind model
	model := &schema.RequestBookGet{}
	if err := binding.BindModel(l, c, model, binding.BindFromParams()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// validate model
	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// get a book
	response := h.UseCase.Get(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

// Update updates a book.
func (h *Handler) Update(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Update")

	// bind model
	model := &schema.RequestBookUpdate{}
	if err := binding.BindModel(l, c, model, binding.BindFromParams(), binding.BindFromBody()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// validate model
	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// update a book
	response := h.UseCase.Update(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

// Delete deletes a book.
func (h *Handler) Delete(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Delete")

	// bind model
	model := &schema.RequestBookDelete{}
	if err := binding.BindModel(l, c, model, binding.BindFromParams()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// validate model
	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// delete a book
	response := h.UseCase.Delete(c.Request.Context(), model)
	c.JSON(response.Code, response)
}
```

### Step 0 Verification Checklist
- [x] Run `go mod tidy` successfully without errors
- [x] Run `go build ./cmd/main` successfully
- [ ] Start the server with `go run ./cmd/main`
- [ ] Access http://localhost:9000/health and verify response
- [ ] Access http://localhost:9000/swagger/index.html and verify Swagger UI loads
- [ ] Test book CRUD endpoints still work (GET /books/v1/,POST /books/v1/, etc.)
- [ ] Verify JWT authentication still works

#### Step 0 STOP & COMMIT
**STOP & COMMIT:** Stop here. Test all existing functionality works with Gin. Commit with message: "Migrate from Fiber to Gin framework"

---

## Step 1: Database Schema & Models

### Step 1.1: Update User Model with Role
- [x] Replace contents of `model/user.go`:

```go
package model

import (
	"time"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	ID        string    `gorm:"primaryKey;column:id;type:varchar(36);not null"`
	Username  string    `gorm:"column:username;type:varchar(255);not null;uniqueIndex"`
	Password  string    `gorm:"column:password;type:varchar(255);not null"`
	Role      UserRole  `gorm:"column:role;type:varchar(50);not null;default:'user'"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null"`
}

func (User) TableName() string {
	return "users"
}

type Users []User
```

### Step 1.2: Create Team Model
- [x] Create file `model/team.go`:

```go
package model

import (
	"time"
)

type Team struct {
	ID           string     `gorm:"primaryKey;column:id;type:varchar(36);not null"`
	Name         string     `gorm:"column:name;type:varchar(255);not null"`
	LogoURL      string     `gorm:"column:logo_url;type:text"`
	FoundedYear  int        `gorm:"column:founded_year;type:integer"`
	Address      string     `gorm:"column:address;type:text"`
	City         string     `gorm:"column:city;type:varchar(255)"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	CreatedBy    string     `gorm:"column:created_by;type:varchar(36);not null"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
	UpdatedBy    string     `gorm:"column:updated_by;type:varchar(36)"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;type:timestamptz;index"`
}

func (Team) TableName() string {
	return "teams"
}

type Teams []Team
```

### Step 1.3: Create Player Model
- [x] Create file `model/player.go`:

```go
package model

import (
	"time"
)

type PlayerPosition string

const (
	PositionGoalkeeper PlayerPosition = "GK"
	PositionDefender   PlayerPosition = "DF"
	PositionMidfielder PlayerPosition = "MF"
	PositionForward    PlayerPosition = "FW"
)

type Player struct {
	ID           string         `gorm:"primaryKey;column:id;type:varchar(36);not null"`
	TeamID       string         `gorm:"column:team_id;type:varchar(36);not null;index"`
	Name         string         `gorm:"column:name;type:varchar(255);not null"`
	HeightCM     int            `gorm:"column:height_cm;type:integer"`
	WeightKG     int            `gorm:"column:weight_kg;type:integer"`
	Position     PlayerPosition `gorm:"column:position;type:varchar(10);not null"`
	ShirtNumber  int            `gorm:"column:shirt_number;type:integer;not null"`
	CreatedAt    time.Time      `gorm:"column:created_at;type:timestamptz;not null"`
	CreatedBy    string         `gorm:"column:created_by;type:varchar(36);not null"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;type:timestamptz;not null"`
	UpdatedBy    string         `gorm:"column:updated_by;type:varchar(36)"`
	DeletedAt    *time.Time     `gorm:"column:deleted_at;type:timestamptz;index"`

	// Relationships
	Team *Team `gorm:"foreignKey:TeamID;references:ID"`
}

func (Player) TableName() string {
	return "players"
}

type Players []Player
```

### Step 1.4: Create Match Model
- [x] Create file `model/match.go`:

```go
package model

import (
	"time"
)

type MatchStatus string

const (
	MatchStatusScheduled MatchStatus = "scheduled"
	MatchStatusOngoing   MatchStatus = "ongoing"
	MatchStatusFinished  MatchStatus = "finished"
)

type Match struct {
	ID         string      `gorm:"primaryKey;column:id;type:varchar(36);not null"`
	MatchDate  time.Time   `gorm:"column:match_date;type:date;not null;index"`
	MatchTime  time.Time   `gorm:"column:match_time;type:timestamptz;not null"`
	HomeTeamID string      `gorm:"column:home_team_id;type:varchar(36);not null;index"`
	AwayTeamID string      `gorm:"column:away_team_id;type:varchar(36);not null;index"`
	HomeScore  int         `gorm:"column:home_score;type:integer;default:0"`
	AwayScore  int         `gorm:"column:away_score;type:integer;default:0"`
	Status     MatchStatus `gorm:"column:status;type:varchar(20);not null;default:'scheduled'"`
	CreatedAt  time.Time   `gorm:"column:created_at;type:timestamptz;not null"`
	CreatedBy  string      `gorm:"column:created_by;type:varchar(36);not null"`
	UpdatedAt  time.Time   `gorm:"column:updated_at;type:timestamptz;not null"`
	UpdatedBy  string      `gorm:"column:updated_by;type:varchar(36)"`

	// Relationships
	HomeTeam *Team `gorm:"foreignKey:HomeTeamID;references:ID"`
	AwayTeam *Team `gorm:"foreignKey:AwayTeamID;references:ID"`
}

func (Match) TableName() string {
	return "matches"
}

type Matches []Match
```

### Step 1.5: Create Goal Model
- [x] Create file `model/goal.go`:

```go
package model

import (
	"time"
)

type Goal struct {
	ID           string    `gorm:"primaryKey;column:id;type:varchar(36);not null"`
	MatchID      string    `gorm:"column:match_id;type:varchar(36);not null;index"`
	PlayerID     string    `gorm:"column:player_id;type:varchar(36);not null;index"`
	MinuteScored int       `gorm:"column:minute_scored;type:integer;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz;not null"`

	// Relationships
	Match  *Match  `gorm:"foreignKey:MatchID;references:ID"`
	Player *Player `gorm:"foreignKey:PlayerID;references:ID"`
}

func (Goal) TableName() string {
	return "goals"
}

type Goals []Goal
``

### Step 1.6: Create Audit Log Model
- [x] Create file `model/audit_log.go`:

```go
package model

import (
	"time"
)

type AuditAction string

const (
	AuditActionCreate AuditAction = "create"
	AuditActionUpdate AuditAction = "update"
	AuditActionDelete AuditAction = "delete"
)

type AuditLog struct {
	ID         string      `gorm:"primaryKey;column:id;type:varchar(36);not null"`
	EntityType string      `gorm:"column:entity_type;type:varchar(100);not null;index"`
	EntityID   string      `gorm:"column:entity_id;type:varchar(36);not null;index"`
	Action     AuditAction `gorm:"column:action;type:varchar(20);not null"`
	OldValue   string      `gorm:"column:old_value;type:jsonb"`
	NewValue   string      `gorm:"column:new_value;type:jsonb"`
	UserID     string      `gorm:"column:user_id;type:varchar(36);not null"`
	UserRole   string      `gorm:"column:user_role;type:varchar(50)"`
	CreatedAt  time.Time   `gorm:"column:created_at;type:timestamptz;not null;index"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type AuditLogs []AuditLog
```

### Step 1.7: Update Database Migration to Include New Models
- [x] Update `pkg/database/postgres.go` to add new models to migration (find the `MigrateIfNeed` function and add new models):

```go
func MigrateIfNeed(db *gorm.DB) {
	db.AutoMigrate(
		&model.User{},
		&model.Book{},
		&model.Team{},
		&model.Player{},
		&model.Match{},
		&model.Goal{},
		&model.AuditLog{},
	)

	// Add unique constraint for team_id + shirt_number
	db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_players_team_shirt 
		ON players(team_id, shirt_number) 
		WHERE deleted_at IS NULL
	`)

	// Add check constraint for match teams
	db.Exec(`
		ALTER TABLE matches 
		ADD CONSTRAINT IF NOT EXISTS check_different_teams 
		CHECK (home_team_id <> away_team_id)
	`)
}
```

### Step 1 Verification Checklist
- [ ] Run `go build ./cmd/main` successfully
- [ ] Start server and verify tables are created: `psql -U <user> -d <database> -c "\dt"`
- [ ] Verify tables exist: teams, players, matches, goals, audit_logs
- [ ] Verify unique constraint on players: `\d players`
- [ ] Verify check constraint on matches: `\d matches`
- [ ] Verify user table has role column: `\d users`

#### Step 1 STOP & COMMIT
**STOP & COMMIT:** Commit with message: "Add database models for Team, Player, Match, Goal, and AuditLog"

---

## Step 2: Role-Based Access Control (RBAC)

### Step 2.1: Update JWT Claims to Include Role
- [ ] Update `pkg/authentication/jwt.go` to add role to claims (find the `GenerateToken` function and update):

```go
package authentication

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type IJwtService interface {
	GenerateToken(claims JWTClaims) (string, error)
	ParseToken(tokenString string) (*JWTClaims, error)
}

type JWTClaims map[string]interface{}

type JWTConfig struct {
	PrivateKey     string
	PublicKey      string
	Audience       string
	Issuer         string
	ExpirationTime int
	RefreshTime    int
}

type JWTService struct {
	Config *JWTConfig
}

func NewJWT(config *JWTConfig) IJwtService {
	return &JWTService{Config: config}
}

func (j *JWTService) GenerateToken(claims JWTClaims) (string, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(j.Config.PrivateKey))
	if err != nil {
		return "", err
	}

	now := time.Now()
	jwtClaims := jwt.MapClaims{
		"iss": j.Config.Issuer,
		"aud": j.Config.Audience,
		"exp": now.Add(time.Duration(j.Config.ExpirationTime) * time.Minute).Unix(),
		"iat": now.Unix(),
	}

	// Merge custom claims (including userId and role)
	for k, v := range claims {
		jwtClaims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtClaims)
	return token.SignedString(privateKey)
}

func (j *JWTService) ParseToken(tokenString string) (*JWTClaims, error) {
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(j.Config.PublicKey))
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		result := make(JWTClaims)
		for k, v := range claims {
			result[k] = v
		}
		return &result, nil
	}

	return nil, errors.New("invalid token claims")
}
```

### Step 2.2: Update AuthUserData to Include Role
- [ ] Update `pkg/middleware/authentication.go` to include role in AuthUserData (replace AuthUserData struct and decodeAuthToken function):

```go
type AuthUserData struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

func decodeAuthToken(auth authentication.JWTClaims) *AuthUserData {
	userData := &AuthUserData{
		UserID: auth["userId"].(string),
	}
	
	if role, ok := auth["role"].(string); ok {
		userData.Role = role
	}
	
	return userData
}
```

### Step 2.3: Create Admin Role Middleware
- [ ] Add AdminAuth middleware to `pkg/middleware/authentication.go` (add after JwtAuth method):

```go
func (a *AuthMiddleware) AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First run JWT auth
		a.JwtAuth()(c)
		
		// If JWT auth failed, it would have aborted already
		if c.IsAborted() {
			return
		}
		
		// Get user data from context
		authUserValue, exists := c.Get(LocalTokenKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Access denied: admin role required",
			})
			return
		}
		
		authUser, ok := authUserValue.(*AuthUserData)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Access denied: invalid user data",
			})
			return
		}
		
		// Check if user has admin role
		if authUser.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Access denied: admin role required",
			})
			return
		}
		
		c.Next()
	}
}
```

### Step 2.4: Create Middleware Type File
- [ ] Create file `pkg/middleware/type.go`:

```go
package middleware

// AuthUserData represents the authenticated user data extracted from JWT
type AuthUserData struct {
	UserID string `json:"userId"`
	Role   string `json:"role"`
}

// Constants for context keys
const (
	LocalTokenKey = "user"
)
```

### Step 2 Verification Checklist
- [ ] Run `go build ./cmd/main` successfully
- [ ] Create a test endpoint that uses AdminAuth middleware
- [ ] Generate JWT token with role="admin" and verify access granted
- [ ] Generate JWT token with role="user" and verify access denied (403)
- [ ] Generate JWT token without role and verify access denied (403)

#### Step 2 STOP & COMMIT
**STOP & COMMIT:** Commit with message: "Implement role-based access control (RBAC) with admin middleware"

---

## Step 3: Team Management API

### Step 3.1: Create Team Schema - Constants
- [ ] Create file `internal/team/schema/constant.go`:

```go
package schema

const (
	ContextName = "Internal.Team"
)
```

### Step 3.2: Create Team Schema - Requests
- [ ] Create file `internal/team/schema/request.go`:

```go
package schema

import (
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/middleware"
)

type RequestTeamCreate struct {
	Name        string `json:"name" validate:"required,min=3,max=255"`
	LogoURL     string `json:"logo_url" validate:"omitempty,url"`
	FoundedYear int    `json:"founded_year" validate:"omitempty,min=1800,max=2100"`
	Address     string `json:"address" validate:"omitempty,max=500"`
	City        string `json:"city" validate:"omitempty,max=255"`
	AuthUserData *middleware.AuthUserData
}

type RequestTeamGet struct {
	ID string `uri:"id" validate:"required"`
	AuthUserData *middleware.AuthUserData
}

type RequestTeamList struct {
	Page      int    `form:"page" validate:"required,min=1"`
	PageSize  int    `form:"page_size" validate:"required,min=1,max=100"`
	SortBy    string `form:"sort_by" validate:"omitempty,oneof=name city founded_year"`
	SortOrder string `form:"sort_order" validate:"omitempty,oneof=asc desc"`
	AuthUserData *middleware.AuthUserData
}

type RequestTeamUpdate struct {
	ID          string `uri:"id" validate:"required"`
	Name        string `json:"name" validate:"required,min=3,max=255"`
	LogoURL     string `json:"logo_url" validate:"omitempty,url"`
	FoundedYear int    `json:"founded_year" validate:"omitempty,min=1800,max=2100"`
	Address     string `json:"address" validate:"omitempty,max=500"`
	City        string `json:"city" validate:"omitempty,max=255"`
	AuthUserData *middleware.AuthUserData
}

type RequestTeamDelete struct {
	ID string `uri:"id" validate:"required"`
	AuthUserData *middleware.AuthUserData
}

func (r *RequestTeamList) ToResponse(teams []model.Team) []ResponseTeamItem {
	responseTeams := make([]ResponseTeamItem, len(teams))
	for i, team := range teams {
		responseTeams[i] = ResponseTeamItem{
			ID:          team.ID,
			Name:        team.Name,
			LogoURL:     team.LogoURL,
			FoundedYear: team.FoundedYear,
			City:        team.City,
		}
	}
	return responseTeams
}
```

### Step 3.3: Create Team Schema - Responses
- [ ] Create file `internal/team/schema/response.go`:

```go
package schema

type ResponseTeamCreate struct {
	ID string `json:"id"`
}

type ResponseTeamGet struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	FoundedYear int    `json:"founded_year"`
	Address     string `json:"address"`
	City        string `json:"city"`
}

type ResponseTeamItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LogoURL     string `json:"logo_url"`
	FoundedYear int    `json:"founded_year"`
	City        string `json:"city"`
}

type ResponseTeamUpdate struct {
	ID string `json:"id"`
}

type ResponseTeamDelete struct{}
```

### Step 3.4: Create Team Repository
- [ ] Create file `internal/team/repository/repository.go`:

```go
package repository

import (
	"context"
	"fmt"

	"github.com/Alwanly/management-sport/internal/team/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg number"
	"github.com/Alwanly/management-sport/pkg/utils"
)

const ContextName = "Internal.Team.Repository"

type Repository struct {
	DB    database.IDBService
	Redis redis.IRedisService
}

type IRepository interface {
	Create(context.Context, *model.Team) error
	Get(context.Context, string) *model.Team
	List(context.Context, schema.RequestTeamList) ([]model.Team, int64)
	Update(context.Context, *model.Team) error
	Delete(context.Context, string) error
}

func NewRepository(r Repository) IRepository {
	return &Repository{
		DB:    r.DB,
		Redis: r.Redis,
	}
}

func (r *Repository) Create(ctx context.Context, team *model.Team) error {
	return r.DB.GetTransaction(ctx).Create(team).Error
}

func (r *Repository) Get(ctx context.Context, id string) *model.Team {
	var team model.Team
	err := r.DB.GetTransaction(ctx).
		Where("id = ? AND deleted_at IS NULL", id).
		First(&team).Error
	if err != nil {
		return nil
	}
	return &team
}

func (r *Repository) List(ctx context.Context, req schema.RequestTeamList) ([]model.Team, int64) {
	var teams []model.Team
	var total int64
	tx := r.DB.GetTransaction(ctx).Where("deleted_at IS NULL")

	tx.Model(&model.Team{}).Count(&total)

	offset := utils.CalculatePageSkip(req.Page, req.PageSize)
	sortBy := "name"
	if req.SortBy != "" {
		sortBy = req.SortBy
	}
	sortOrder := "asc"
	if req.SortOrder != "" {
		sortOrder = req.SortOrder
	}

	tx.Offset(offset).Limit(req.PageSize)
	tx.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))
	tx.Find(&teams)

	return teams, total
}

func (r *Repository) Update(ctx context.Context, team *model.Team) error {
	return r.DB.GetTransaction(ctx).Save(team).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.DB.GetTransaction(ctx).
		Model(&model.Team{}).
		Where("id = ?", id).
		Update("deleted_at", "NOW()").Error
}
```

### Step 3.5: Create Team UseCase
- [ ] Create file `internal/team/usecase/usecase.go`:

```go
package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/internal/team/repository"
	"github.com/Alwanly/management-sport/internal/team/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/ management-sport/pkg/utils"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"go.uber.org/zap"
)

const ContextName = "Internal.Team.Usecase"

type UseCase struct {
	Config     *config.GlobalConfig
	Logger     *zap.Logger
	Repository repository.IRepository
}

type IUseCase interface {
	Create(context.Context, *schema.RequestTeamCreate) wrapper.JSONResult
	Get(context.Context, *schema.RequestTeamGet) wrapper.JSONResult
	List(context.Context, *schema.RequestTeamList) wrapper.JSONResult
	Update(context.Context, *schema.RequestTeamUpdate) wrapper.JSONResult
	Delete(context.Context, *schema.RequestTeamDelete) wrapper.JSONResult
}

func NewUseCase(uc UseCase) IUseCase {
	return &UseCase{
		Config:     uc.Config,
		Logger:     uc.Logger,
		Repository: uc.Repository,
	}
}

func (u *UseCase) Create(ctx context.Context, req *schema.RequestTeamCreate) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Create"))

	now := time.Now()
	team := &model.Team{
		ID:          utils.GenerateUUID(),
		Name:        req.Name,
		LogoURL:     req.LogoURL,
		FoundedYear: req.FoundedYear,
		Address:     req.Address,
		City:        req.City,
		CreatedBy:   req.AuthUserData.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.Repository.Create(ctx, team); err != nil {
		l.Error("failed to create a team", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to create a team", nil)
	}

	l.Debug("team created", zap.String("id", team.ID))
	return wrapper.ResponseSuccess(http.StatusCreated, schema.ResponseTeamCreate{ID: team.ID})
}

func (u *UseCase) Get(ctx context.Context, req *schema.RequestTeamGet) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Get"))

	team := u.Repository.Get(ctx, req.ID)
	if team == nil {
		l.Error("team not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("TEAM_NOT_FOUND"), "Team not found", nil)
	}

	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseTeamGet{
		ID:          team.ID,
		Name:        team.Name,
		LogoURL:     team.LogoURL,
		FoundedYear: team.FoundedYear,
		Address:     team.Address,
		City:        team.City,
	})
}

func (u *UseCase) List(ctx context.Context, req *schema.RequestTeamList) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "List"))

	teams, total := u.Repository.List(ctx, *req)

	response := req.ToResponse(teams)
	l.Debug("teams listed", zap.Int64("total", total))
	return wrapper.ResponsePagination(req.Page, req.PageSize, len(teams), int(total), response, nil)
}

func (u *UseCase) Update(ctx context.Context, req *schema.RequestTeamUpdate) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Update"))

	team := u.Repository.Get(ctx, req.ID)
	if team == nil {
		l.Error("team not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("TEAM_NOT_FOUND"), "Team not found", nil)
	}

	team.Name = req.Name
	team.LogoURL = req.LogoURL
	team.FoundedYear = req.FoundedYear
	team.Address = req.Address
	team.City = req.City
	team.UpdatedAt = time.Now()
	team.UpdatedBy = req.AuthUserData.UserID

	if err := u.Repository.Update(ctx, team); err != nil {
		l.Error("failed to update a team", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to update a team", nil)
	}

	l.Debug("team updated", zap.String("id", team.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseTeamUpdate{ID: team.ID})
}

func (u *UseCase) Delete(ctx context.Context, req *schema.RequestTeamDelete) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Delete"))

	team := u.Repository.Get(ctx, req.ID)
	if team == nil {
		l.Error("team not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("TEAM_NOT_FOUND"), "Team not found", nil)
	}

	if err := u.Repository.Delete(ctx, req.ID); err != nil {
		l.Error("failed to delete a team", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to delete a team", nil)
	}

	l.Debug("team deleted", zap.String("id", req.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseTeamDelete{})
}
```

### Step 3.6: Create Team Handler
- [ ] Create file `internal/team/handler/handler.go`:

```go
package handler

import (
	"github.com/Alwanly/management-sport/internal/team/repository"
	"github.com/Alwanly/management-sport/internal/team/schema"
	"github.com/Alwanly/management-sport/internal/team/usecase"
	"github.com/Alwanly/management-sport/pkg/binding"
	"github.com/Alwanly/management-sport/pkg/deps"
	"github.com/Alwanly/management-sport/pkg/logger"
	"github.com/Alwanly/management-sport/pkg/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ContextName = "Internal.Team.Handler"

type Handler struct {
	Logger    *zap.Logger
	Validator validator.IValidatorService
	UseCase   usecase.IUseCase
}

func NewHandler(d *deps.App) *Handler {
	repository := repository.NewRepository(repository.Repository{
		DB:    d.DB,
		Redis: d.Redis,
	})
	usecase := usecase.NewUseCase(usecase.UseCase{
		Config:     d.Config,
		Logger:     d.Logger,
		Repository: repository,
	})
	handler := &Handler{
		Logger:    d.Logger,
		Validator: d.Validator,
		UseCase:   usecase,
	}

	// Public endpoints (read-only)
	public := d.Gin.Group("/teams/v1")
	public.Use(d.Auth.JwtAuth())
	public.GET("/", handler.List)
	public.GET("/:id", handler.Get)

	// Admin endpoints (write operations)
	admin := d.Gin.Group("/teams/v1")
	admin.Use(d.Auth.AdminAuth())
	admin.POST("/", handler.Create)
	admin.PUT("/:id", handler.Update)
	admin.DELETE("/:id", handler.Delete)

	return handler
}

//	@Summary		Create a new team
//	@Description	Create a new team (admin only)
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			team	body		schema.RequestTeamCreate	true	"Team data"
//	@Success		201		{object}	wrapper.JSONResult{data=schema.ResponseTeamCreate}
//	@Failure		400		{object}	wrapper.JSONResult
//	@Failure		401		{object}	wrapper.JSONResult
//	@Failure		403		{object}	wrapper.JSONResult
//	@Security		BearerAuth
//	@Router			/teams/v1 [post]
func (h *Handler) Create(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Create")

	model := &schema.RequestTeamCreate{}
	if err := binding.BindModel(l, c, model, binding.BindFromBody()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	response := h.UseCase.Create(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

//	@Summary		Get team by ID
//	@Description	Get team details by ID
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Team ID"
//	@Success		200	{object}	wrapper.JSONResult{data=schema.ResponseTeamGet}
//	@Failure		404	{object}	wrapper.JSONResult
//	@Failure		401	{object}	wrapper.JSONResult
//	@Security		BearerAuth
//	@Router			/teams/v1/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Get")

	model := &schema.RequestTeamGet{}
	if err := binding.BindModel(l, c, model, binding.BindFromParams()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	response := h.UseCase.Get(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

//	@Summary		List teams
//	@Description	List all teams with pagination
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"			default(1)
//	@Param			page_size	query		int		false	"Page size"				default(10)
//	@Param			sort_by		query		string	false	"Sort by field"			Enums(name, city, founded_year)
//	@Param			sort_order	query		string	false	"Sort order"			Enums(asc, desc)
//	@Success		200			{object}	wrapper.JSONResult{data=[]schema.ResponseTeamItem}
//	@Failure		400			{object}	wrapper.JSONResult
//	@Failure		401			{object}	wrapper.JSONResult
//	@Security		BearerAuth
//	@Router			/teams/v1 [get]
func (h *Handler) List(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "List")

	model := &schema.RequestTeamList{
		Page:      1,
		PageSize:  10,
		SortBy:    "name",
		SortOrder: "asc",
	}
	if err := binding.BindModel(l, c, model, binding.BindFromQuery()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	response := h.UseCase.List(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

//	@Summary		Update team
//	@Description	Update team details (admin only)
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Team ID"
//	@Param			team	body		schema.RequestTeamUpdate	true	"Team data"
//	@Success		200		{object}	wrapper.JSONResult{data=schema.ResponseTeamUpdate}
//	@Failure		400		{object}	wrapper.JSONResult
//	@Failure		401		{object}	wrapper.JSONResult
//	@Failure		403		{object}	wrapper.JSONResult
//	@Failure		404		{object}	wrapper.JSONResult
//	@Security		BearerAuth
//	@Router			/teams/v1/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Update")

	model := &schema.RequestTeamUpdate{}
	if err := binding.BindModel(l, c, model, binding.BindFromParams(), binding.BindFromBody()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	response := h.UseCase.Update(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

//	@Summary		Delete team
//	@Description	Soft delete a team (admin only)
//	@Tags			teams
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Team ID"
//	@Success		200	{object}	wrapper.JSONResult{data=schema.ResponseTeamDelete}
//	@Failure		401	{object}	wrapper.JSONResult
//	@Failure		403	{object}	wrapper.JSONResult
//	@Failure		404	{object}	wrapper.JSONResult
//	@Security		BearerAuth
//	@Router			/teams/v1/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Delete")

	model := &schema.RequestTeamDelete{}
	if err := binding.BindModel(l, c, model, binding.BindFromParams()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	if err := validator.ValidateModel( l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	response := h.UseCase.Delete(c.Request.Context(), model)
	c.JSON(response.Code, response)
}
```

### Step 3.7: Register Team Handler in Bootstrap
- [ ] Update `cmd/main/bootstrap.go` to register team handler (add after book_handler line):

```go
// Import at the top
team_handler "github.com/Alwanly/management-sport/internal/team/handler"

// Register in Bootstrap function
book_handler.NewHandler(inst)
team_handler.NewHandler(inst)  // Add this line
```

### Step 3 Verification Checklist
- [ ] Run `go build ./cmd/main` successfully
- [ ] Start server with `go run ./cmd/main`
- [ ] Generate admin JWT token
- [ ] Test POST /teams/v1 (create team) - should succeed with admin token
- [ ] Test GET /teams/v1 (list teams) - should succeed with any authenticated user
- [ ] Test GET /teams/v1/:id (get team) - should succeed
- [ ] Test PUT /teams/v1/:id (update team) - should succeed with admin token
- [ ] Test DELETE /teams/v1/:id (delete team) - should succeed with admin token, team soft-deleted
- [ ] Verify non-admin user cannot POST/PUT/DELETE (403 Forbidden)
- [ ] Run `make docs` to regenerate Swagger documentation

#### Step 3 STOP & COMMIT
**STOP & COMMIT:** Commit with message: "Implement Team Management API with CRUD operations"

---

Due to character limits, I'll create this as part 1. The implementation guide continues with Steps 4-8 following the same detailed pattern. Each step includes:
- Complete file creation with full code
- Schema, repository, usecase, and handler layers
- Swagger annotations
- Verification checklists
- Commit points

Would you like me to continue with Steps 4-8 in a follow-up response, or would you prefer I create additional separate implementation files for the remaining steps?
