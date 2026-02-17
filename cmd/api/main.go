package main

import (
	"context"
	"log"
	"os"

	"github.com/rkapps/fin-tracker-backend-go/internal/handlers"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"

	"github.com/rkapps/fin-tracker-backend-go/internal/logger"
	_ "github.com/rkapps/fin-tracker-backend-go/internal/migrations"
	"github.com/rkapps/fin-tracker-backend-go/internal/services"

	firebase "firebase.google.com/go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/storage-backend-go/migrations"
	mongodb "github.com/rkapps/storage-backend-go/mongodb"
)

const (
	FINANCE_DB_NAME = "test"
)

func main() {

	//Set logger
	logger.SetLogger()

	fbApp, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	fbAuthClient, err := fbApp.Auth(context.Background())
	if err != nil {
		log.Fatalf("Firebase authorization error: %v\n", err)
	}

	mongoConnStr := os.Getenv("MONGO_ATLAS_CONN_STR")
	log.Printf("MongoConnectionStr: %s", mongoConnStr)

	reg := mongodb.GetBsonRegistryForDecimal()
	database, err := mongodb.NewMongoDatabaseWithRegistry(mongoConnStr, FINANCE_DB_NAME, reg)
	if err != nil {
		log.Fatalf("Error connecting to Mongo DB: %v", err)
	}

	err = migrations.RunMigrations(database)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.New()
	router.Use(cors.Default())
	router.SetTrustedProxies(nil)

	// Register Handlers
	//Mongo Service
	storage := storage.NewMongoStorage(database)
	stocksService := services.NewStocksService(storage)
	portfoliosService := services.NewPortfoliosService(storage)

	//Stocks handler
	stocksHandler := handlers.NewStocksHandler(router, stocksService)
	stocksHandler.RegisterRoutes(router)

	//Portfolios handler
	portfoliosHandler := handlers.NewPortfoliosHandler(router, portfoliosService)
	portfoliosHandler.RegisterRoutes(router, fbAuthClient)

	port := ":8080"
	log.Printf("Listening on port %s", port)
	router.Run(port)

}
