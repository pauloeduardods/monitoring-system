package user_manager_use_cases

import (
	"context"
	"monitoring-system/internal/modules/user-manager/domain/auth"
	"monitoring-system/pkg/logger"
)

type LoginUserUseCase struct {
	logger      logger.Logger
	authService auth.AuthService
}

func NewLoginUserUseCase(logger logger.Logger, authService auth.AuthService) *LoginUserUseCase {
	return &LoginUserUseCase{
		logger:      logger,
		authService: authService,
	}
}

func (uc LoginUserUseCase) Execute(ctx context.Context, input auth.LoginInput) (auth.Token, error) {
	if err := input.Validate(); err != nil {
		return auth.Token{}, err
	}

	token, err := uc.authService.Login(ctx, input)
	if err != nil {
		return auth.Token{}, err
	}
	return token, nil
}
