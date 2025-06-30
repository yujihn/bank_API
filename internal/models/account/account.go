package account

import (
	"github.com/shopspring/decimal"
	"time"
)

// Account представляет модель банковского счета
type Account struct {
	ID        int64           `db:"id"       json:"id"`           // Уникальный идентификатор счета
	UserID    int64           `db:"user_id"  json:"user_id"`      // Идентификатор владельца счета
	Balance   decimal.Decimal `db:"balance"  json:"balance"`      // Текущий баланс счета
	Currency  Currency        `db:"currency" json:"currency"`     // Валюта счета
	CreatedAt time.Time       `db:"created_at" json:"created_at"` // Дата и время создания счета
}
