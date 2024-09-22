package auth

import (
	"context"
	"monitoring-system/src/pkg/app_error"
	"regexp"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var (
	UsernameLength   = app_error.NewApiError(400, "Username length")
	PasswordValidate = app_error.NewApiError(400, "Password requirements not satisfied")
)

type AuthService interface {
	Register(ctx context.Context, input RegisterInput) error
	Login(ctx context.Context, input LoginInput) (Token, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type RegisterInput struct {
	Username string
	Password string
}

func (i RegisterInput) Validate() error {
	if len(i.Username) <= 5 {
		return UsernameLength
	}
	if len(i.Password) < 8 {
		return app_error.NewApiError(400, "password must be at least 8 characters long")
	}

	if matched, _ := regexp.MatchString(`[A-Z]`, i.Password); !matched {
		return app_error.NewApiError(400, "password must contain at least one uppercase letter")
	}

	if matched, _ := regexp.MatchString(`[a-z]`, i.Password); !matched {
		return app_error.NewApiError(400, "password must contain at least one lowercase letter")
	}

	if matched, _ := regexp.MatchString(`[0-9]`, i.Password); !matched {
		return app_error.NewApiError(400, "password must contain at least one number")
	}

	if matched, _ := regexp.MatchString(`[!@#\$%\^&\*]`, i.Password); !matched {
		return app_error.NewApiError(400, "password must contain at least one special character (!@#$%^&*)")
	}

	return nil
}

type LoginInput struct {
	Username string
	Password string
}

func (i LoginInput) Validate() error {
	return nil
}

type AuthRepository interface {
	GetByUsername(ctx context.Context, username string) (*AuthEntity, error)
	Save(ctx context.Context, username, password string) error
}

type Token struct {
	Token string `json:"token"`
	// RefreshToken string `json:"refresh_token"` //TODO: Implement refresh token
}

type AuthEntity struct {
	ID       uuid.UUID
	Username string
	Password string `json:"-"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}
