package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/yujihn/bank_API/internal/config"
	"github.com/yujihn/bank_API/internal/dto"
	"github.com/yujihn/bank_API/internal/models"
	"github.com/yujihn/bank_API/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// Различные ошибки, которые могут возникнуть в процессе аутентификации
var (
	ErrInvalidCredentials = errors.New("неверные учетные данные")     // Ошибка при неправильных данных входа
	ErrUserExists         = errors.New("пользователь уже существует") // Ошибка при попытке зарегистрировать существующего пользователя
)

// AuthService интерфейс для сервиса аутентификации
type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (int64, error) // Регистрация нового пользователя
	Login(ctx context.Context, req dto.LoginRequest) (string, error)      // Вход и получение JWT
	ParseToken(tokenString string) (int64, error)                         // Разбор JWT и получение ID пользователя
}

// authService реализует интерфейс AuthService
type authService struct {
	userRepo repository.UserRepository // Репозиторий пользователей
	jwtCfg   config.JWTConfig          // Конфигурация JWT
}

// NewAuthService создает новый сервис аутентификации
func NewAuthService(userRepo repository.UserRepository, jwtCfg config.JWTConfig) AuthService {
	return &authService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

// Register регистрирует нового пользователя
func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (int64, error) {
	// Хеширование пароля с использованием bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	user := &models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	id, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Login аутентифицирует пользователя и возвращает JWT-токен
func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	// Проверка пароля с помощью bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// Генерация JWT-токена
	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// generateToken создает JWT с данными пользователя
func (s *authService) generateToken(userID int64) (string, error) {
	// Создаем claims для JWT
	claims := jwt.MapClaims{
		"sub": userID,                                    // subject (ID пользователя)
		"exp": time.Now().Add(s.jwtCfg.ExpiresIn).Unix(), // время истечения
		"iat": time.Now().Unix(),                         // время выпуска
	}

	// Создаем новый токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен секретным ключом
	tokenString, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken разбирает JWT и возвращает ID пользователя
func (s *authService) ParseToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неожиданный метод подписи токена")
		}
		return []byte(s.jwtCfg.Secret), nil
	})

	if err != nil {
		return 0, err
	}

	// Проверка валидности токена
	if !token.Valid {
		return 0, errors.New("невалидный токен")
	}

	// Получение claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("невалидные claims")
	}

	// Извлечение ID пользователя из claims
	userID, ok := claims["sub"].(float64)
	if !ok {
		return 0, errors.New("невалидный ID пользователя")
	}

	return int64(userID), nil
}
