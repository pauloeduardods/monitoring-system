package modules

import (
	"database/sql"
	"monitoring-system/internal/auth"
	"monitoring-system/internal/domain/camera"
	"monitoring-system/pkg/logger"
)

type Modules struct {
	Repositories *Repositories
	Services     *Services
	Internal     *Internal
}

type Repositories struct {
	Auth auth.AuthRepository
}

type Services struct {
	Auth auth.Auth
}

type Internal struct {
	Camera camera.Camera
}

func NewRepositories(logger logger.Logger, sqlDb *sql.DB) (*Repositories, error) {
	auth, err := auth.NewAuthRepository(sqlDb, logger)
	if err != nil {
		logger.Error("Error creating auth repository %v", err)
		return nil, err
	}

	return &Repositories{
		Auth: auth,
	}, nil
}

func NewServices(repos *Repositories, logger logger.Logger) (*Services, error) {
	auth, err := auth.NewAuthService(repos.Auth, logger)
	if err != nil {
		logger.Error("Error creating auth service %v", err)
		return nil, err
	}

	return &Services{
		Auth: auth,
	}, nil
}

func NewInternal(cam camera.Camera) *Internal {
	return &Internal{
		Camera: cam,
	}
}

func New(logger logger.Logger, sqlDb *sql.DB, cam camera.Camera) (*Modules, error) {
	repos, err := NewRepositories(logger, sqlDb)
	if err != nil {
		return nil, err
	}
	services, err := NewServices(repos, logger)
	if err != nil {
		return nil, err
	}
	internal := NewInternal(cam)

	return &Modules{
		Repositories: repos,
		Services:     services,
		Internal:     internal,
	}, nil
}
