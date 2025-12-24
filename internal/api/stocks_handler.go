package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"rkapps/fin-tracker-backend-go/internal/stocks"
	"strings"

	"github.com/gin-gonic/gin"
)

type StocksHandler struct {
	Service stocks.Service
}

func NewStocksHandler(router *gin.Engine, service stocks.Service) *StocksHandler {
	return &StocksHandler{
		Service: service,
	}
}

func (h *StocksHandler) RegisterRoutes(router *gin.Engine) {

	sGroup := router.Group("/stocks")
	sGroup.GET("", h.GetTickers)
	sGroup.GET("/:symbol", h.GetTicker)
	sGroup.POST("/load", h.LoadTickers)
	sGroup.POST("/search", h.SearchTickers)
	sGroup.GET("/updateEOD", h.UpdateEOD)
	sGroup.GET("/updateRealtime", h.UpdateRealtime)
}

// GetTicker gets a single ticker based on the symbol
func (h *StocksHandler) GetTicker(c *gin.Context) {
	symbol := c.Param("symbol")
	tk, err := h.Service.GetTicker(c, symbol)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, tk)
}

// GetTickers get multiple tickers by symbols
func (h *StocksHandler) GetTickers(c *gin.Context) {

	symbols := strings.Split(c.Query("symbols"), ",")
	slog.Debug("GetTickers", "Symbols", symbols)
	tks, _ := h.Service.GetTickers(c, symbols)
	c.JSON(http.StatusOK, tks)
}

// LoadTickers get multiple tickers by symbols
func (h *StocksHandler) LoadTickers(c *gin.Context) {

	var ts stocks.Tickers
	if err := c.BindJSON(&ts); err != nil {
		return
	}

	slog.Info("LoadTickers", "Loading", len(ts))
	if err := h.Service.LoadTickers(c, ts); err != nil {
		slog.Info("LoadTickers", "Error", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	slog.Info("LoadTickers", "Loaded", len(ts))
	c.JSON(http.StatusOK, ts)
}

// SearchTickers get multiple tickers by symbols
func (h *StocksHandler) SearchTickers(c *gin.Context) {
	var ts stocks.TickerSearch
	json.NewDecoder(c.Request.Body).Decode(&ts)
	slog.Info("SearchTickers", "TickerSearch", ts)
	tks, err := h.Service.SearchTicker(c, ts)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	slog.Info("SearchTickers", "Tickers", len(tks))
	c.JSON(http.StatusOK, tks)
}

// UpdateEOD updates all stocks with EOD data
func (h *StocksHandler) UpdateEOD(c *gin.Context) {
	h.Service.UpdateEOD(c)
}

// UpdateRealtime updates all stocks with realtime data
func (h *StocksHandler) UpdateRealtime(c *gin.Context) {
	h.Service.UpdateRealtime(c)
}
