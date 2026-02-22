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
	public.GET(":id", handler.Get)

	// Admin endpoints (write operations)
	admin := d.Gin.Group("/teams/v1")
	admin.Use(d.Auth.AdminAuth())
	admin.POST("/", handler.Create)
	admin.PUT(":id", handler.Update)
	admin.DELETE(":id", handler.Delete)

	return handler
}

// Create godoc
// @Summary      Create a new team
// @Description  Create a new team (admin only)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Param        request body schema.RequestTeamCreate true "Create team"
// @Success      201 {object} wrapper.JSONResult{data=schema.ResponseTeamCreate}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /teams/v1 [post]
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

// Get godoc
// @Summary      Get team by id
// @Description  Retrieve a single team
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Param        id path string true "Team ID"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponseTeamGet}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /teams/v1/{id} [get]
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

// List godoc
// @Summary      List teams
// @Description  List teams with pagination
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number"
// @Param        page_size query int false "Page size"
// @Param        sort_by query string false "Sort by field"
// @Param        sort_order query string false "Sort order (asc|desc)"
// @Success      200 {object} wrapper.JSONResult{data=[]schema.ResponseTeamItem}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /teams/v1 [get]
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

// Update godoc
// @Summary      Update a team
// @Description  Update a team's information (admin only)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Param        id path string true "Team ID"
// @Param        request body schema.RequestTeamUpdate true "Update team"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponseTeamUpdate}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /teams/v1/{id} [put]
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

// Delete godoc
// @Summary      Delete a team
// @Description  Delete a team (admin only)
// @Tags         Teams
// @Accept       json
// @Produce      json
// @Param        id path string true "Team ID"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponseTeamDelete}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /teams/v1/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Delete")

	model := &schema.RequestTeamDelete{}
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
