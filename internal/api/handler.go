package api

import (
	"Service/internal/models"
	"Service/internal/server"
	"Service/my_project/generated/auth"
	"context"
	"go.uber.org/zap"
)

type AuthHandler struct {
	user service.User
	log  *zap.SugaredLogger
	auth.UnimplementedAuthServiceServer
}

func NewAuthHandler(user service.User, log *zap.SugaredLogger) *AuthHandler {
	return &AuthHandler{
		user: user,
		log:  log,
	}
}

func (s *AuthHandler) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {

	var user models.Users
	user.Username = req.GetUsername()
	user.Password = req.GetPassword()
	//user.Email = req.GetEmail()

	err := s.user.Register(ctx, user)
	if err != nil {
		s.log.Errorw("error to create user in service", "error", err)
		return nil, err
	}

	s.log.Infow("register user in service", "username", user.Username)

	return &auth.RegisterResponse{Message: "User registered successfully"}, nil
}

func (s *AuthHandler) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	var user models.Users
	user.Username = req.GetUsername()
	user.Password = req.GetPassword()

	token, err := s.user.Login(ctx, user)
	if err != nil {
		s.log.Errorw("error to login in service", "error", err)
		return nil, err
	}

	s.log.Infow("login in service", "username", user.Username)

	return &auth.LoginResponse{Token: token}, nil
}
