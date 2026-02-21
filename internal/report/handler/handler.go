package handler

import (
	"github.com/Alwanly/management-sport/internal/report/repository"
	"github.com/Alwanly/management-sport/internal/report/usecase"
	"github.com/Alwanly/management-sport/pkg/deps"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ContextName = "Internal.Report.Handler"

type Handler struct {
	Logger  *zap.Logger
	UseCase usecase.IUseCase
}

func NewHandler(d *deps.App) *Handler {
	repo := repository.NewRepository(d.DB)
	uc := usecase.NewUseCase(usecase.UseCase{Config: d.Config, Logger: d.Logger, Repository: repo})
	h := &Handler{Logger: d.Logger, UseCase: uc}

	group := d.Gin.Group("/reports/v1")
	group.Use(d.Auth.JwtAuth())
	group.GET("/goals-per-player", func(c *gin.Context) {
		resp := h.UseCase.GoalsPerPlayer(c.Request.Context())
		c.JSON(resp.Code, resp)
	})
	group.GET("/team-goals", func(c *gin.Context) {
		resp := h.UseCase.TeamGoals(c.Request.Context())
		c.JSON(resp.Code, resp)
	})

	return h
}
