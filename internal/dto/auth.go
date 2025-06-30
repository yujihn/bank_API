package dto

// RegisterRequest представляет запрос на регистрацию нового пользователя
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`    // Электронная почта (обязательное поле, формат email)
	Password string `json:"password" binding:"required,min=6"` // Пароль (обязательное поле, минимум 6 символов)
}

// LoginRequest представляет запрос на аутентификацию пользователя
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"` // Электронная почта (обязательное поле, формат email)
	Password string `json:"password" binding:"required"`    // Пароль (обязательное поле)
}

// AuthResponse представляет ответ с JWT-токеном после успешной аутентификации
type AuthResponse struct {
	Token string `json:"token"` // JWT-токен для последующих запросов
}
