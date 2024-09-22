package user_manager_use_cases

import (
	"monitoring-system/src/internal/modules/user-manager/domain/auth"
	"monitoring-system/src/pkg/logger"
)

type UseCases struct {
	Register *RegisterUserUseCase
	Login    *LoginUserUseCase
}

func NewUseCases(logger logger.Logger, authService auth.AuthService) *UseCases {
	return &UseCases{
		Register: NewRegisterUserUseCase(logger, authService),
		Login:    NewLoginUserUseCase(logger, authService),
	}
}
