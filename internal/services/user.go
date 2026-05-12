package services

import (
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
)

type UserService struct {
	storage storage.StorageService
}

func NewUserService(storage storage.StorageService) UserService {
	return UserService{storage: storage}
}

func (s UserService) GetUsers() []*domain.User {
	return s.storage.GetUsers()
}

func (s UserService) GetUser(id string) (*domain.User, error) {
	return s.storage.GetUser(id)
}

func (s UserService) SaveUser(user *domain.User) error {

	// apply defaults if not set
	if user.LotMatchingMethod == "" {
		user.LotMatchingMethod = domain.LotMatchingHIFO
	}
	if user.CurrencyCode == "" {
		user.CurrencyCode = "USD"
	}
	return s.storage.SaveUser(user)
}
