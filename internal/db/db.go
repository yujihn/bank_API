package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yujihn/bank_API/internal/config"
)

// BuildDSN формирует строку подключения к базе данных PostgreSQL на основе конфигурации
func BuildDSN(cfg config.DBConfig) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode)
}

// New создает пул соединений с базой данных PostgreSQL
func New(ctx context.Context, cfg config.DBConfig) (*pgxpool.Pool, error) {
	dsn := BuildDSN(cfg)

	// Парсинг строки подключения в конфигурацию пула
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинге конфигурации пула: %w", err)
	}

	// Настройка параметров пула соединений
	poolConfig.MaxConns = 10                      // Максимальное количество соединений
	poolConfig.MinConns = 5                       // Минимальное количество соединений
	poolConfig.MaxConnLifetime = 1 * time.Hour    // Максимальное время жизни соединения
	poolConfig.MaxConnIdleTime = 30 * time.Minute // Максимальное время простоя соединения

	// Создание пула соединений с таймаутом
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании пула соединений: %w", err)
	}

	// Проверка соединения с базой данных
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ошибка при проверке соединения с базой данных: %w", err)
	}

	return pool, nil
}
