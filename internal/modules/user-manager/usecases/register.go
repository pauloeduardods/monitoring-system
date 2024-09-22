package user_manager_use_cases

import (
	"context"
	"monitoring-system/internal/modules/user-manager/domain/auth"
	"monitoring-system/pkg/logger"
)

type RegisterUserUseCase struct {
	logger      logger.Logger
	authService auth.AuthService
}

func NewRegisterUserUseCase(logger logger.Logger, authService auth.AuthService) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		logger:      logger,
		authService: authService,
	}
}

func (uc RegisterUserUseCase) Execute(ctx context.Context, input auth.RegisterInput) (err error) {
	if err := input.Validate(); err != nil {
		return err
	}

	if err := uc.authService.Register(ctx, input); err != nil {
		return err
	}
	return nil
}
