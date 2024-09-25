package middleware

import (
	"monitoring-system/src/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RecoveryHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("Recovered from panic!!!", r)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
				c.Abort()
			}
		}()
		c.Next()
	}
}
