package middleware

import (
	"monitoring-system/src/internal/modules/user-manager/domain/auth"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware interface {
	AuthMiddleware() gin.HandlerFunc
	AuthMiddlewareWS() gin.HandlerFunc
	AuthMiddlewareRegister() gin.HandlerFunc
}

type AuthMiddlewareImpl struct {
	auth     auth.AuthService
	authRepo auth.AuthRepository
}

func NewAuthMiddleware(a auth.AuthService, repo auth.AuthRepository) AuthMiddleware {
	return &AuthMiddlewareImpl{
		auth:     a,
		authRepo: repo,
	}
}

func (a *AuthMiddlewareImpl) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		token := authHeader[7:]

		claims, err := a.auth.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// user, err := a.authRepo.GetByUsername(c.Request.Context(), claims.Username)
		// if err != nil {
		// 	c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		// }

		c.Set("jwtToken", token)
		c.Set("claims", claims)
		// c.Set("user", user)

		c.Next()
	}
}

func (a *AuthMiddlewareImpl) AuthMiddlewareRegister() gin.HandlerFunc {
	return func(c *gin.Context) {
		usersCount, err := a.authRepo.CountUsers(c.Request.Context())
		if err != nil || usersCount == 0 {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		token := authHeader[7:]

		claims, err := a.auth.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// user, err := a.authRepo.GetByUsername(c.Request.Context(), claims.Username)
		// if err != nil {
		// 	c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		// 	return
		// }

		c.Set("jwtToken", token)
		c.Set("claims", claims)
		// c.Set("user", user)

		c.Next()
	}
}

func (a *AuthMiddlewareImpl) AuthMiddlewareWS() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Query("token")

		claims, err := a.auth.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		// user, err := a.authRepo.GetByUsername(c.Request.Context(), claims.Username)
		// if err != nil {
		// 	c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
		// }

		c.Set("jwtToken", token)
		c.Set("claims", claims)
		// c.Set("user", user)

		c.Next()
	}
}
