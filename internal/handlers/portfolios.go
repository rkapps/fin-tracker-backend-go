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

func (h *PortfoliosHandler) RegisterRoutes(router *gin.Engine, fbAuthClient *auth.Client) {

	sGroup := router.Group("/portfolios")
	sGroup.GET("/accounts", AuthHandler(fbAuthClient, h.UserService, h.GetAccounts))
	sGroup.POST("/accounts/load", AuthHandler(fbAuthClient, h.UserService, h.LoadAccounts))
	sGroup.GET("/accounts/:id/delete", AuthHandler(fbAuthClient, h.UserService, h.DeleteAccount))
	sGroup.GET("/accountNames", AuthHandler(fbAuthClient, h.UserService, h.GetAccountNames))
	sGroup.POST("/loadActivities", AuthHandler(fbAuthClient, h.UserService, h.LoadActivities))
	// sGroup.GET("/portfolios/syncWallets", AuthHandler(fbAuthClient, h.UserService, h.syncWallets))
	// sGroup.GET("/portfolios/syncExchanges", AuthHandler(fbAuthClient, h.UserService, h.syncExchanges))
	// sGroup.GET("/portfolios/refreshAll", AuthHandler(fbAuthClient, h.UserService, h.refreshAll))

	sGroup.GET("/activities", AuthHandler(fbAuthClient, h.UserService, h.GetInvestmentsActivities))
	sGroup.GET("/holdings", AuthHandler(fbAuthClient, h.UserService, h.GetInvestmentsHoldings))
	sGroup.GET("/lots", AuthHandler(fbAuthClient, h.UserService, h.GetInvestmentsActivityLots))
	sGroup.GET("/income", AuthHandler(fbAuthClient, h.UserService, h.GetInvestmentsIncome))
	sGroup.GET("/gainloss", AuthHandler(fbAuthClient, h.UserService, h.GetInvestmentsGainLoss))
	// sGroup.GET("/portfolios/symbols", AuthHandler(fbAuthClient, h.UserService, h.GetInvestmentsSymbols))

}

// DeleteAccount
func (h *PortfoliosHandler) DeleteAccount(c *gin.Context) {
}

// GetAccounts gets the accounts in the portfolio
func (h *PortfoliosHandler) GetAccounts(c *gin.Context) {
	// h.getUser(c)
}

// GetAccountNames gets the account names
func (h *PortfoliosHandler) GetAccountNames(c *gin.Context) {
	// h.getUser(c)
}

// GetInvestmentActivities returns a list of activities
func (h *PortfoliosHandler) GetInvestmentsActivities(c *gin.Context) {
}

// GetInvestmentActivityLots returns a list of activity lots
func (h *PortfoliosHandler) GetInvestmentsActivityLots(c *gin.Context) {
}

// GetInvestmentHoldings returns a list of activity lots
func (h *PortfoliosHandler) GetInvestmentsHoldings(c *gin.Context) {
}

// GetInvestmentsIncome returns income
func (h *PortfoliosHandler) GetInvestmentsIncome(c *gin.Context) {
}

// GetInvestmentGainloss returns gain loss
func (h *PortfoliosHandler) GetInvestmentsGainLoss(c *gin.Context) {
}

// LoadAccounts loads the accounts in the portfolio
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

// LoadActivities loads the activities in the portfolio
func (h *PortfoliosHandler) LoadActivities(c *gin.Context) {
}
