package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/fin-tracker-backend-go/internal/portfolios/accounts"
	"github.com/rkapps/fin-tracker-backend-go/internal/services"
)

type PortfoliosHandler struct {
	Service     services.PortfoliosService
	UserService services.UserService
}

func NewPortfoliosHandler(router *gin.Engine, service services.PortfoliosService, userService services.UserService) *PortfoliosHandler {
	return &PortfoliosHandler{service, userService}
}

func (h *PortfoliosHandler) RegisterRoutes(router *gin.Engine, fbauthclient *auth.Client) {

	sGroup := router.Group("/portfolios")
	sGroup.GET("/accounts", AuthHandler(fbauthclient, h.UserService, h.GetAccounts))
	sGroup.POST("/accounts/load", AuthHandler(fbauthclient, h.UserService, h.LoadAccounts))
}

// GetAccounts gets the accounts in the portfolio
func (h *PortfoliosHandler) GetAccounts(c *gin.Context) {
	// h.getUser(c)
}

// LoadAccounts gets the accounts in the portfolio
func (h *PortfoliosHandler) LoadAccounts(c *gin.Context) {

	user, err := getUser(c, h.UserService)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var accts accounts.Accounts
	err = json.NewDecoder(c.Request.Body).Decode(&accts)
	if err != nil {
		if err == io.EOF {
			slog.Debug("LoadAccounts", "Request Body is empty", err)
			c.JSON(http.StatusBadRequest, err)
		} else {
			slog.Debug("LoadAccounts", "Decode error", err)
			c.JSON(http.StatusBadRequest, err)
		}
		return
	}
	slog.Debug(fmt.Sprintf("LoadAccounts count: %d", len(accts)))

	err = h.Service.LoadAccounts(c, *user, accts)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

}
