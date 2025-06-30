package config

import (
	"os"
	"time"
)

// JWTConfig содержит настройки для JWT-токенов, используемых для аутентификации
type JWTConfig struct {
	Secret    string        // Секретный ключ для подписи JWT
	ExpiresIn time.Duration // Время жизни токена
}

// LoadJWT загружает конфигурацию JWT из переменных окружения или использует значения по умолчанию
func LoadJWT() JWTConfig {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// Значение по умолчанию, если переменная окружения не установлена
		secret = "default-bank-api-jwt-secret-key"
	}

	return JWTConfig{
		Secret:    secret,         // Секретный ключ
		ExpiresIn: 24 * time.Hour, // Время жизни токена: 24 часа
	}
}
