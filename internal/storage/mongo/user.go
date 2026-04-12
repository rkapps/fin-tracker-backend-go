package mongo

import (
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

func (s MongoStorage) GetUser(id string) (*domain.User, error) {
	return s.users().FindByID(s.context(), id)
}

func (s MongoStorage) SaveUser(user *domain.User) error {
	err := s.users().UpdateOne(s.context(), user)
	if err != nil {
		err = s.users().InsertOne(s.context(), user)
	}
	return err
}
