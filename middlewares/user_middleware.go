package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthUser(c *gin.Context) {
	if c.Request.URL.Path == "/user/refresh-token" {
		refreshToken, err := c.Cookie("refreshToken")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		c.Set("refreshToken", refreshToken)
	}
	clientToken := c.Request.Header.Get("token")

	if clientToken == "" {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "No Authoriztion header provided!",
		})
		c.Abort()
		return
	}
	c.Set("token", clientToken)

	c.Next()
}
