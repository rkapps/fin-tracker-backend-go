package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"rkapps/fin-tracker-backend-go/internal/portfolios"
	"rkapps/fin-tracker-backend-go/internal/portfolios/accounts"
	"rkapps/fin-tracker-backend-go/internal/portfolios/user"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/storage-backend-go/mongodb"
)

type PortfoliosHandler struct {
	client  *mongodb.MongoClient
	Service portfolios.Service
}

func NewPortfoliosHandler(router *gin.Engine, client *mongodb.MongoClient, service portfolios.Service) *PortfoliosHandler {
	return &PortfoliosHandler{client: client, Service: service}
}

func (h *PortfoliosHandler) RegisterRoutes(router *gin.Engine, fbauthclient *auth.Client) {

	sGroup := router.Group("/portfolios")
	sGroup.GET("/accounts", AuthHandler(fbauthclient, h.GetAccounts))
	sGroup.POST("/accounts/load", AuthHandler(fbauthclient, h.LoadAccounts))
}

// GetAccounts gets the accounts in the portfolio
func (h *PortfoliosHandler) GetAccounts(c *gin.Context) {
	h.getUser(c)
}

// LoadAccounts gets the accounts in the portfolio
func (h *PortfoliosHandler) LoadAccounts(c *gin.Context) {

	user := h.getUser(c)
	var accts accounts.Accounts
	err := json.NewDecoder(c.Request.Body).Decode(&accts)
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
	slog.Debug("LoadAccounts", "Accounts", len(accts))

	err = h.Service.LoadAccounts(c, *user, accts)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

}

func (h *PortfoliosHandler) getUser(c *gin.Context) *user.User {
	value, _ := c.Get("uid")
	uid := value.(string)
	userColl := mongodb.NewMongoRepository[*user.User](*h.client)
	u, _ := userColl.FindByID(c, uid)
	if u == nil {
		u := user.User{}
		u.ID = uid
		userColl.InsertOne(c, &u)
	}

	return u
}
