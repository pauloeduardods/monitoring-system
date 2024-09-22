package auth

import (
	"context"
	"monitoring-system/internal/modules/user-manager/domain/auth"
	"monitoring-system/pkg/app_error"
	"monitoring-system/pkg/logger"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("my_secret_key") //TODO: move to env

type AuthService struct {
	authRepository auth.AuthRepository
	logger         logger.Logger
}

func NewAuth(authRepository auth.AuthRepository, logger logger.Logger) (auth.AuthService, error) {
	return &AuthService{authRepository: authRepository, logger: logger}, nil
}

func (s *AuthService) Register(ctx context.Context, input auth.RegisterInput) error {
	if err := input.Validate(); err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.authRepository.Save(ctx, input.Username, string(hashedPassword))
	return err
}

func (s *AuthService) Login(ctx context.Context, input auth.LoginInput) (auth.Token, error) {
	if err := input.Validate(); err != nil {
		return auth.Token{}, err
	}
	entity, err := s.authRepository.GetByUsername(ctx, input.Username)

	errInvalidUsernameOrPassword := app_error.NewApiError(401, "Invalid username or password")

	if err != nil {
		return auth.Token{}, errInvalidUsernameOrPassword
	}

	err = bcrypt.CompareHashAndPassword([]byte(entity.Password), []byte(input.Password))
	if err != nil {
		e := err.Error()
		if strings.Contains(e, "hashedPassword is not the hash of the given password") {
			return auth.Token{}, errInvalidUsernameOrPassword
		}
		return auth.Token{}, err
	}

	token, err := s.generateToken(input.Username)
	if err != nil {
		return auth.Token{}, err
	}

	return auth.Token{Token: token}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*auth.Claims, error) {
	return s.validateToken(tokenString)
}

func (s *AuthService) generateToken(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &auth.Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "monitoring-system",
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Subject:   "auth",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        "123",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func (s *AuthService) validateToken(tokenString string) (*auth.Claims, error) {
	claims := &auth.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}
