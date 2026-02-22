package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/internal/team/repository"
	"github.com/Alwanly/management-sport/internal/team/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/utils"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"go.uber.org/zap"
)

const ContextName = "Internal.Team.Usecase"

type UseCase struct {
	Config     *config.GlobalConfig
	Logger     *zap.Logger
	Repository repository.IRepository
}

type IUseCase interface {
	Create(context.Context, *schema.RequestTeamCreate) wrapper.JSONResult
	Get(context.Context, *schema.RequestTeamGet) wrapper.JSONResult
	List(context.Context, *schema.RequestTeamList) wrapper.JSONResult
	Update(context.Context, *schema.RequestTeamUpdate) wrapper.JSONResult
	Delete(context.Context, *schema.RequestTeamDelete) wrapper.JSONResult
	ProcessLogoUpload(teamName string, file *multipart.FileHeader) (string, error)
	DeleteOldLogo(logoURL string) error
}

func NewUseCase(uc UseCase) IUseCase {
	return &UseCase{
		Config:     uc.Config,
		Logger:     uc.Logger,
		Repository: uc.Repository,
	}
}

func (u *UseCase) Create(ctx context.Context, req *schema.RequestTeamCreate) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Create"))

	now := time.Now()
	team := &model.Team{
		ID:          utils.GenerateUUID(),
		Name:        req.Name,
		LogoURL:     "",
		FoundedYear: req.FoundedYear,
		Address:     req.Address,
		City:        req.City,
		CreatedBy:   req.AuthUserData.UserID,
		CreatedAt:   now,
		UpdatedAt:   now,
		UpdatedBy:   req.AuthUserData.UserID,
	}

	if err := u.Repository.Create(ctx, team); err != nil {
		l.Error("failed to create a team", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to create a team", nil)
	}

	l.Debug("team created", zap.String("id", team.ID))
	return wrapper.ResponseSuccess(http.StatusCreated, schema.ResponseTeamCreate{ID: team.ID})
}

func (u *UseCase) ProcessLogoUpload(teamName string, file *multipart.FileHeader) (string, error) {
	l := u.Logger.With(zap.String("usecase", "ProcessLogoUpload"))

	// Parse allowed image types from config
	allowedTypesStr := strings.Split(u.Config.AllowedImageTypes, ",")

	// Validate file
	if err := utils.ValidateImageFile(file, u.Config.MaxUploadSize, allowedTypesStr); err != nil {
		l.Error("file validation failed", zap.Error(err))
		return "", err
	}

	// Get file extension
	ext := filepath.Ext(file.Filename)

	// Generate unique filename
	filename := utils.GenerateUniqueFilename(teamName, ext)

	// Build directory path
	directory := filepath.Join(u.Config.UploadDirectory, "logo_teams")

	// Save file
	fullPath, err := utils.SaveUploadedFile(file, directory, filename)
	if err != nil {
		l.Error("failed to save uploaded file", zap.Error(err))
		return "", fmt.Errorf("failed to save uploaded file")
	}

	// Return URL path (not filesystem path)
	urlPath := filepath.ToSlash(filepath.Join("/images", "logo_teams", filename))
	l.Info("logo uploaded successfully",
		zap.String("path", urlPath),
		zap.String("filesystem_path", fullPath))

	return urlPath, nil
}

func (u *UseCase) DeleteOldLogo(logoURL string) error {
	if logoURL == "" {
		return nil
	}

	// Convert URL path to filesystem path
	// Example: /images/logo_teams/logo-team-name.jpg -> ./images/logo_teams/logo-team-name.jpg
	relativePath := strings.TrimPrefix(logoURL, "/images/")
	fullPath := filepath.Join(u.Config.UploadDirectory, relativePath)

	// Delete the file
	if err := utils.DeleteFile(fullPath); err != nil {
		u.Logger.Error("failed to delete old logo",
			zap.String("path", fullPath),
			zap.Error(err))
		return err
	}

	u.Logger.Info("old logo deleted successfully", zap.String("path", fullPath))
	return nil
}

func (u *UseCase) Get(ctx context.Context, req *schema.RequestTeamGet) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Get"))

	team := u.Repository.Get(ctx, req.ID)
	if team == nil {
		l.Error("team not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("TEAM_NOT_FOUND"), "Team not found", nil)
	}

	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseTeamGet{
		ID:          team.ID,
		Name:        team.Name,
		LogoURL:     team.LogoURL,
		FoundedYear: team.FoundedYear,
		Address:     team.Address,
		City:        team.City,
	})
}

func (u *UseCase) List(ctx context.Context, req *schema.RequestTeamList) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "List"))

	teams, total := u.Repository.List(ctx, *req)

	response := req.ToResponse(teams)
	l.Debug("teams listed", zap.Int64("total", total))
	return wrapper.ResponsePagination(req.Page, req.PageSize, len(teams), int(total), response, nil)
}

func (u *UseCase) Update(ctx context.Context, req *schema.RequestTeamUpdate) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Update"))

	team := u.Repository.Get(ctx, req.ID)
	if team == nil {
		l.Error("team not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("TEAM_NOT_FOUND"), "Team not found", nil)
	}

	team.Name = req.Name
	team.FoundedYear = req.FoundedYear
	team.Address = req.Address
	team.City = req.City
	team.UpdatedAt = time.Now()
	team.UpdatedBy = req.AuthUserData.UserID

	if err := u.Repository.Update(ctx, team); err != nil {
		l.Error("failed to update a team", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to update a team", nil)
	}

	l.Debug("team updated", zap.String("id", team.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseTeamUpdate{ID: team.ID})
}

func (u *UseCase) Delete(ctx context.Context, req *schema.RequestTeamDelete) wrapper.JSONResult {
	l := u.Logger.With(zap.String("usecase", "Delete"))

	team := u.Repository.Get(ctx, req.ID)
	if team == nil {
		l.Error("team not found", zap.String("id", req.ID))
		return wrapper.ResponseFailed(http.StatusNotFound, contract.CreateStatusCode("TEAM_NOT_FOUND"), "Team not found", nil)
	}

	if err := u.Repository.Delete(ctx, req.ID); err != nil {
		l.Error("failed to delete a team", zap.Error(err))
		return wrapper.ResponseFailed(500, contract.StatusCodeInternalServerError, "Failed to delete a team", nil)
	}

	l.Debug("team deleted", zap.String("id", req.ID))
	return wrapper.ResponseSuccess(http.StatusOK, schema.ResponseTeamDelete{})
}
