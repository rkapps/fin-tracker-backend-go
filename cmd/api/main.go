package main

import (
	"context"
	"log"
	"os"
	"rkapps/fin-tracker-backend-go/internal/api"
	"rkapps/fin-tracker-backend-go/internal/logger"
	_ "rkapps/fin-tracker-backend-go/internal/migrations"

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
	mongoConnStr = "mongodb://localhost:33333/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.5.10"

	reg := mongodb.GetBsonRegistryForDecimal()
	client, err := mongodb.NewMongoClientWithRegistry(mongoConnStr, FINANCE_DB_NAME, reg)
	if err != nil {
		log.Fatalf("Error connecting to Mongo DB: %v", err)
	}

	err = migrations.RunMigrations(client)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.New()
	router.Use(cors.Default())
	router.SetTrustedProxies(nil)

	// Register Handlers
	api.RegisterHandlers(router, client, fbAuthClient)

	port := ":8080"
	log.Printf("Listening on port %s", port)
	router.Run(port)

}
