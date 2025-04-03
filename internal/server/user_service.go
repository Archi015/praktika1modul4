package service

import (
	"Service/internal/models"
	"Service/internal/repo"
	"context"
	"errors"
	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type AuthService struct {
	repo      repo.Repository
	secretKey string
}

// UserServiceServer реализует интерфейс gRPC.
type userService struct {
	repo        repo.User
	log         *zap.SugaredLogger
	authService *AuthService
}

// Метод регистрации нового пользователя
func NewUserService(repo repo.User, log *zap.SugaredLogger) User {
	return &userService{
		repo: repo,
		log:  log,
	}
}

func (u *userService) Register(ctx context.Context, user models.Users) error {

	ok := u.repo.CheckExists(ctx, user)
	if !ok {
		return errors.New("user already created")
	}

	user.Password = generateHashPassword(user.Password)

	err := u.repo.Register(ctx, user)
	if err != nil {
		u.log.Errorw("error to create user in db", "error", err)
		return err
	}

	return nil
}

func (s *AuthService) generateJWT(user models.Users) (string, error) {
	claims := jwt.MapClaims{
		"sub":      user.Username,
		"username": user.Username,
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (u *userService) Login(ctx context.Context, user models.Users) (string, error) {
	// Хешируем пароль пользователя
	user.Password = generateHashPassword(user.Password)

	// Проверка, существует ли пользователь
	active, err := u.repo.Login(ctx, user)
	if err != nil {
		u.log.Errorw("error to login", "error", err)
		return "", err
	}

	// Если пользователь не активен, возвращаем ошибку
	if !active {
		u.log.Infow("user not active")
		return "", errors.New("can't login")
	}

	// Генерация токена
	// Для этого необходимо использовать экземпляр AuthService, у которого есть метод generateJWT
	token, err := u.authService.generateJWT(user) // Предполагается, что у вас есть поле authService в userService
	if err != nil {
		u.log.Errorw("error to generate JWT", "error", err)
		return "", err
	}

	return token, nil
}

func generateHashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(hash)
}
