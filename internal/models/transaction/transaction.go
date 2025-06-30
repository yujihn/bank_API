package transaction

import (
	"github.com/shopspring/decimal"
	"time"
)

// Transaction представляет модель банковской транзакции
type Transaction struct {
	ID        int64           `db:"id"          json:"id"`         // Уникальный идентификатор транзакции
	AccountID int64           `db:"account_id"  json:"account_id"` // Идентификатор связанного счета
	Amount    decimal.Decimal `db:"amount"      json:"amount"`     // Сумма транзакции
	Type      Type            `db:"type"        json:"type"`       // Тип транзакции (например, перевод, пополнение)
	Status    Status          `db:"status"      json:"status"`     // Статус транзакции (например, выполнена, ошибка)
	CreatedAt time.Time       `db:"created_at"  json:"created_at"` // Дата и время создания транзакции
}
