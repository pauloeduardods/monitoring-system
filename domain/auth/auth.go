package auth

import (
	"monitoring-system/pkg/app_error"
	"monitoring-system/pkg/logger"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Auth interface {
	Register(username, password string) error
	Login(username, password string) (Token, error)
	ValidateToken(tokenString string) (*Claims, error)
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

type AuthService struct {
	authRepository AuthRepository
	logger         logger.Logger
}

type AuthRepository interface {
	GetByUsername(username string) (*AuthEntity, error)
	Save(username, password string) error
}

func NewAuth(authRepository AuthRepository, logger logger.Logger) (Auth, error) {
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

	errInvalidUsernameOrPassword := app_error.NewApiError(401, "Invalid username or password")

	if err != nil {
		return Token{}, errInvalidUsernameOrPassword
	}

	err = bcrypt.CompareHashAndPassword([]byte(entity.Password), []byte(password))
	if err != nil {
		e := err.Error()
		if strings.Contains(e, "hashedPassword is not the hash of the given password") {
			return Token{}, errInvalidUsernameOrPassword
		}
		return Token{}, err
	}

	token, err := s.generateToken(username)
	if err != nil {
		return Token{}, err
	}

	return Token{Token: token}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	return s.validateToken(tokenString)
}
