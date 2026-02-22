package handler

import (
	"github.com/Alwanly/management-sport/internal/goal/repository"
	"github.com/Alwanly/management-sport/internal/goal/schema"
	"github.com/Alwanly/management-sport/internal/goal/usecase"
	"github.com/Alwanly/management-sport/pkg/binding"
	"github.com/Alwanly/management-sport/pkg/deps"
	"github.com/Alwanly/management-sport/pkg/logger"
	"github.com/Alwanly/management-sport/pkg/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ContextName = "Internal.Goal.Handler"

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

	// Public endpoints
	public := d.Gin.Group("/goals/v1")
	public.Use(d.Auth.JwtAuth())
	public.GET("/", handler.List)
	public.GET(":id", handler.Get)

	// Admin endpoints
	admin := d.Gin.Group("/goals/v1")
	admin.Use(d.Auth.AdminAuth())
	admin.POST("/", handler.Create)
	admin.DELETE(":id", handler.Delete)

	return handler
}

// Create godoc
// @Summary      Create a new goal
// @Description  Create a new goal (admin only)
// @Tags         Goals
// @Accept       json
// @Produce      json
// @Param        request body schema.RequestGoalCreate true "Create goal"
// @Success      201 {object} wrapper.JSONResult{data=schema.ResponseGoalCreate}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /goals/v1 [post]
func (h *Handler) Create(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Create")

	model := &schema.RequestGoalCreate{}
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

// Get godoc
// @Summary      Get goal by id
// @Description  Retrieve a single goal
// @Tags         Goals
// @Accept       json
// @Produce      json
// @Param        id path string true "Goal ID"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponseGoalGet}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /goals/v1/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Get")

	model := &schema.RequestGoalGet{}
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

// List godoc
// @Summary      List goals
// @Description  List goals with pagination
// @Tags         Goals
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number"
// @Param        page_size query int false "Page size"
// @Param        sort_by query string false "Sort by field"
// @Param        sort_order query string false "Sort order (asc|desc)"
// @Success      200 {object} wrapper.JSONResult{data=[]schema.ResponseGoalItem}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /goals/v1 [get]
func (h *Handler) List(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "List")

	model := &schema.RequestGoalList{
		Page:      1,
		PageSize:  10,
		SortBy:    "created_at",
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

// Delete godoc
// @Summary      Delete a goal
// @Description  Delete a goal (admin only)
// @Tags         Goals
// @Accept       json
// @Produce      json
// @Param        id path string true "Goal ID"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponseGoalDelete}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /goals/v1/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Delete")

	model := &schema.RequestGoalDelete{}
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

	response := h.UseCase.Delete(c.Request.Context(), model)
	c.JSON(response.Code, response)
}
