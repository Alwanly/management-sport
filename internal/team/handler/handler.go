package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Alwanly/management-sport/internal/team/repository"
	"github.com/Alwanly/management-sport/internal/team/schema"
	"github.com/Alwanly/management-sport/internal/team/usecase"
	"github.com/Alwanly/management-sport/pkg/binding"
	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/deps"
	"github.com/Alwanly/management-sport/pkg/logger"
	"github.com/Alwanly/management-sport/pkg/middleware"
	"github.com/Alwanly/management-sport/pkg/validator"
	"github.com/Alwanly/management-sport/pkg/wrapper"
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
	admin.Use(d.Auth.JwtAuth(), d.Auth.AdminAuth())
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
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		l.Error("failed to parse multipart form", zap.Error(err))
		result := wrapper.ResponseFailed(
			http.StatusBadRequest,
			contract.StatusCodeBindingFailed,
			"Failed to parse form data",
			nil,
		)
		c.JSON(result.Code, result)
		return
	}

	// Bind form fields to model
	model := &schema.RequestTeamCreate{}
	if err := c.ShouldBind(model); err != nil {
		l.Error("failed to bind form data", zap.Error(err))
		result := wrapper.ResponseFailed(
			http.StatusBadRequest,
			contract.StatusCodeBindingFailed,
			contract.ErrorValidatePayload,
			nil,
		)
		c.JSON(result.Code, result)
		return
	}

	// Get auth data from context
	if authUserValue, exists := c.Get(middleware.LocalTokenKey); exists {
		if authUser, ok := authUserValue.(*middleware.AuthUserData); ok {
			model.AuthUserData = authUser
		}
	}

	// Validate model
	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// Handle file upload (optional)
	var logoURL string
	file, err := c.FormFile("logo")
	if err == nil {
		// File was provided, process it
		logoURL, err = h.UseCase.ProcessLogoUpload(model.Name, file)
		if err != nil {
			l.Error("failed to process logo upload", zap.Error(err))
			result := wrapper.ResponseFailed(
				http.StatusBadRequest,
				contract.StatusCodeValidationFailed,
				err.Error(),
				nil,
			)
			c.JSON(result.Code, result)
			return
		}
		l.Info("logo uploaded", zap.String("url", logoURL))
	}

	// Create team
	response := h.UseCase.Create(c.Request.Context(), model)

	// If team created successfully and logo was uploaded, update the logo URL
	if response.StatusCode == contract.StatusCodeSuccess && logoURL != "" {
		if createResp, ok := response.Data.(schema.ResponseTeamCreate); ok {
			// Update the team's logo URL in database
			teamToUpdate := h.UseCase.Get(c.Request.Context(), &schema.RequestTeamGet{
				ID:           createResp.ID,
				AuthUserData: model.AuthUserData,
			})

			if teamToUpdate.StatusCode == contract.StatusCodeSuccess {
				// Update with logo path - update directly
				h.updateTeamLogo(c.Request.Context(), createResp.ID, logoURL, model.AuthUserData.UserID)
			}
		}
	}

	c.JSON(response.Code, response)
}

func (h *Handler) updateTeamLogo(ctx context.Context, teamID string, logoURL string, userID string) {
	team := h.UseCase.(*usecase.UseCase).Repository.Get(ctx, teamID)
	if team != nil {
		team.LogoURL = logoURL
		team.UpdatedBy = userID
		team.UpdatedAt = time.Now()
		h.UseCase.(*usecase.UseCase).Repository.Update(ctx, team)
	}
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

	// Get team ID from path
	teamID := c.Param("id")
	if teamID == "" {
		result := wrapper.ResponseFailed(
			http.StatusBadRequest,
			contract.StatusCodeBindingFailed,
			"Team ID is required",
			nil,
		)
		c.JSON(result.Code, result)
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		l.Error("failed to parse multipart form", zap.Error(err))
		result := wrapper.ResponseFailed(
			http.StatusBadRequest,
			contract.StatusCodeBindingFailed,
			"Failed to parse form data",
			nil,
		)
		c.JSON(result.Code, result)
		return
	}

	// Bind form fields to model
	model := &schema.RequestTeamUpdate{}
	if err := c.ShouldBind(model); err != nil {
		l.Error("failed to bind form data", zap.Error(err))
		result := wrapper.ResponseFailed(
			http.StatusBadRequest,
			contract.StatusCodeBindingFailed,
			contract.ErrorValidatePayload,
			nil,
		)
		c.JSON(result.Code, result)
		return
	}

	// Set ID from path parameter
	model.ID = teamID

	// Get auth data from context
	if authUserValue, exists := c.Get(middleware.LocalTokenKey); exists {
		if authUser, ok := authUserValue.(*middleware.AuthUserData); ok {
			model.AuthUserData = authUser
		}
	}

	// Validate model
	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// Get existing team to check for old logo
	existingTeam := h.UseCase.Get(c.Request.Context(), &schema.RequestTeamGet{
		ID:           teamID,
		AuthUserData: model.AuthUserData,
	})

	if existingTeam.StatusCode != contract.StatusCodeSuccess {
		c.JSON(existingTeam.Code, existingTeam)
		return
	}

	var oldLogoURL string
	if teamData, ok := existingTeam.Data.(schema.ResponseTeamGet); ok {
		oldLogoURL = teamData.LogoURL
	}

	// Handle file upload (optional)
	var newLogoURL string
	file, err := c.FormFile("logo")
	if err == nil {
		// File was provided, process it
		newLogoURL, err = h.UseCase.ProcessLogoUpload(model.Name, file)
		if err != nil {
			l.Error("failed to process logo upload", zap.Error(err))
			result := wrapper.ResponseFailed(
				http.StatusBadRequest,
				contract.StatusCodeValidationFailed,
				err.Error(),
				nil,
			)
			c.JSON(result.Code, result)
			return
		}
		l.Info("logo uploaded", zap.String("url", newLogoURL))

		// Delete old logo if it exists and is different
		if oldLogoURL != "" && oldLogoURL != newLogoURL {
			if err := h.UseCase.DeleteOldLogo(oldLogoURL); err != nil {
				l.Warn("failed to delete old logo", zap.Error(err))
				// Continue anyway, don't fail the update
			}
		}

		// Update team logo in database directly
		h.updateTeamLogo(c.Request.Context(), teamID, newLogoURL, model.AuthUserData.UserID)
	}

	// Update team
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
