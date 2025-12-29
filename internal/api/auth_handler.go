package api

import (
	"log/slog"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
)

func AuthHandler(fbAuthClient *auth.Client, f func(c *gin.Context)) func(c *gin.Context) {

	return func(c *gin.Context) {

		var header = c.GetHeader("Authorization")
		var token string
		slog.Debug("AuthHandler", "Token", token)
		if len(header) == 0 {
			c.JSON(http.StatusUnauthorized, "Authorization header not available.")
			return
		}
		if len(header) > 7 && strings.ToLower(header[0:6]) == "bearer" {
			token = header[7:]
		} else {
			c.JSON(http.StatusUnauthorized, "Authorization token not available.")
			return
		}

		authToken, err := fbAuthClient.VerifyIDToken(c, token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, "Authorization token not verified.")
			return
		}
		slog.Debug("AuthHandler", "UID", authToken.UID)
		c.Set("uid", authToken.UID)
		f(c)
	}
}
