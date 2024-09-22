package middleware

import (
	"monitoring-system/src/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RecoveryHandler(log logger.Logger) gin.RecoveryFunc {
	return func(c *gin.Context, err any) {
		c.Next()
		log.Error("Error occurred %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{"message": "Service Unavailable"})
		c.Abort()
	}
}
