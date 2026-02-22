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
	group.GET("/goals-per-player", h.GoalsPerPlayer)
	group.GET("/team-goals", h.TeamGoals)
	group.GET("/matches", h.MatchReport)

	return h
}

// GoalsPerPlayer godoc
// @Summary      Goals per player
// @Description  Returns number of goals per player
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Success      200 {object} wrapper.JSONResult{data=[]schema.ResponseGoalsPerPlayerItem}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /reports/v1/goals-per-player [get]
func (h *Handler) GoalsPerPlayer(c *gin.Context) {
	resp := h.UseCase.GoalsPerPlayer(c.Request.Context())
	c.JSON(resp.Code, resp)
}

// TeamGoals godoc
// @Summary      Goals per team
// @Description  Returns number of goals per team
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Success      200 {object} wrapper.JSONResult{data=[]schema.ResponseTeamGoalsItem}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /reports/v1/team-goals [get]
func (h *Handler) TeamGoals(c *gin.Context) {
	resp := h.UseCase.TeamGoals(c.Request.Context())
	c.JSON(resp.Code, resp)
}

// MatchReport godoc
// @Summary      Match report
// @Description  Returns finished matches with schedule, teams, scores, and results
// @Tags         Reports
// @Accept       json
// @Produce      json
// @Success      200 {object} wrapper.JSONResult{data=[]schema.ResponseMatchReportItem}
// @Failure      400 {object} wrapper.JSONResult
// @Failure      500 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /reports/v1/matches [get]
func (h *Handler) MatchReport(c *gin.Context) {
	resp := h.UseCase.MatchReport(c.Request.Context())
	c.JSON(resp.Code, resp)
}
