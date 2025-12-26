package api

import (
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func AuthHandler(fbAuthClient *auth.Client, f func(c *gin.Context)) func(c *gin.Context) {

	return func(c *gin.Context) {

		var header = c.GetHeader("Authorization")
		var token string

		if len(header) > 7 && strings.ToLower(header[0:6]) == "bearer" {
			token = header[7:]
		} else {
			c.IndentedJSON(http.StatusOK, "Authorization token not available.")
			return
		}

		authToken, err := fbAuthClient.VerifyIDToken(c, token)
		if err != nil {
			c.IndentedJSON(http.StatusOK, "Authorization token not verified.")
			return
		}

		c.Set("uid", authToken.UID)
		f(c)
	}
}
