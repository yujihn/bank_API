package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yujihn/bank_API/internal/service"
)

// Ключ для хранения ID пользователя в контексте
type contextKey string

const UserIDKey contextKey = "userID"

// JWTMiddleware обеспечивает проверку JWT-токена и добавление ID пользователя в контекст
type JWTMiddleware struct {
	authService service.AuthService // Сервис для работы с JWT
	logger      *logrus.Logger      // Логгер для логирования
}

// NewJWTMiddleware создает новый middleware для проверки JWT
func NewJWTMiddleware(authService service.AuthService, logger *logrus.Logger) *JWTMiddleware {
	return &JWTMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// Middleware проверяет наличие и валидность JWT-токена, добавляя ID пользователя в контекст
func (m *JWTMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получение заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Требуется авторизация", http.StatusUnauthorized)
			return
		}

		// Проверка формата: Bearer <токен>
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			http.Error(w, "Неверный формат токена", http.StatusUnauthorized)
			return
		}

		// Извлечение токена из заголовка
		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)

		// Проверка и разбор токена, получение ID пользователя
		userID, err := m.authService.ParseToken(tokenString)
		if err != nil {
			m.logger.WithError(err).Warn("Ошибка проверки токена")
			http.Error(w, "Неверный или просроченный токен", http.StatusUnauthorized)
			return
		}

		// Добавление ID пользователя в контекст запроса
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID извлекает ID пользователя из контекста
// возвращает ошибку, если значение не найдено или имеет неправильный тип
func GetUserID(ctx context.Context) (int64, error) {
	val := ctx.Value(UserIDKey)
	userID, ok := val.(int64)
	if !ok {
		return 0, errors.New("ID пользователя не найден или имеет неверный тип в контексте")
	}
	return userID, nil
}
