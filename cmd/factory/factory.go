package factory

import (
	"context"
	"database/sql"
	"monitoring-system/domain/auth"
	"monitoring-system/domain/camera_manager"
	"monitoring-system/internal/repository"
	"monitoring-system/pkg/logger"
)

type Factory struct {
	Repositories *Repositories
	Internal     *Internal
}

type Repositories struct {
	Auth auth.AuthRepository
}

type Internal struct {
	CameraManager camera_manager.CameraManager
	Auth          auth.Auth
}

func NewRepositories(logger logger.Logger, sqlDb *sql.DB) (*Repositories, error) {
	authRepo, err := repository.NewAuthRepository(sqlDb, logger)
	if err != nil {
		logger.Error("Error creating auth repository %v", err)
		return nil, err
	}

	return &Repositories{
		Auth: authRepo,
	}, nil
}

func NewInternal(cm camera_manager.CameraManager, auth auth.Auth) *Internal {
	return &Internal{
		CameraManager: cm,
		Auth:          auth,
	}
}

func New(ctx context.Context, logger logger.Logger, sqlDb *sql.DB) (*Factory, error) {
	repos, err := NewRepositories(logger, sqlDb)
	if err != nil {
		return nil, err
	}
	auth, err := auth.NewAuth(repos.Auth, logger)
	if err != nil {
		return nil, err
	}

	cm, err := camera_manager.NewCameraManager(ctx, logger)
	if err != nil {
		return nil, err

	}

	err = cm.CheckSystemCameras()
	if err != nil {
		return nil, err
	}

	internal := NewInternal(cm, auth)

	return &Factory{
		Repositories: repos,

		Internal: internal,
	}, nil
}
