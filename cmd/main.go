package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/yujihn/bank_API/internal/config"
	"github.com/yujihn/bank_API/internal/db"
	"github.com/yujihn/bank_API/internal/handler"
	"github.com/yujihn/bank_API/internal/middleware"
	"github.com/yujihn/bank_API/internal/repository"
	"github.com/yujihn/bank_API/internal/service"
)

// Выполнение миграции базы данных с помощью инструмента migrate
func runMigrations(dsn string) {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		logrus.Fatalf("Ошибка при инициализации миграций: %v", err)
	}

	err = m.Up()

	switch {
	case errors.Is(err, migrate.ErrNoChange):
		logrus.Info("Миграции не требуются, схема базы данных актуальна")
		return

	case err != nil:
		logrus.Fatalf("Ошибка при применении миграций: %v", err)
	}

	logrus.Info("Миграции успешно применены")
}

func main() {
	// Создание и настройка логгера
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// Загрузка конфигурационных данных
	ctx := context.Background()
	dbCfg := config.LoadDB()
	jwtCfg := config.LoadJWT()
	cryptoCfg := config.LoadCrypto()

	// Формирование DSN и запуск миграций базы данных
	dsn := db.BuildDSN(dbCfg)
	runMigrations(dsn)

	// Подключение к базе данных
	pool, err := db.New(ctx, dbCfg)
	if err != nil {
		logger.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer pool.Close()
	logger.Info("Подключение к базе данных успешно установлено")

	// Инициализация репозиториев для работы с данными
	userRepo := repository.NewUserRepository(pool)
	accountRepo := repository.NewAccountRepository(pool)
	transactionRepo := repository.NewTransactionRepository(pool)
	cardRepo := repository.NewCardRepository(pool)

	// Создание сервисов бизнес-логики
	authService := service.NewAuthService(userRepo, jwtCfg)
	accountService := service.NewAccountService(accountRepo, transactionRepo)
	cardService := service.NewCardService(cardRepo, pool, cryptoCfg.HMACKey)

	// Создание обработчиков HTTP-запросов
	authHandler := handler.NewAuthHandler(authService, logger)
	accountHandler := handler.NewAccountHandler(accountService, logger)
	cardHandler := handler.NewCardHandler(cardService, logger)

	// Middleware для проверки JWT токена
	jwtMiddleware := middleware.NewJWTMiddleware(authService, logger)

	// Настройка маршрутизации API
	r := mux.NewRouter().PathPrefix("/api").Subrouter()

	// Публичные маршруты
	r.HandleFunc("/register", authHandler.Register).Methods(http.MethodPost)
	r.HandleFunc("/login", authHandler.Login).Methods(http.MethodPost)

	// Защищенные маршруты (JWT авторизация)
	apiRouter := r.PathPrefix("").Subrouter()
	apiRouter.Use(jwtMiddleware.Middleware)

	// Маршруты для управления счетами
	apiRouter.HandleFunc("/accounts", accountHandler.CreateAccount).Methods(http.MethodPost)
	apiRouter.HandleFunc("/accounts", accountHandler.GetAccounts).Methods(http.MethodGet)
	apiRouter.HandleFunc("/accounts/{id}/balance", accountHandler.UpdateBalance).Methods(http.MethodPatch)
	apiRouter.HandleFunc("/accounts/{id}/transactions", accountHandler.GetTransactions).Methods(http.MethodGet)
	apiRouter.HandleFunc("/transfer", accountHandler.Transfer).Methods(http.MethodPost)

	// Маршруты для управления картами
	apiRouter.HandleFunc("/cards", cardHandler.CreateCard).Methods(http.MethodPost)
	apiRouter.HandleFunc("/cards", cardHandler.GetCards).Methods(http.MethodGet)
	apiRouter.HandleFunc("/cards/{id}", cardHandler.GetCardDetails).Methods(http.MethodGet)
	apiRouter.HandleFunc("/payments", cardHandler.ProcessPayment).Methods(http.MethodPost)

	// Настройка параметров HTTP-сервера
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Запуск сервера в отдельной горутине
	go func() {
		logger.Infof("Сервер запущен на порту %s", "8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Обработка сигналов завершения работы (например, Ctrl+C)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Завершение работы сервера...")

	// Ожидание завершения текущих обработок и корректное завершение сервера
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Fatalf("Ошибка при остановке сервера: %v", err)
	}
	logger.Info("Сервер успешно остановлен")
}
