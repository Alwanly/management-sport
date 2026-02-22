package handler

import (
	"github.com/Alwanly/management-sport/internal/audit/repository"
	"github.com/Alwanly/management-sport/internal/audit/schema"
	"github.com/Alwanly/management-sport/internal/audit/usecase"
	"github.com/Alwanly/management-sport/pkg/binding"
	"github.com/Alwanly/management-sport/pkg/deps"
	"github.com/Alwanly/management-sport/pkg/logger"
	"github.com/Alwanly/management-sport/pkg/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ContextName = "Internal.Audit.Handler"

type Handler struct {
	Logger    *zap.Logger
	Validator validator.IValidatorService
	UseCase   usecase.IUseCase
}

func NewHandler(d *deps.App) *Handler {
	repo := repository.NewRepository(repository.Repository{DB: d.DB, Redis: d.Redis})
	uc := usecase.NewUseCase(usecase.UseCase{Config: d.Config, Logger: d.Logger, Repository: repo})
	h := &Handler{Logger: d.Logger, Validator: d.Validator, UseCase: uc}

	// Admin-only audit endpoints
	admin := d.Gin.Group("/audits/v1")
	admin.Use(d.Auth.JwtAuth(), d.Auth.AdminAuth())
	admin.GET("/", h.List)
	admin.GET(":id", h.Get)

	return h
}

// List godoc
// @Summary      List audit logs
// @Description  List audit logs (admin only)
// @Tags         Audits
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number"
// @Param        page_size query int false "Page size"
// @Success      200 {object} wrapper.JSONResult{data=[]schema.ResponseAuditItem}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /audits/v1 [get]
func (h *Handler) List(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "List")
	model := &schema.RequestAuditList{Page: 1, PageSize: 10}
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
	resp := h.UseCase.List(c.Request.Context(), model)
	c.JSON(resp.Code, resp)
}

// Get godoc
// @Summary      Get audit by id
// @Description  Retrieve a single audit log (admin only)
// @Tags         Audits
// @Accept       json
// @Produce      json
// @Param        id path string true "Audit ID"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponseAuditGet}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /audits/v1/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Get")
	model := &schema.RequestAuditGet{}
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
	resp := h.UseCase.Get(c.Request.Context(), model)
	c.JSON(resp.Code, resp)
}
