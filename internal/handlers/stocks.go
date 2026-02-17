package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/services"
)

type StocksHandler struct {
	Service services.StocksService
}

func NewStocksHandler(router *gin.Engine, service services.StocksService) *StocksHandler {
	return &StocksHandler{
		Service: service,
	}
}

func (h *StocksHandler) RegisterRoutes(router *gin.Engine) {

	sGroup := router.Group("/stocks")
	sGroup.GET("", h.GetTickers)
	sGroup.GET("/:symbol", h.GetTicker)
	sGroup.GET("/:symbol/history", h.GetTickerHistory)
	sGroup.GET("/groups", h.GetTickerGroups)
	sGroup.POST("/load", h.LoadTickers)
	sGroup.POST("/search", h.SearchTickers)
	// sGroup.GET("/update", h.UpdateTickers)
	// sGroup.GET("/updateEOD", h.UpdateEOD)
	// sGroup.GET("/updateRealtime", h.UpdateRealtime)
}

// GetTicker gets a single ticker based on the symbol
func (h *StocksHandler) GetTicker(c *gin.Context) {
	symbol := c.Param("symbol")
	tk, err := h.Service.GetTicker(c, symbol)
	if err != nil {
		slog.Error("GetTicker", "Error", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, tk)
}

// GetTickerGroups returns ticker groups based on sector and industry
func (h *StocksHandler) GetTickerGroups(c *gin.Context) {
	tgs, err := h.Service.GetTickerGroups(c)
	if err != nil {
		slog.Error("GetTickerGroups", "Error", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, tgs)
}

// GetTicker gets a single ticker based on the symbol
func (h *StocksHandler) GetTickerHistory(c *gin.Context) {

	symbol := c.Param("symbol")
	slog.Info("GetTickerHistory", "symbol", symbol)
	tk, err := h.Service.GetTickerHistory(c, symbol)
	if err != nil {
		slog.Error("GetTickerHistory", "Error", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, tk)
}

// GetTickers get multiple tickers by symbols
func (h *StocksHandler) GetTickers(c *gin.Context) {

	symbols := strings.Split(c.Query("symbols"), ",")
	slog.Info("GetTickers", "Symbols", symbols)
	tks, err := h.Service.GetTickers(c, symbols)
	if err != nil {
		slog.Error("GetTickers", "Error", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, tks)
}

// LoadTickers get multiple tickers by symbols
func (h *StocksHandler) LoadTickers(c *gin.Context) {

	var ts domain.Tickers
	if err := c.BindJSON(&ts); err != nil {
		slog.Info("LoadTickers", "Error", err)
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
	var ts domain.TickerSearch
	json.NewDecoder(c.Request.Body).Decode(&ts)
	slog.Info("SearchTickers", "TickerSearch", ts)
	tks, err := h.Service.SearchTicker(c, ts)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, tks)
}

// // UpdateEOD updates all stocks with EOD data
// func (h *StocksHandler) UpdateEOD(c *gin.Context) {
// 	h.Service.UpdateEOD(c)
// }

// // UpdateRealtime updates all stocks with realtime data
// func (h *StocksHandler) UpdateRealtime(c *gin.Context) {
// 	h.Service.UpdateRealtime(c)
// }

// // UpdateTickers get multiple tickers by symbols
// func (h *StocksHandler) UpdateTickers(c *gin.Context) {

// 	symbols := strings.Split(c.Query("symbols"), ",")
// 	slog.Debug("GetTickers", "Symbols", symbols)
// 	err := h.Service.UpdateTickers(c, symbols)
// 	if err != nil {
// 		c.Error(err)
// 		return
// 	}
// 	c.JSON(http.StatusOK, nil)
// }
