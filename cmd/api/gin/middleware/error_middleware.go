package middleware

import (
	"monitoring-system/pkg/app_error"
	"monitoring-system/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		for _, err := range c.Errors {
			switch e := err.Err.(type) {
			case *app_error.ApiError:
				c.AbortWithStatusJSON(e.StatusCode, e)
				c.Abort()
			case validator.ValidationErrors:
				errMsg := make(map[string]string)
				for _, fieldErr := range e {
					errMsg[fieldErr.Field()] = fieldErr.Tag()
				}
				c.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
					"message": "Validation Error",
					"errors":  errMsg,
				})
				c.Abort()
				return
			default:
				log.Error("Error occurred %v", e)
				c.AbortWithStatusJSON(http.StatusInternalServerError, map[string]string{"message": e.Error()})
				c.Abort()
			}
		}
	}
}
