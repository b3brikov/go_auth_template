package main

import (
	jwtman "auth_service/internal/JWT/access"
	"auth_service/internal/config"
	grpccontroller "auth_service/internal/grpc_controller"
	"auth_service/internal/logger"
	"auth_service/internal/services/auth"
	redis "auth_service/internal/storage/Redis"
	postgresstorage "auth_service/internal/storage/postgresStorage"
	"auth_service/protos/gen/go/authservicegen"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

func main() {
	// Загружаем конфиг
	cfg := config.MustLoad()

	// Запуск логгера
	logger, err := logger.NewLogger()
	if err != nil {
		panic("Не удалось инициализировать логгер: " + err.Error())
	}

	// Подключение к Postgres
	storage := postgresstorage.NewPostgres(cfg.Storage_path, logger)

	// Подключение к Redis
	rds := redis.NewRedisClient()
	logger.Info("Redis подключен успешно")

	// JWT менеджер
	jwt := &jwtman.JWTManager{
		SecretKey:     []byte("test"),
		TokenDuration: 15 * time.Minute,
	}

	// Auth сервис
	authSvc := auth.NewAuth(logger, storage, rds, jwt)

	// // Контроллер
	// controller := controller.NewController(authSvc, logger)

	// HTTP сервер
	// srv := server.NewServer(controller, logger)

	// Канал для ловли сигнала остановки
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// srv.Start()

	grpcServer := grpc.NewServer()

	authservicegen.RegisterAuthServiceServer(
		grpcServer,
		grpccontroller.NewGRPCController(authSvc, logger),
	)

	go func() {
		lis, _ := net.Listen("tcp", ":50051")
		grpcServer.Serve(lis)
	}()

	<-stop
	logger.Info("Получен сигнал завершения, останавливаем сервер...")

	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	// if err := srv.Shutdown(ctx); err != nil {
	// 	logger.Error("Ошибка при завершении сервера", slog.Any("error", err))
	// }
	grpcServer.GracefulStop()

	logger.Info("Сервер корректно остановлен")
}
