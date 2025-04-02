package service

import (
	"Service/internal/models"
	"context"
)

type User interface {
	Register(ctx context.Context, user models.Users) error
	Login(ctx context.Context, user models.Users) (string, error)
}

type Service struct {
	User User
}

func NewService(user User) *Service {
	return &Service{
		User: user,
	}
}
