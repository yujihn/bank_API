package models

import "time"

// User представляет модель пользователя
type User struct {
	ID        int64     `db:"id" json:"id"`                 // Уникальный идентификатор пользователя
	Email     string    `db:"email" json:"email"`           // Электронная почта пользователя
	Password  string    `db:"password_hash" json:"-"`       // Хэш пароля (не выводится в JSON)
	CreatedAt time.Time `db:"created_at" json:"created_at"` // Дата и время регистрации пользователя
}
