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
	sGroup.GET("", AuthHandler(fbauthclient, h.Service, h.GetUser))
	sGroup.PUT("", AuthHandler(fbauthclient, h.Service, h.SaveUser))
}

func (h *UserHandler) GetUser(c *gin.Context) {
	user, err := getUser(c, h.Service)
	slog.Info(fmt.Sprintf("GetUser: %v", user))
	if err != nil {
		slog.Debug("GetUser", "Error", err)
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) SaveUser(c *gin.Context) {
	value, _ := c.Get("id")
	uid := value.(string)
	var user *domain.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
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
