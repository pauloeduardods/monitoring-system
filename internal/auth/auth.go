package auth

import (
	"monitoring-system/pkg/logger"

	"golang.org/x/crypto/bcrypt"
)

type Auth interface {
	Register(username, password string) error
	Login(username, password string) (Token, error)
	ValidateToken(tokenString string) (*Claims, error)
}

type AuthService struct {
	authRepository AuthRepository
	logger         logger.Logger
}

type Token struct {
	Token string `json:"token"`
	// RefreshToken string `json:"refresh_token"` //TODO: Implement refresh token
}

func NewAuthService(authRepository AuthRepository, logger logger.Logger) (Auth, error) {
	return &AuthService{authRepository: authRepository, logger: logger}, nil
}

func (s *AuthService) Register(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.authRepository.Save(username, string(hashedPassword))
	return err
}

func (s *AuthService) Login(username, password string) (Token, error) {
	entity, err := s.authRepository.GetByUsername(username)

	if err != nil {
		return Token{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(entity.Password), []byte(password))
	if err != nil {
		return Token{}, err
	}

	token, err := s.generateToken(username)
	if err != nil {
		return Token{}, err
	}

	return Token{Token: token}, nil
}
