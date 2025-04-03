package repo

import (
	"Service/internal/models"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

// Слой репозитория, здесь должны быть все методы, связанные с базой данных

// SQL-запрос на вставку задачи (sql-запросы в константе)
const (
	createUserQuery  = `INSERT INTO users (username, password_hash, email) VALUES ($1, $2, $3)`
	loginQuery       = `SELECT is active FROM users WHERE username = $1 and password hash = $2`
	checkExistsQuery = `SELECT username FROM users WHERE username = $1 OR email = $2`
)

type userRepository struct {
	pool *pgxpool.Pool
}

type User interface {
	Register(ctx context.Context, user models.Users) error
	Login(ctx context.Context, user models.Users) (bool, error)
	CheckExists(ctx context.Context, user models.Users) bool
}

type Repository struct {
	User User
}

func NewRepository(user User) *Repository {
	return &Repository{
		User: user,
	}
}

// Userepository - интерфейс с методом создания задачи

func NewUserRepository(pool *pgxpool.Pool) User {
	return &userRepository{pool: pool}
}

func (u *userRepository) Register(ctx context.Context, user models.Users) error {
	_, err := u.pool.Exec(ctx, createUserQuery, user.Username, user.Password, user.Email)
	if err != nil {
		log.Printf("Error register user: %v", err)
		return err
	}
	return nil
}

func (u *userRepository) CheckExists(ctx context.Context, user models.Users) bool {
	var username string
	_ = u.pool.QueryRow(ctx, checkExistsQuery, user.Username, user.Email).Scan(&username)
	if username == "" {
		log.Printf("User %s does not exist", user.Username)
		return true
	}

	return false
}

func (u *userRepository) Login(ctx context.Context, user models.Users) (bool, error) {
	var active bool
	err := u.pool.QueryRow(ctx, loginQuery, user.Username, user.Password).Scan(&active)
	if err != nil {
		log.Printf("Error login or password: %v", err)
		return false, err
	}

	return active, nil
}
