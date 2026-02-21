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
