package dto

// CreateCardRequest представляет запрос на создание новой карты
type CreateCardRequest struct {
	PGPKey string `json:"pgp_key"` // Публичный ключ PGP для шифрования данных карты
}

// CreateCardResponse содержит данные созданной карты
type CreateCardResponse struct {
	ID         int64  `json:"id"`          // ID карты
	UserID     int64  `json:"user_id"`     // ID владельца карты
	CreatedAt  string `json:"created_at"`  // Дата и время создания карты
	CardNumber string `json:"card_number"` // Маскированный номер карты
	Expire     string `json:"expire"`      // Дата истечения срока действия карты
	CVV        string `json:"cvv"`         // CVV-код (обычно скрыт или маскирован)
}

// CardResponse содержит базовые данные карты без секретных данных
type CardResponse struct {
	ID        int64  `json:"id"`         // ID карты
	UserID    int64  `json:"user_id"`    // ID владельца
	CreatedAt string `json:"created_at"` // Дата и время создания
}

// CardDetailsResponse содержит подробную информацию о карте
type CardDetailsResponse struct {
	ID         int64  `json:"id"`          // ID карты
	CardNumber string `json:"card_number"` // Маскированный номер карты
	Expire     string `json:"expire"`      // Дата истечения срока действия
}

// CardListResponse представляет список карт пользователя
type CardListResponse struct {
	Cards []CardResponse `json:"cards"` // Массив карт
}

// CardPaymentRequest представляет запрос на оплату с карты
type CardPaymentRequest struct {
	CardID int64  `json:"card_id"` // ID карты для оплаты
	Amount string `json:"amount"`  // Сумма платежа
	CVV    string `json:"cvv"`     // CVV-код карты
	PGPKey string `json:"pgp_key"` // Публичный ключ PGP для шифрования данных
}

// CardPaymentResponse содержит результат операции оплаты
type CardPaymentResponse struct {
	Success     bool   `json:"success"`               // Успешность операции
	PaymentID   string `json:"payment_id,omitempty"`  // Идентификатор платежа (если успешно)
	Description string `json:"description,omitempty"` // Описание результата или ошибок
}
