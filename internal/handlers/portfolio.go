package handlers

import (
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/fin-tracker-backend-go/internal/services"
)

type PortfolioHandler struct {
	Service services.PortfolioService
}

func NewPortfolioHandler(router *gin.Engine, service services.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{service}
}

func (p *PortfolioHandler) RegisterRoutes(router *gin.Engine, fbAuthClient *auth.Client) {

	sGroup := router.Group("/portfolio")
	sGroup.GET("/holdings", AuthHandler(fbAuthClient, p.GetHoldings))
	sGroup.GET("/activities", AuthHandler(fbAuthClient, p.GetActivities))

}

// GetAccounts gets the accounts in the portfolio
func (p *PortfolioHandler) GetHoldings(c *gin.Context) {
	// h.getUser(c)
	uid, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	hldgs, err := p.Service.GetHoldings(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, hldgs)

}

// GetActivities gets the activities in the portfolio
func (p *PortfolioHandler) GetActivities(c *gin.Context) {
	// h.getUser(c)
	uid, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	actvs, err := p.Service.GetActivities(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, actvs)

}
