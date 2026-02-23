package handler

import (
	"context"

	"github.com/Alwanly/management-sport/internal/auth/repository"
	"github.com/Alwanly/management-sport/internal/auth/schema"
	"github.com/Alwanly/management-sport/internal/auth/usecase"
	"github.com/Alwanly/management-sport/pkg/binding"
	"github.com/Alwanly/management-sport/pkg/deps"
	"github.com/Alwanly/management-sport/pkg/logger"
	"github.com/Alwanly/management-sport/pkg/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ContextName = "Internal.Auth.Handler"

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
		JWT:        d.Auth.Jwt,
	})

	handler := &Handler{
		Logger:    d.Logger,
		Validator: d.Validator,
		UseCase:   usecase,
	}

	// Ensure default admin exists on startup
	ctx := context.Background()
	if err := usecase.EnsureDefaultAdmin(ctx); err != nil {
		d.Logger.Error("Failed to ensure default admin", zap.Error(err))
	}

	// Public authentication endpoints
	auth := d.Gin.Group("/auth/v1")
	auth.Use(d.Auth.BasicAuth()) // Allow access without JWT for login and registration
	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Login)
	auth.POST("/admin/login", handler.AdminLogin)

	return handler
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return JWT token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body schema.RequestLogin true "Login credentials"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponseLogin}
// @Failure      400 {object} wrapper.JSONResult
// @Failure      401 {object} wrapper.JSONResult
// @Failure      500 {object} wrapper.JSONResult
// @Security     BasicAuth
// @Router       /auth/v1/login [post]
func (h *Handler) Login(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Login")

	model := &schema.RequestLogin{}
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

	response := h.UseCase.Login(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

// AdminLogin godoc
// @Summary      Admin login
// @Description  Authenticate admin user and return JWT token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body schema.RequestAdminLogin true "Admin login credentials"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponseAdminLogin}
// @Failure      400 {object} wrapper.JSONResult
// @Failure      401 {object} wrapper.JSONResult
// @Failure      403 {object} wrapper.JSONResult
// @Failure      500 {object} wrapper.JSONResult
// @Security     BasicAuth
// @Router       /auth/v1/admin/login [post]
func (h *Handler) AdminLogin(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "AdminLogin")

	model := &schema.RequestAdminLogin{}
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

	response := h.UseCase.AdminLogin(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

// Register godoc
// @Summary      Register new user
// @Description  Create a new user account with username and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body schema.RequestRegister true "Registration details"
// @Success      201 {object} wrapper.JSONResult{data=schema.ResponseRegister}
// @Failure      400 {object} wrapper.JSONResult
// @Failure      409 {object} wrapper.JSONResult
// @Failure      500 {object} wrapper.JSONResult
// @Security     BasicAuth
// @Router       /auth/v1/register [post]
func (h *Handler) Register(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Register")

	model := &schema.RequestRegister{}
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

	response := h.UseCase.Register(c.Request.Context(), model)
	c.JSON(response.Code, response)
}
