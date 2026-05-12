package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/services"
	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
)

type AccountsHandler struct {
	Service services.AccountsService
}

func NewAccountsHandler(router *gin.Engine, service services.AccountsService) *AccountsHandler {
	return &AccountsHandler{service}
}

func (a *AccountsHandler) RegisterRoutes(router *gin.Engine, fbAuthClient *auth.Client) {

	sGroup := router.Group("/accounts")
	sGroup.GET("", AuthHandler(fbAuthClient, a.GetAccounts))
	sGroup.POST("", AuthHandler(fbAuthClient, a.CreateAccount))
	sGroup.GET(":id", AuthHandler(fbAuthClient, a.GetAccount))
	sGroup.PUT(":id", AuthHandler(fbAuthClient, a.UpdateAccount))
	sGroup.DELETE(":id", AuthHandler(fbAuthClient, a.DeleteAccount))
	sGroup.POST(":id/activities", AuthHandler(fbAuthClient, a.ImportActivities))

	// sGroup.POST("/load", AuthHandler(fbAuthClient, h.UserService, h.LoadAccounts))
	// sGroup.GET("/:id/delete", AuthHandler(fbAuthClient, h.UserService, h.DeleteAccount))
	// sGroup.GET("/names", AuthHandler(fbAuthClient, h.UserService, h.GetAccountNames))

	// sGroup.GET("/activities", AuthHandler(fbAuthClient, h.UserService, h.GetActivities))
	// sGroup.GET("/holdings", AuthHandler(fbAuthClient, h.UserService, h.GetHoldings))
	// sGroup.GET("/lots", AuthHandler(fbAuthClient, h.UserService, h.GetActivityLots))
	// sGroup.GET("/income", AuthHandler(fbAuthClient, h.UserService, h.GetIncome))
	// sGroup.GET("/gainloss", AuthHandler(fbAuthClient, h.UserService, h.GetGainLoss))
}

// DeleteAccount
func (a *AccountsHandler) DeleteAccount(c *gin.Context) {

	uid, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	id := c.Param("id")

	err = a.Service.DeleteAccount(c, uid, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
}

// GetAccounts gets the accounts in the portfolio
func (a *AccountsHandler) GetAccount(c *gin.Context) {
	// h.getUser(c)
	uid, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	id := c.Param("id")
	accts, err := a.Service.GetAccount(uid, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, accts)
	c.JSON(http.StatusOK, id)
}

// GetAccounts gets the accounts in the portfolio
func (a *AccountsHandler) GetAccounts(c *gin.Context) {
	// h.getUser(c)
	uid, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	accts, err := a.Service.GetAccounts(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, accts)

}

func (a *AccountsHandler) CreateAccount(c *gin.Context) {

	uid, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	slog.Info("CreateAccount", "UId", uid)

	var data *domain.Account
	err = json.NewDecoder(c.Request.Body).Decode(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	slog.Info("CreateAccount", "UId", uid, "Data", data)

	acct, err := a.Service.CreateAccount(c, uid, data)
	if err != nil {
		slog.Debug("CreateAccount", "Error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, acct)

}

func (a *AccountsHandler) UpdateAccount(c *gin.Context) {

	uid, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	id := c.Param("id")

	var data *domain.Account
	err = json.NewDecoder(c.Request.Body).Decode(&data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	err = a.Service.UpdateAccount(c, uid, id, data)
	if err != nil {
		slog.Debug("UpdateAccount", "Error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

}

// // GetAccountNames gets the account names
// func (a *AccountsHandler) GetAccountNames(c *gin.Context) {
// 	// h.getUser(c)
// }

// // GetActivities returns a list of activities
// func (a *AccountsHandler) GetActivities(c *gin.Context) {
// }

// // GetActivityLots returns a list of activity lots
// func (a *AccountsHandler) GetActivityLots(c *gin.Context) {
// }

// // GetHoldings returns a list of activity lots
// func (a *AccountsHandler) GetHoldings(c *gin.Context) {
// }

// // GetIncome returns income
// func (a *AccountsHandler) GetIncome(c *gin.Context) {
// }

// // GetGainloss returns gain loss
// func (a *AccountsHandler) GetGainLoss(c *gin.Context) {
// }

// // LoadAccounts loads the accounts in the portfolio
// func (a *AccountsHandler) LoadAccounts(c *gin.Context) {

// 	user, err := getUser(c, h.UserService)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, err)
// 		return
// 	}
// 	var accts accounts.Accounts
// 	err = json.NewDecoder(c.Request.Body).Decode(&accts)
// 	if err != nil {
// 		if err == io.EOF {
// 			slog.Debug("LoadAccounts", "Request Body is empty", err)
// 			c.JSON(http.StatusBadRequest, err)
// 		} else {
// 			slog.Debug("LoadAccounts", "Decode error", err)
// 			c.JSON(http.StatusBadRequest, err)
// 		}
// 		return
// 	}
// 	slog.Debug(fmt.Sprintf("LoadAccounts count: %d", len(accts)))

// 	err = h.Service.LoadAccounts(c, *user, accts)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, err)
// 		return
// 	}

// }

// LoadActivities loads the activities in the portfolio
func (a *AccountsHandler) ImportActivities(c *gin.Context) {

	uid, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	acctId := c.Param("id")
	sstartDate := c.Query("startDate")
	startDate := utils.DateFromString(sstartDate)
	slog.Info("ImportActivities", "UId", uid, "acctId", acctId, "StartDate", startDate)

	var data []*domain.ActivityImport
	err = json.NewDecoder(c.Request.Body).Decode(&data)
	if err != nil {
		slog.Debug("ImportActivities", "Error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	a.Service.ImportActivities(c, uid, acctId, startDate, data)
	slog.Info("ImportActivities", "Count", len(data))
}
