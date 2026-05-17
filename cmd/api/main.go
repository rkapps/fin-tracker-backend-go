package main

import (
	"context"
	"os"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common"
	logger "github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/handlers"

	_ "github.com/rkapps/fin-tracker-backend-go/internal/migrations"

	firebase "firebase.google.com/go"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rkapps/storage-backend-go/migrations"
)

func main() {

	//Set logger

	logConfig := logger.New()
	mlog := logConfig.For("main")

	mlog.Logger.Info("main", "Lotgevel", logConfig)
	config := &firebase.Config{
		ProjectID: os.Getenv("FINTRACKER_PROJECT_ID"),
	}
	dbname := os.Getenv("FINTRACKER_DB_NAME")
	if len(dbname) == 0 {
		mlog.Error("FINTRACKER_DB_NAME environment variable not set.")
		os.Exit(1)
	}

	fbApp, err := firebase.NewApp(context.Background(), config)
	if err != nil {
		mlog.Error("Initializing app", "error", err)
		os.Exit(1)
	}

	fbAuthClient, err := fbApp.Auth(context.Background())
	if err != nil {
		mlog.Error("Firebase authorization", "error", err)
		os.Exit(1)
	}

	apiApp, err := common.GetApiApp(dbname, logConfig)
	if err != nil {
		mlog.Error("GetApiApp", "error", err)
		os.Exit(1)
	}
	err = migrations.RunMigrations(apiApp.Database)
	if err != nil {
		mlog.Error("RunMigrations", "error", err)
		os.Exit(1)
	}

	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:4200",
			"https://fin-tracker-rkapps.web.app",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		AllowCredentials: true,
	}))
	router.SetTrustedProxies(nil)

	//Portfolio handler
	portfolioHanlder := handlers.NewPortfolioHandler(router, logConfig, apiApp.PortfolioService)
	portfolioHanlder.RegisterRoutes(router, fbAuthClient)

	// accounts handler
	accountsHandler := handlers.NewAccountsHandler(router, apiApp.AccountsService)
	accountsHandler.RegisterRoutes(router, fbAuthClient)

	transactionsHandler := handlers.NewTransactionsHandler(router, apiApp.TransactionsService)
	transactionsHandler.RegisterRoutes(router, fbAuthClient)

	userHandler := handlers.NewUserHandler(router, apiApp.UserService)
	userHandler.RegisterRoutes(router, fbAuthClient)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local dev
	}
	mlog.Info("Server", "Listening on port", port)
	router.Run(":" + port)

}
