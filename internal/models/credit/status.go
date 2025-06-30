package credit

// Status представляет статус кредита
type Status string

const (
	ACTIVE  Status = "ACTIVE"  // Кредит активен
	CLOSED  Status = "CLOSED"  // Кредит закрыт
	OVERDUE Status = "OVERDUE" // Кредит просрочен
)
