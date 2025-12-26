package portfolios

import "github.com/rkapps/storage-backend-go/mongodb"

type PortfoliosService struct {
	client *mongodb.MongoClient
}

func NewMongoService(client *mongodb.MongoClient) Service {

	return PortfoliosService{
		client: client,
	}
}
