package config

import (
	"github.com/sirupsen/logrus"
)

// CryptoConfig содержит криптографические ключи для шифрования и подписи данных
type CryptoConfig struct {
	PGPKey  string // Ключ для PGP-шифрования данных
	HMACKey string // Ключ для генерации HMAC-подписей
}

// LoadCrypto загружает конфигурацию криптографических ключей из переменных окружения
func LoadCrypto() CryptoConfig {
	cfg := CryptoConfig{
		// Получение PGP-ключ из переменной окружения или использование значения по умолчанию
		PGPKey: getEnv("BANK_PGP_KEY", "bankDefaultPGPKey2024"),
		// Получение HMAC-ключ из переменной окружения или использование значения по умолчанию
		HMACKey: getEnv("BANK_HMAC_KEY", "bankDefaultHMACKey2024"),
	}

	// Логирование успешной загрузки конфигурации (сам ключи не выводятся)
	logrus.Info("Конфигурация криптографических ключей успешно загружена")

	return cfg
}
