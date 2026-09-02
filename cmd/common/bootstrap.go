package common

import (
	"encoding/base64"
	"log"
	"os"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto/cardano"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto/coinbase"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto/ethereum"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto/kraken"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto/solana"
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

func GetApiApp(
	trackerDbName string,
	financeDbName string,
	logConfig *logger.Config,
) (ApiApp, error) {

	blog := logConfig.For("api.bootstrap")

	uri := os.Getenv("FINTRACKER_MONGO_URI")
	database, err := getMongoDb(uri, trackerDbName, blog)
	if err != nil {
		return ApiApp{}, err
	}

	syncRegistry := getSyncProviderRegistry(logConfig)
	transformerRegistry := getTransformerRegistry(logConfig)
	c := getEncryptionService()

	// Create storeage
	storage := mongo.NewFinTrackerMongoStorage(database)
	// providerService := services.NewProviderService(storage)
	userService := services.NewUserService(storage)
	transactionsService := services.NewTransactionsService(storage)

	// create account service
	accountsService := services.NewAccountsService(storage, storage, syncRegistry, c, logConfig)

	uri = os.Getenv("RUSTIC_FINANCE_MONGO_URI")
	rdatabase, err := getMongoDb(uri, financeDbName, blog)
	if err != nil {
		return ApiApp{}, err
	}
	// create ticker storage
	tstorage := mongo.NewTickerMongoStorage(rdatabase)
	tickersService := services.NewStocksService(tstorage)
	portfolioService := services.NewPortfolioService(userService, storage, tickersService,
		storage, syncRegistry, transformerRegistry, c, logConfig,
	)

	return ApiApp{Database: database, UserService: userService,
		AccountsService:     accountsService,
		TransactionsService: transactionsService, PortfolioService: portfolioService,
	}, nil
}

func GetPipelineApp(
	trackerDbName string,
	financeDbName string,
	logConfig *logger.Config,
) (PipelineApp, error) {

	blog := logConfig.For("pipeline.bootstrap")

	uri := os.Getenv("FINTRACKER_MONGO_URI")
	database, err := getMongoDb(uri, trackerDbName, blog)
	if err != nil {
		return PipelineApp{}, err
	}

	syncRegistry := getSyncProviderRegistry(logConfig)
	transformerRegistry := getTransformerRegistry(logConfig)
	c := getEncryptionService()

	// Create storeage
	storage := mongo.NewFinTrackerMongoStorage(database)
	userService := services.NewUserService(storage)
	// providerService := services.NewProviderService(storage)

	uri = os.Getenv("RUSTIC_FINANCE_MONGO_URI")
	rdatabase, err := getMongoDb(uri, financeDbName, blog)
	if err != nil {
		return PipelineApp{}, err
	}
	blog.Info("GetPipelineApp", "Database", rdatabase)

	// create ticker storage
	tstorage := mongo.NewTickerMongoStorage(rdatabase)
	tickersService := services.NewStocksService(tstorage)
	blog.Info("GetPipelineApp", "tstorage", storage)

	portfolioService := services.NewPortfolioService(storage, storage, tickersService,
		storage, syncRegistry, transformerRegistry, c, logConfig,
	)

	return PipelineApp{Database: database, UserService: userService, PortfolioService: portfolioService}, nil
}

func getMongoDb(uri string, dbname string, logger *logger.Logger) (*mongodb.MongoDatabase, error) {
	logger.Info("getMongoDb", "Connection string: ", uri, "DbName", dbname)
	reg := mongodb.GetBsonRegistryForDecimal()
	return mongodb.NewMongoDatabaseWithRegistry(uri, dbname, reg)
}

func getSyncProviderRegistry(logConfig *logger.Config) core.SyncRegistry {

	blockfrost_project_id := os.Getenv("CARDANO_BLOCKFROST_PROJECT_ID")
	alchemy_api_key := os.Getenv("SOLANA_ALCHEMY_API_KEY")
	etherscan_api_key := os.Getenv("ETHERSCAN_API_KEY")

	registry := core.NewSyncRegistry()
	registry.Register(coinbase.New(coinbase.NewHTTPClient(), logConfig))
	registry.Register(cardano.New(cardano.NewBlockFrostHTTPClient(blockfrost_project_id), logConfig))
	registry.Register(kraken.New(kraken.NewHTTPClient(), logConfig))
	registry.Register(solana.New(solana.NewAlchemyHTTPClient(alchemy_api_key), logConfig))

	//ethereum
	registry.Register(ethereum.NewEthereum(ethereum.NewEtherscanClient(etherscan_api_key), logConfig))
	return *registry
}

func getTransformerRegistry(logConfig *logger.Config) core.TransformerRegistry {
	registry := core.NewTransformerRegistry()
	registry.Register(coinbase.NewCoinbaseAccountTransformer(logConfig))
	registry.Register(kraken.NewKrakenAccountTransformer(logConfig))
	registry.Register(cardano.NewCardanoAccountTransformer(logConfig))
	registry.Register(solana.NewSolanaAccountTransformer(logConfig))
	registry.Register(ethereum.NewEthereumTransformer(logConfig))
	return *registry
}

func getEncryptionService() core.EncryptionService {
	key, _ := base64.StdEncoding.DecodeString(os.Getenv("FINTRACKER_CREDENTIAL_ENCRYPTION_KEY"))
	encryptionService, err := core.NewAESGCMEncryptionService(key)
	if err != nil {
		log.Fatal(err)
	}
	return encryptionService
}
