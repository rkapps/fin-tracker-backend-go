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

func (s UserService) GetUser(id string) (*domain.User, error) {
	return s.storage.GetUser(id)
}

func (s UserService) SaveUser(user *domain.User) error {
	return s.storage.SaveUser(user)
}
