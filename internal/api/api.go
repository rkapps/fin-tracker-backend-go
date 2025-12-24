package api

import (
	"rkapps/fin-tracker-backend-go/internal/stocks"

	"github.com/gin-gonic/gin"
	"github.com/rkapps/storage-backend-go/mongodb"
)

func RegisterHandlers(router *gin.Engine, client *mongodb.MongoClient) error {

	//Mongo Service
	stocksService := stocks.NewMongoService(client)

	//Stocks handler
	stocksHandler := NewStocksHandler(router, stocksService)
	stocksHandler.RegisterRoutes(router)

	return nil
}
