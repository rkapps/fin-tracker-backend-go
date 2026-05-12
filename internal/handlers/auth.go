package handlers

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
		if len(header) == 0 {
			slog.Info("AuthHandler", "error", "Authorization header not available")
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Authorization header not available.")
			return
		}
		if len(header) > 7 && strings.ToLower(header[0:6]) == "bearer" {
			token = header[7:]
		} else {
			slog.Info("AuthHandler", "error", "Authorization token not available")
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Authorization token not available.")
			return
		}

		authToken, err := fbAuthClient.VerifyIDToken(c, token)
		if err != nil {
			slog.Info("AuthHandler", "error", "Authorization header not verified")
			c.AbortWithStatusJSON(http.StatusUnauthorized, "Authorization token not verified.")
			return
		}
		// slog.Debug("AuthHandler", "UID", authToken.UID)
		c.Set("uid", authToken.UID)
		f(c)
	}
}

func getUID(c *gin.Context) (string, error) {
	value, _ := c.Get("uid")
	id := value.(string)
	return id, nil
}
