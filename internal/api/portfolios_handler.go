package api

import (
	"context"
	"rkapps/fin-tracker-backend-go/internal/portfolios"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/storage-backend-go/mongodb"
)

type PortfoliosHandler struct {
	client  *mongodb.MongoClient
	Service portfolios.PortfoliosService
}

func NewPortfoliosHandler(router *gin.Engine, client *mongodb.MongoClient, service portfolios.Service) *PortfoliosHandler {
	return &PortfoliosHandler{client: client}
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
	h.getUser(c)
}

func (h *PortfoliosHandler) getUser(c *gin.Context) context.Context {
	value, _ := c.Get("uid")
	uid := value.(string)
	user := mongodb.NewMongoRepository[*User](*h.client)
	u, _ := user.FindByID(c, uid)
	if u == nil {
		u := User{}
		u.ID = uid
		user.InsertOne(c, &u)
	}

	return context.WithValue(c, UserContextUID, u)
}
