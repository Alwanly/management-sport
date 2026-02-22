package handler

import (
	"github.com/Alwanly/management-sport/internal/player/repository"
	"github.com/Alwanly/management-sport/internal/player/schema"
	"github.com/Alwanly/management-sport/internal/player/usecase"
	"github.com/Alwanly/management-sport/pkg/binding"
	"github.com/Alwanly/management-sport/pkg/deps"
	"github.com/Alwanly/management-sport/pkg/logger"
	"github.com/Alwanly/management-sport/pkg/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ContextName = "Internal.Player.Handler"

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
	public := d.Gin.Group("/players/v1")
	public.Use(d.Auth.JwtAuth())
	public.GET("/", handler.List)
	public.GET(":id", handler.Get)

	// Admin endpoints
	admin := d.Gin.Group("/players/v1")
	admin.Use(d.Auth.JwtAuth(), d.Auth.AdminAuth())
	admin.POST("/", handler.Create)
	admin.PUT(":id", handler.Update)
	admin.DELETE(":id", handler.Delete)

	return handler
}

// Create godoc
// @Summary      Create a new player
// @Description  Create a new player (admin only)
// @Tags         Players
// @Accept       json
// @Produce      json
// @Param        request body schema.RequestPlayerCreate true "Create player"
// @Success      201 {object} wrapper.JSONResult{data=schema.ResponsePlayerCreate}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /players/v1 [post]
func (h *Handler) Create(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Create")

	model := &schema.RequestPlayerCreate{}
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
// @Summary      Get player by id
// @Description  Retrieve a single player
// @Tags         Players
// @Accept       json
// @Produce      json
// @Param        id path string true "Player ID"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponsePlayerGet}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /players/v1/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Get")

	model := &schema.RequestPlayerGet{}
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
// @Summary      List players
// @Description  List players with pagination
// @Tags         Players
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number"
// @Param        page_size query int false "Page size"
// @Param        sort_by query string false "Sort by field"
// @Param        sort_order query string false "Sort order (asc|desc)"
// @Success      200 {object} wrapper.JSONResult{data=[]schema.ResponsePlayerItem}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /players/v1 [get]
func (h *Handler) List(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "List")

	model := &schema.RequestPlayerList{
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

// Update godoc
// @Summary      Update a player
// @Description  Update a player's information (admin only)
// @Tags         Players
// @Accept       json
// @Produce      json
// @Param        id path string true "Player ID"
// @Param        request body schema.RequestPlayerUpdate true "Update player"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponsePlayerUpdate}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /players/v1/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Update")

	model := &schema.RequestPlayerUpdate{}
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

// Delete godoc
// @Summary      Delete a player
// @Description  Delete a player (admin only)
// @Tags         Players
// @Accept       json
// @Produce      json
// @Param        id path string true "Player ID"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponsePlayerDelete}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /players/v1/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Delete")

	model := &schema.RequestPlayerDelete{}
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
