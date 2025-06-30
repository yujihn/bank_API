package transaction

// Status представляет статус транзакции
type Status string

const (
	PENDING   Status = "PENDING"   // Ожидает обработки
	COMPLETED Status = "COMPLETED" // Успешно завершена
	FAILED    Status = "FAILED"    // Неудачная транзакция
)
