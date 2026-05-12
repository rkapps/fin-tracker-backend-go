package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/services"
)

type UserHandler struct {
	Service services.UserService
}

func NewUserHandler(router *gin.Engine, service services.UserService) *UserHandler {
	return &UserHandler{service}
}

func (h *UserHandler) RegisterRoutes(router *gin.Engine, fbauthclient *auth.Client) {

	sGroup := router.Group("/user")
	sGroup.GET("", AuthHandler(fbauthclient, h.GetUser))
	sGroup.PUT("", AuthHandler(fbauthclient, h.SaveUser))
}

func (h *UserHandler) GetUser(c *gin.Context) {
	uid, err := getUID(c)
	slog.Info(fmt.Sprintf("GetUID: %v", uid))
	if err != nil {
		slog.Debug("GetUser", "Error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}
	user, err := h.Service.GetUser(uid)
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User not found",
		})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) SaveUser(c *gin.Context) {
	uid, err := getUID(c)
	var user *domain.User
	err = json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	user.ID = uid
	err = h.Service.SaveUser(user)
	if err != nil {
		slog.Debug("SaveUser", "Error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

}
