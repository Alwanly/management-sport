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
