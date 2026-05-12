package common

import (
	"log/slog"
	"os"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/services"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
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

func GetApiApp(dbname string, logConfig *logger.Config) (ApiApp, error) {

	database, err := getMongoDb(dbname)
	if err != nil {
		return ApiApp{}, err
	}
	// Create storeage
	storage := mongo.NewMongoStorage(database)
	accountsService := services.NewAccountsService(storage)
	userService := services.NewUserService(storage)
	transactionsService := services.NewTransactionsService(storage)
	tickersService := services.NewStocksService(storage)
	portfolioService := services.NewPortfolioService(logConfig, tickersService, storage)

	return ApiApp{Database: database, UserService: userService,
		AccountsService:     accountsService,
		TransactionsService: transactionsService, PortfolioService: portfolioService,
	}, nil
}

func GetPipelineApp(logConfig *logger.Config) (PipelineApp, error) {
	database, err := getMongoDb(storage.FINTRACKER_DB_NAME)
	if err != nil {
		return PipelineApp{}, err
	}
	// Create storeage
	storage := mongo.NewMongoStorage(database)
	userService := services.NewUserService(storage)
	tickersService := services.NewStocksService(storage)
	portfolioService := services.NewPortfolioService(logConfig, tickersService, storage)

	return PipelineApp{Database: database, UserService: userService, PortfolioService: portfolioService}, nil
}

func getMongoDb(dbname string) (*mongodb.MongoDatabase, error) {
	mongoConnStr := os.Getenv("MONGO_URI")
	// log.Printf("MongoConnectionStr: %s", mongoConnStr)
	mongoConnStr = "mongodb://localhost:33333/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+2.5.10"
	slog.Info("MongoDb connection string: " + mongoConnStr)

	reg := mongodb.GetBsonRegistryForDecimal()
	return mongodb.NewMongoDatabaseWithRegistry(mongoConnStr, dbname, reg)
}
