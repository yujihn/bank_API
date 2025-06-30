package transaction

// Type представляет тип транзакции
type Type string

const (
	DEPOSIT    Type = "DEPOSIT"    // Пополнение счета
	WITHDRAWAL Type = "WITHDRAWAL" // Снятие средств
	TRANSFER   Type = "TRANSFER"   // Перевод между счетами
)
