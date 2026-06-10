package common

import (
	"log/slog"
	"os"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/services"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage/mongo"
	"github.com/rkapps/storage-backend-go/mongodb"
)

type ApiApp struct {
	Database            *mongodb.MongoDatabase
	UserService         services.UserService
	AccountsService     services.AccountsService
	TransactionsService services.TransactionsService
	PortfolioService    services.PortfolioService
}

type PipelineApp struct {
	Database         *mongodb.MongoDatabase
	UserService      services.UserService
	PortfolioService services.PortfolioService
}

func GetApiApp(trackerDbName string, financeDbName string, logConfig *logger.Config) (ApiApp, error) {

	uri := os.Getenv("FINTRACKER_MONGO_URI")
	database, err := getMongoDb(uri, trackerDbName)
	if err != nil {
		return ApiApp{}, err
	}
	// Create storeage
	storage := mongo.NewFinTrackerMongoStorage(database)
	accountsService := services.NewAccountsService(storage)
	userService := services.NewUserService(storage)
	transactionsService := services.NewTransactionsService(storage)

	uri = os.Getenv("FINANCE_MONGO_URI")
	database, err = getMongoDb(uri, financeDbName)
	if err != nil {
		return ApiApp{}, err
	}
	// create ticker storage
	tstorage := mongo.NewTickerMongoStorage(database)
	tickersService := services.NewStocksService(tstorage)
	portfolioService := services.NewPortfolioService(logConfig, tickersService, storage)

	return ApiApp{Database: database, UserService: userService,
		AccountsService:     accountsService,
		TransactionsService: transactionsService, PortfolioService: portfolioService,
	}, nil
}

func GetPipelineApp(trackerDbName string, financeDbName string, logConfig *logger.Config) (PipelineApp, error) {

	uri := os.Getenv("FINTRACKER_MONGO_URI")
	database, err := getMongoDb(uri, trackerDbName)
	if err != nil {
		return PipelineApp{}, err
	}
	// Create storeage
	storage := mongo.NewFinTrackerMongoStorage(database)
	userService := services.NewUserService(storage)

	uri = os.Getenv("FINANCE_MONGO_URI")
	database, err = getMongoDb(uri, financeDbName)
	if err != nil {
		return PipelineApp{}, err
	}
	// create ticker storage
	tstorage := mongo.NewTickerMongoStorage(database)
	tickersService := services.NewStocksService(tstorage)
	portfolioService := services.NewPortfolioService(logConfig, tickersService, storage)

	return PipelineApp{Database: database, UserService: userService, PortfolioService: portfolioService}, nil
}

func getMongoDb(uri string, dbname string) (*mongodb.MongoDatabase, error) {
	slog.Info("MongoDb connection string: " + uri)
	reg := mongodb.GetBsonRegistryForDecimal()
	return mongodb.NewMongoDatabaseWithRegistry(uri, dbname, reg)
}
