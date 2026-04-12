package handlers

import (
	"errors"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/services"
)

func AuthHandler(fbAuthClient *auth.Client, userService services.UserService, f func(c *gin.Context)) func(c *gin.Context) {

	return func(c *gin.Context) {

		var header = c.GetHeader("Authorization")
		var token string
		// slog.Debug("AuthHandler", "Token", token)
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
		// slog.Debug("AuthHandler", "UID", authToken.UID)
		c.Set("id", authToken.UID)
		f(c)
	}
}

func getUser(c *gin.Context, userService services.UserService) (*domain.User, error) {
	value, _ := c.Get("id")
	id := value.(string)

	// slog.Debug("AuthHandler", "getUser", id)
	user, _ := userService.GetUser(id)
	if user == nil {
		return nil, errors.New("User not setup")
	}
	return user, nil
}
