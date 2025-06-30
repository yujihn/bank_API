package models

import "time"

// Card представляет модель банковской карты
type Card struct {
	ID         int64     `db:"id"        json:"id"`          // Уникальный идентификатор карты
	UserID     int64     `db:"user_id"   json:"user_id"`     // Идентификатор владельца карты
	CardNumber []byte    `db:"card_number" json:"-"`         // Шифрованный номер карты (не выводится в JSON)
	Expire     []byte    `db:"expire"      json:"-"`         // Срок действия карты (шифрованный, не выводится в JSON)
	CVVHash    string    `db:"cvv_hash"    json:"-"`         // Хэш CVV-кода (не выводится в JSON)
	CreatedAt  time.Time `db:"created_at" json:"created_at"` // Дата и время создания записи о карте
}
