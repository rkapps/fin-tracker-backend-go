package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/services"
	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
)

type PortfolioHandler struct {
	Service   services.PortfolioService
	logConfig *logger.Config
	logger    *logger.Logger
}

func NewPortfolioHandler(router *gin.Engine, logConfig *logger.Config, service services.PortfolioService) *PortfolioHandler {
	plog := logConfig.For("portfolio.handler")
	return &PortfolioHandler{service, logConfig, plog}
}

func (p *PortfolioHandler) RegisterRoutes(router *gin.Engine, fbAuthClient *auth.Client) {

	sGroup := router.Group("/portfolio")
	sGroup.GET("/summary", AuthHandler(fbAuthClient, p.GetSummary))
	sGroup.GET("/holdings", AuthHandler(fbAuthClient, p.GetHoldings))
	sGroup.GET("/income", AuthHandler(fbAuthClient, p.GetIncome))
	sGroup.GET("/gainloss", AuthHandler(fbAuthClient, p.GetGainLoss))
	sGroup.GET("/activities", AuthHandler(fbAuthClient, p.GetActivities))

}

// GetSummary gets the accounts in the portfolio
func (p *PortfolioHandler) GetSummary(c *gin.Context) {
	uid, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	asumys, err := p.Service.GetSummary(uid)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, asumys)

}

// GetAccounts gets the accounts in the portfolio
func (p *PortfolioHandler) GetHoldings(c *gin.Context) {
	uid, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	category := c.Query("category")
	atype := c.Query("type")
	ids := c.Query("acctIds")
	var acctIds []string
	if len(ids) > 0 {
		ids = utils.TrimCommas(ids)
		acctIds = strings.Split(ids, ",")
	}

	p.logger.Info("GetHoldings", "Category-type", fmt.Sprintf("%s-%s", category, atype), "AcctIds", acctIds)

	hldgs, err := p.Service.GetHoldings(uid, category, atype, acctIds)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, hldgs)

}

// GetActivities gets the activities in the portfolio
func (p *PortfolioHandler) GetActivities(c *gin.Context) {
	uid, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	ids := c.Query("acctIds")
	var acctIds []string
	if len(ids) > 0 {
		ids = utils.TrimCommas(ids)
		acctIds = strings.Split(ids, ",")
	}

	category := c.Query("category")
	atype := c.Query("type")
	sStartDate := c.Query("startDate")

	startDate := time.Time{}
	endDate := time.Time{}
	if len(sStartDate) > 0 {
		startDate = utils.DateFromString(sStartDate)
	}

	p.logger.Info("GetActivities", "Category-Type", fmt.Sprintf("%s-%s-%v", category, atype, startDate), "AcctIds", acctIds)

	actvs, err := p.Service.GetActivities(uid, category, atype, acctIds, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, actvs)

}

// GetIncome gets the income for the portfolio
func (p *PortfolioHandler) GetIncome(c *gin.Context) {
	uid, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	ids := c.Query("acctIds")
	var acctIds []string
	if len(ids) > 0 {
		ids = utils.TrimCommas(ids)
		acctIds = strings.Split(ids, ",")
	}

	sYear := c.Query("year")
	year := 0
	if len(sYear) > 0 {
		year, _ = strconv.Atoi(sYear)
	}

	startDate := time.Time{}
	endDate := time.Time{}

	if year > 0 {
		startDate = utils.TruncateToStartOfYear(year)
		endDate = utils.TruncateToEndOYear(year)
	}

	category := c.Query("category")
	atype := c.Query("type")
	p.logger.Info("GetIncome", "Category-Type", fmt.Sprintf("%s-%s", category, atype), "AcctIds", acctIds)
	incomes, err := p.Service.GetIncome(uid, category, atype, acctIds, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, incomes)

}

// GetIncome gets the gainloss for the portfolio
func (p *PortfolioHandler) GetGainLoss(c *gin.Context) {
	_, err := getUID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	ids := c.Query("acctIds")
	var acctIds []string
	if len(ids) > 0 {
		ids = utils.TrimCommas(ids)
		acctIds = strings.Split(ids, ",")
	}

	category := c.Query("category")
	atype := c.Query("type")
	p.logger.Info("GetGainloss", "Category-Type", fmt.Sprintf("%s-%s", category, atype), "AcctIds", acctIds)

	c.JSON(http.StatusOK, "")

}
