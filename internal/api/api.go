package api

import (
	"rkapps/fin-tracker-backend-go/internal/portfolios"
	"rkapps/fin-tracker-backend-go/internal/stocks"

	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/storage-backend-go/mongodb"
)

func RegisterHandlers(router *gin.Engine, client *mongodb.MongoClient, fbauthclient *auth.Client) error {

	//Mongo Service
	stocksService := stocks.NewMongoService(client)
	portfoliosService := portfolios.NewMongoService(client)

	//Stocks handler
	stocksHandler := NewStocksHandler(router, stocksService)
	stocksHandler.RegisterRoutes(router)

	//Portfolios handler
	portfoliosHandler := NewPortfoliosHandler(router, client, portfoliosService)
	portfoliosHandler.RegisterRoutes(router, fbauthclient)

	return nil
}
