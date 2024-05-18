package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type Cors struct {
	Origin      string
	Methods     string
	Headers     string
	Credentials bool
}

func (co *Cors) CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", co.Origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", co.Methods)
		c.Writer.Header().Set("Access-Control-Allow-Headers", co.Headers)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", strconv.FormatBool(co.Credentials))

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
