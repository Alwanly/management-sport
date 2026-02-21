package deps

import (
	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg/middleware"
	"github.com/Alwanly/management-sport/pkg/redis"
	"github.com/Alwanly/management-sport/pkg/validator"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type App struct {
	Config    *config.GlobalConfig
	Logger    *zap.Logger
	DB        *database.DBService
	Redis     *redis.Service
	Auth      *middleware.AuthMiddleware
	Validator validator.IValidatorService

	// APIs
	Fiber *fiber.App
}
