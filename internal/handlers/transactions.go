package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/services"
	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
)

type TransactionsHandler struct {
	Service     services.TransactionsService
	UserService services.UserService
}

func NewTransactionsHandler(router *gin.Engine, service services.TransactionsService, userService services.UserService) *TransactionsHandler {
	return &TransactionsHandler{service, userService}
}

func (h *TransactionsHandler) RegisterRoutes(router *gin.Engine, fbAuthClient *auth.Client) {

	sGroup := router.Group("/transactions")
	sGroup.GET("/search", AuthHandler(fbAuthClient, h.UserService, h.SearchTransactions))
	sGroup.GET("/summary", AuthHandler(fbAuthClient, h.UserService, h.SummaryTransactions))
	sGroup.POST("/import", AuthHandler(fbAuthClient, h.UserService, h.ImportTransactions))
}

// Search implements Service.
func (h *TransactionsHandler) SearchTransactions(c *gin.Context) {

	user, err := getUser(c, h.UserService)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	searchText := c.Query("searchText")
	sstartDate := c.Query("startDate")
	startDate := utils.DateFromString(sstartDate)
	sendDate := c.Query("endDate")
	endDate := utils.DateFromString(sendDate)

	slog.Info("SearchTransactions started", "Startdate", startDate, "EndDate", endDate)

	txns, err := h.Service.SearchTransactions(*user, startDate, endDate, searchText)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
	}

	for _, txn := range txns {
		if strings.Compare(txn.Dbcr, "credit") == 0 {
			txn.CAmount += txn.Amount
		} else {
			txn.DAmount += txn.Amount
		}
	}
	slog.Info(fmt.Sprintf("SearchTransactions count: %d", len(txns)))

	c.JSON(http.StatusOK, txns)
}

func (h *TransactionsHandler) SummaryTransactions(c *gin.Context) {

	user, err := getUser(c, h.UserService)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	sstartDate := c.Query("startDate")
	startDate := utils.DateFromString(sstartDate)
	sendDate := c.Query("endDate")
	endDate := utils.DateFromString(sendDate)

	slog.Info("SummaryTransactions started", "Startdate", startDate, "EndDate", endDate)

	txns, err := h.Service.SummaryTransactions(*user, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
	}

	slog.Info(fmt.Sprintf("SummaryTransactions count: %d", len(txns)))
	c.JSON(http.StatusOK, txns)
}

func (h *TransactionsHandler) ImportTransactions(c *gin.Context) {

	user, err := getUser(c, h.UserService)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	sstartDate := c.Query("startDate")
	startDate := utils.DateFromString(sstartDate)
	sendDate := c.Query("endDate")
	endDate := utils.DateFromString(sendDate)

	var txns []*domain.Transaction
	err = json.NewDecoder(c.Request.Body).Decode(&txns)
	if err != nil {
		slog.Debug("TransactionsHandler", "ImportTransactions", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
	}

	err = h.Service.ImportTransactions(*user, startDate, endDate, txns)
	if err != nil {
		slog.Debug("TransactionsHandler", "ImportTransactions", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
	}
	slog.Info(fmt.Sprintf("ImportTransactions count: %d", len(txns)))
	c.JSON(http.StatusOK, "")

}
