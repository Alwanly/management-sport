package usecase

import (
	"context"
	"net/http"
	"time"

	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/internal/auth/repository"
	"github.com/Alwanly/management-sport/internal/auth/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/authentication"
	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/utils"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"go.uber.org/zap"
)

const ContextName = "Internal.Auth.Usecase"

type UseCase struct {
	Config     *config.GlobalConfig
	Logger     *zap.Logger
	Repository repository.IRepository
	JWT        authentication.IJwtService
}

type IUseCase interface {
	Login(context.Context, *schema.RequestLogin) wrapper.JSONResult
	AdminLogin(context.Context, *schema.RequestAdminLogin) wrapper.JSONResult
	Register(context.Context, *schema.RequestRegister) wrapper.JSONResult
	EnsureDefaultAdmin(context.Context) error
}

func NewUseCase(uc UseCase) IUseCase {
	return &UseCase{
		Config:     uc.Config,
		Logger:     uc.Logger,
		Repository: uc.Repository,
		JWT:        uc.JWT,
	}
}

func (u *UseCase) Login(ctx context.Context, req *schema.RequestLogin) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Login"))

	user := u.Repository.GetByUsername(ctx, req.Username)
	if user == nil {
		l.Debug("user not found", zap.String("username", req.Username))
		return wrapper.ResponseFailed(
			http.StatusUnauthorized,
			contract.StatusCodeUserOrPasswordInvalid,
			"Invalid username or password",
			nil,
		)
	}

	if !authentication.VerifyPassword(req.Password, user.Password) {
		l.Debug("invalid password", zap.String("username", req.Username))
		return wrapper.ResponseFailed(
			http.StatusUnauthorized,
			contract.StatusCodeUserOrPasswordInvalid,
			"Invalid username or password",
			nil,
		)
	}

	if user.Role != model.RoleUser {
		l.Debug("non-user attempting regular login", zap.String("username", req.Username), zap.String("role", string(user.Role)))
		return wrapper.ResponseFailed(
			http.StatusForbidden,
			contract.StatusCodeUnauthorized,
			"Access denied",
			nil,
		)
	}

	claims := authentication.JWTClaims{
		"userId": user.ID,
		"role":   string(user.Role),
	}

	token, err := u.JWT.GenerateToken(claims)
	if err != nil {
		l.Error("failed to create token", zap.Error(err))
		return wrapper.ResponseFailed(
			http.StatusInternalServerError,
			contract.StatusCodeInternalServerError,
			"Failed to create token",
			nil,
		)
	}

	l.Debug("user logged in", zap.String("userId", user.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseLogin{
		Token:  token,
		UserID: user.ID,
		Role:   string(user.Role),
	})
}

func (u *UseCase) AdminLogin(ctx context.Context, req *schema.RequestAdminLogin) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "AdminLogin"))

	user := u.Repository.GetByUsername(ctx, req.Username)
	if user == nil {
		l.Debug("admin not found", zap.String("username", req.Username))
		return wrapper.ResponseFailed(
			http.StatusUnauthorized,
			contract.StatusCodeUserOrPasswordInvalid,
			"Invalid username or password",
			nil,
		)
	}

	if !authentication.VerifyPassword(req.Password, user.Password) {
		l.Debug("invalid password", zap.String("username", req.Username))
		return wrapper.ResponseFailed(
			http.StatusUnauthorized,
			contract.StatusCodeUserOrPasswordInvalid,
			"Invalid username or password",
			nil,
		)
	}

	if user.Role != model.RoleAdmin {
		l.Debug("non-admin attempting admin login", zap.String("username", req.Username), zap.String("role", string(user.Role)))
		return wrapper.ResponseFailed(
			http.StatusForbidden,
			contract.StatusCodeUnauthorized,
			"Access denied: admin role required",
			nil,
		)
	}

	claims := authentication.JWTClaims{
		"userId": user.ID,
		"role":   string(user.Role),
	}

	token, err := u.JWT.GenerateToken(claims)
	if err != nil {
		l.Error("failed to create token", zap.Error(err))
		return wrapper.ResponseFailed(
			http.StatusInternalServerError,
			contract.StatusCodeInternalServerError,
			"Failed to create token",
			nil,
		)
	}

	l.Debug("admin logged in", zap.String("userId", user.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseAdminLogin{
		Token:  token,
		UserID: user.ID,
		Role:   string(user.Role),
	})
}

func (u *UseCase) Register(ctx context.Context, req *schema.RequestRegister) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Register"))

	if u.Repository.ExistsByUsername(ctx, req.Username) {
		l.Debug("username already exists", zap.String("username", req.Username))
		return wrapper.ResponseFailed(
			http.StatusConflict,
			contract.CreateStatusCode("000014"),
			"Username already exists",
			nil,
		)
	}

	hashedPassword, err := authentication.HashPassword(req.Password)
	if err != nil {
		l.Error("failed to hash password", zap.Error(err))
		return wrapper.ResponseFailed(
			http.StatusInternalServerError,
			contract.StatusCodeInternalServerError,
			"Failed to create user",
			nil,
		)
	}

	now := time.Now()
	user := &model.User{
		ID:        utils.GenerateUUID(),
		Username:  req.Username,
		Password:  hashedPassword,
		Role:      model.RoleUser,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.Repository.Create(ctx, user); err != nil {
		l.Error("failed to create user", zap.Error(err))
		return wrapper.ResponseFailed(
			http.StatusInternalServerError,
			contract.StatusCodeInternalServerError,
			"Failed to create user",
			nil,
		)
	}

	l.Info("user registered", zap.String("userId", user.ID), zap.String("username", user.Username))
	return wrapper.ResponseSuccess(http.StatusCreated, schema.ResponseRegister{
		UserID:   user.ID,
		Username: user.Username,
	})
}

func (u *UseCase) EnsureDefaultAdmin(ctx context.Context) error {
	l := u.Logger.With(zap.String("usecase", "EnsureDefaultAdmin"))

	existingAdmin := u.Repository.GetByUsername(ctx, schema.DefaultAdminUsername)
	if existingAdmin != nil {
		l.Debug("default admin already exists")
		return nil
	}

	hashedPassword, err := authentication.HashPassword(schema.DefaultAdminPassword)
	if err != nil {
		l.Error("failed to hash default admin password", zap.Error(err))
		return err
	}

	now := time.Now()
	admin := &model.User{
		ID:        utils.GenerateUUID(),
		Username:  schema.DefaultAdminUsername,
		Password:  hashedPassword,
		Role:      model.RoleAdmin,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.Repository.Create(ctx, admin); err != nil {
		l.Error("failed to create default admin", zap.Error(err))
		return err
	}

	l.Info("default admin created", zap.String("username", schema.DefaultAdminUsername))
	return nil
}
