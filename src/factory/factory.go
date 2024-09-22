package factory

import (
	"context"
	"database/sql"
	monitoring_use_cases "monitoring-system/src/internal/modules/monitoring/usecases"
	"monitoring-system/src/internal/modules/user-manager/domain/auth"
	auth_infra "monitoring-system/src/internal/modules/user-manager/infra/auth"
	user_manager_use_cases "monitoring-system/src/internal/modules/user-manager/usecases"
	"monitoring-system/src/pkg/logger"
)

type Factory struct {
	UserManager UserManager
	Monitoring  Monitoring
}

type UserManager struct {
	Infra    UserManagerInfra
	UseCases user_manager_use_cases.UseCases
}

type UserManagerInfra struct {
	AuthService auth.AuthService
	AuthRepo    auth.AuthRepository
}

type Monitoring struct {
	CameraManager monitoring_use_cases.CameraManager
}

func NewUserManager(ctx context.Context, logger logger.Logger, sqlDb *sql.DB) (*UserManager, error) {
	authRepo, err := auth_infra.NewAuthRepository(ctx, sqlDb, logger)
	if err != nil {
		logger.Error("Error creating auth repository %v", err)
		return nil, err
	}

	authService, err := auth_infra.NewAuth(authRepo, logger)
	if err != nil {
		logger.Error("Error creating auth service %v", err)
		return nil, err
	}

	return &UserManager{
		Infra: UserManagerInfra{
			AuthRepo:    authRepo,
			AuthService: authService,
		},
		UseCases: *user_manager_use_cases.NewUseCases(logger, authService),
	}, nil
}

func NewMonitoring(ctx context.Context, logger logger.Logger) (*Monitoring, error) {
	monitoring, err := monitoring_use_cases.NewCameraManager(ctx, logger)
	if err != nil {
		logger.Error("Error creating monitoring camera manager %v", err)
		return nil, err
	}
	return &Monitoring{
		CameraManager: monitoring,
	}, nil
}

func NewFactory(ctx context.Context, logger logger.Logger, sqlDb *sql.DB) (*Factory, error) {
	userManager, err := NewUserManager(ctx, logger, sqlDb)
	if err != nil {
		return nil, err
	}

	monitoring, err := NewMonitoring(ctx, logger)
	if err != nil {
		return nil, err
	}
	return &Factory{
		UserManager: *userManager,
		Monitoring:  *monitoring,
	}, nil
}