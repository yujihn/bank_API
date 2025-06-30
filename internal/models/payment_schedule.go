package models

import (
	"github.com/shopspring/decimal"
	"time"
)

// PaymentSchedule представляет график платежей по кредиту
type PaymentSchedule struct {
	ID        int64           `db:"id"         json:"id"`         // Уникальный идентификатор платежа
	CreditID  int64           `db:"credit_id"  json:"credit_id"`  // Идентификатор связанного кредита
	DueDate   time.Time       `db:"due_date"   json:"due_date"`   // Дата погашения платежа
	Amount    decimal.Decimal `db:"amount"     json:"amount"`     // Сумма платежа
	Paid      bool            `db:"paid"       json:"paid"`       // Статус оплаты (оплачен/не оплачен)
	CreatedAt time.Time       `db:"created_at" json:"created_at"` // Дата и время создания записи о платеже
}
