package gateway

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

func tlsHandler(port int, host string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(
			secure.Options{
				SSLRedirect: true,
				SSLHost:     host + ":" + strconv.Itoa(port),
			},
		)
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadGateway, gin.H{
					"error": err,
				},
			)
			return
		}

		c.Next()
	}
}
