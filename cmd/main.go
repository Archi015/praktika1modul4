package main

import (
	"Service/internal/api"
	"Service/internal/config"
	logging "Service/internal/logger"
	"Service/internal/repo"
	service "Service/internal/server"
	auth "Service/my_project/generated/auth"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

func main() {

	//енв файл
	if err := godotenv.Load(config.EnvPath); err != nil {
		log.Fatal("No .env file found")
	}

	//считывание информации в config
	var cfg config.AppConfig
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal("Error processing config", zap.Error(err))
	}

	//создаем свой логгер
	logger, err := logging.NewLogger(cfg.LogLevel)
	if err != nil {
		log.Fatal("Error init logger", zap.Error(err))
	}

	//подключение к БД
	pool, err := repo.Connection(context.Background(), cfg.PostgreSQL)
	if err != nil {
		log.Fatal("Error connect to database", zap.Error(err))
	}

	//проверка БД
	if err := repo.CheckConnection(pool, logger); err != nil {
		log.Fatalf("Connection check failed: %v", err)
	}

	//создание структуры пользователя для репозитория
	userRepository := repo.NewUserRepository(pool)
	repos := repo.NewRepository(userRepository)

	//создание структуры пользователя для сервиса
	userService := service.NewUserService(repos.User, logger)
	services := service.NewService(userService)

	//создание уровня хендлер
	handlers := api.NewAuthHandler(services.User, logger)

	server := grpc.NewServer()

	auth.RegisterAuthServiceServer(server, handlers)

	lis, err := net.Listen("tcp", cfg.Grpc.Port)
	if err != nil {
		log.Fatal(errors.Wrap(err, "error initializing server"))
	}

	go func() {
		fmt.Println("starting grpc server on port", cfg.Grpc.Port, time.Now())
		if err := server.Serve(lis); err != nil {
			log.Fatal(errors.Wrap(err, "error initializing grpc server"))
		}
	}()

	//завершение процесса
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	if err := repo.CloseConnection(pool); err != nil {
		log.Fatal(errors.Wrap(err, "error closing connection"))
	}

	fmt.Println("server stopped")
}
