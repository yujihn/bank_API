package credit

import (
	"github.com/shopspring/decimal"
	"time"
)

// Credit представляет модель кредита
type Credit struct {
	ID           int64           `db:"id"            json:"id"`            // Уникальный идентификатор кредита
	AccountID    int64           `db:"account_id"    json:"account_id"`    // Идентификатор связанного счета
	Principal    decimal.Decimal `db:"principal"     json:"principal"`     // Основная сумма кредита
	InterestRate float64         `db:"interest_rate" json:"interest_rate"` // Процентная ставка по кредиту
	TermMonths   int             `db:"term_months"   json:"term_months"`   // Срок кредита в месяцах
	StartDate    time.Time       `db:"start_date"    json:"start_date"`    // Дата начала кредита
	Status       Status          `db:"status"        json:"status"`        // Статус кредита (например, активен, закрыт)
	CreatedAt    time.Time       `db:"created_at"    json:"created_at"`    // Дата и время создания записи о кредите
}
