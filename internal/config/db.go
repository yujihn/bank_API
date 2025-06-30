package config

import (
	"os"
)

// DBConfig содержит параметры для подключения к базе данных
type DBConfig struct {
	Host     string // Адрес сервера базы данных
	Port     string // Порт для подключения к базе данных
	User     string // Имя пользователя базы данных
	Password string // Пароль пользователя базы данных
	DBName   string // Имя базы данных
	SSLMode  string // Режим SSL для подключения
}

// LoadDB загружает параметры подключения к базе данных из переменных окружения
func LoadDB() DBConfig {
	return DBConfig{
		Host:     getEnv("DB_HOST", "localhost"),  // Значение по умолчанию: localhost
		Port:     getEnv("DB_PORT", "5432"),       // Значение по умолчанию: 5432
		User:     getEnv("DB_USER", "user"),       // Значение по умолчанию: user
		Password: getEnv("DB_PASSWORD", "pass"),   // Значение по умолчанию: pass
		DBName:   getEnv("DB_NAME", "mydb"),       // Значение по умолчанию: mydb
		SSLMode:  getEnv("DB_SSLMODE", "disable"), // Значение по умолчанию: disable
	}
}

// getEnv получает значение переменной окружения или возвращает заданное значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
