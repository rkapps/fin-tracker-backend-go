package mongo

import (
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s FinTrackerMongoStorage) GetUsers() []*domain.User {
	users, err := s.users().Find(s.context(), bson.M{}, bson.D{}, 0, 0)
	if err != nil {
		return []*domain.User{}
	} else {
		return users
	}
}

func (s FinTrackerMongoStorage) GetUser(id string) (*domain.User, error) {
	return s.users().FindByID(s.context(), id)
}

func (s FinTrackerMongoStorage) SaveUser(user *domain.User) error {
	err := s.users().UpdateOne(s.context(), user)
	if err != nil {
		err = s.users().InsertOne(s.context(), user)
	}
	return err
}
