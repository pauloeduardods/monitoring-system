package handlers

import (
	"monitoring-system/src/internal/modules/user-manager/domain/auth"
	user_manager_use_cases "monitoring-system/src/internal/modules/user-manager/usecases"
	"monitoring-system/src/pkg/validator"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	auth      user_manager_use_cases.UseCases
	validator validator.Validator
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required" validate:"min=3,max=50"`
	Password string `json:"password" binding:"required" validate:"min=8"`
	// Name     string `json:"name" binding:"required" validate:"min=3,max=50"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" validate:"min=3,max=50"`
	Password string `json:"password" binding:"required" validate:"min=8"`
}

func NewAuthHandler(a user_manager_use_cases.UseCases, validator validator.Validator) *AuthHandler {
	return &AuthHandler{
		auth:      a,
		validator: validator,
	}
}

func (a *AuthHandler) Login() gin.HandlerFunc {
	return func(g *gin.Context) {
		var login LoginRequest
		if err := g.ShouldBindJSON(&login); err != nil {
			g.Error(err)
			return
		}

		err := a.validator.Validate(&login)
		if err != nil {
			g.Error(err)
			return
		}

		res, err := a.auth.Login.Execute(g.Request.Context(), auth.LoginInput{Username: login.Username, Password: login.Password})
		if err != nil {
			g.Error(err)
			return
		} else {
			g.JSON(http.StatusOK, res)
		}
	}
}

func (a *AuthHandler) Register() gin.HandlerFunc {
	return func(g *gin.Context) {
		var signUp RegisterRequest
		if err := g.ShouldBindJSON(&signUp); err != nil {
			g.Error(err)
			return
		}

		err := a.validator.Validate(&signUp)
		if err != nil {
			g.Error(err)
			return
		}

		err = a.auth.Register.Execute(g.Request.Context(), auth.RegisterInput{Username: signUp.Username, Password: signUp.Password})
		if err != nil {
			g.Error(err)
			return
		} else {
			g.JSON(http.StatusOK, gin.H{"message": "User registered successfully"})
		}
	}
}
