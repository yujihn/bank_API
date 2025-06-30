package dto

import (
	"github.com/shopspring/decimal"
	"github.com/yujihn/bank_API/internal/models/account"
	"github.com/yujihn/bank_API/internal/models/transaction"
)

// CreateAccountRequest представляет запрос на создание нового счета
type CreateAccountRequest struct {
	Currency account.Currency `json:"currency"` // Валюта счета
}

// UpdateBalanceRequest представляет запрос на пополнение или списание средств со счета
type UpdateBalanceRequest struct {
	Amount decimal.Decimal `json:"amount"` // Сумма для пополнения или списания
}

// TransferRequest представляет запрос на перевод средств между счетами
type TransferRequest struct {
	FromAccountID int64           `json:"from_account_id"` // ID счета отправителя
	ToAccountID   int64           `json:"to_account_id"`   // ID счета получателя
	Amount        decimal.Decimal `json:"amount"`          // Сумма перевода
}

// AccountResponse представляет ответ с информацией о счете
type AccountResponse struct {
	ID        int64            `json:"id"`         // ID счета
	UserID    int64            `json:"user_id"`    // ID пользователя, владельца счета
	Balance   decimal.Decimal  `json:"balance"`    // Текущий баланс
	Currency  account.Currency `json:"currency"`   // Валюта счета
	CreatedAt string           `json:"created_at"` // Дата и время создания счета
}

// TransactionResponse представляет ответ с информацией о транзакции
type TransactionResponse struct {
	ID        int64              `json:"id"`         // ID транзакции
	AccountID int64              `json:"account_id"` // ID связанного счета
	Amount    decimal.Decimal    `json:"amount"`     // Сумма транзакции
	Type      transaction.Type   `json:"type"`       // Тип транзакции
	Status    transaction.Status `json:"status"`     // Статус транзакции
	CreatedAt string             `json:"created_at"` // Дата и время создания транзакции
}

// AccountsListResponse представляет список счетов
type AccountsListResponse struct {
	Accounts []AccountResponse `json:"accounts"` // Массив счетов
}

// TransactionListResponse представляет список транзакций
type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"` // Массив транзакций
}
