package handlers

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/portfolios/accounts"
	"github.com/rkapps/fin-tracker-backend-go/internal/services"
)

type PortfoliosHandler struct {
	Service services.PortfoliosService
}

func NewPortfoliosHandler(router *gin.Engine, service services.PortfoliosService) *PortfoliosHandler {
	return &PortfoliosHandler{service}
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

func (h *PortfoliosHandler) getUser(c *gin.Context) *domain.User {
	value, _ := c.Get("uid")
	uid := value.(string)

	// userColl := mongodb.GetMongoRepository[string, *domain.User](h.database)
	u, err := h.Service.GetUser(uid)
	if err != nil {
		// c.JSON(http.StatusBadRequest, err)
		return nil
	}
	// u, _ := userColl.FindByID(c, uid)
	// if u == nil {
	// 	u := domain.User{}
	// 	u.ID = uid
	// 	userColl.InsertOne(c, &u)
	// }
	return u
}
