package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/yujihn/bank_API/internal/models/account"
)

// AccountRepository реализует работу с таблицей счетов в базе данных
type AccountRepository struct {
	db *pgxpool.Pool // Пул соединений с базой данных
}

// NewAccountRepository создает новый экземпляр репозитория для работы с аккаунтами
func NewAccountRepository(db *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{db: db}
}

// CreateAccount создает новый счет для пользователя
func (r *AccountRepository) CreateAccount(ctx context.Context, userID int64, currency account.Currency) (*account.Account, error) {
	query := `
		INSERT INTO accounts (user_id, currency)
		VALUES ($1, $2)
		RETURNING id, user_id, balance, currency, created_at
	`
	var acc account.Account
	err := r.db.QueryRow(ctx, query, userID, currency).Scan(
		&acc.ID, &acc.UserID, &acc.Balance, &acc.Currency, &acc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

// GetAccountByID получает счет по его ID
func (r *AccountRepository) GetAccountByID(ctx context.Context, id int64) (*account.Account, error) {
	query := `
		SELECT id, user_id, balance, currency, created_at
		FROM accounts
		WHERE id = $1
	`
	var acc account.Account
	err := r.db.QueryRow(ctx, query, id).Scan(
		&acc.ID, &acc.UserID, &acc.Balance, &acc.Currency, &acc.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

// GetAccountsByUserID получает все счета пользователя по его ID
func (r *AccountRepository) GetAccountsByUserID(ctx context.Context, userID int64) ([]*account.Account, error) {
	query := `
		SELECT id, user_id, balance, currency, created_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY id
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*account.Account
	for rows.Next() {
		var acc account.Account
		if err := rows.Scan(&acc.ID, &acc.UserID, &acc.Balance, &acc.Currency, &acc.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, &acc)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

// UpdateBalance обновляет баланс счета, прибавляя указанную сумму
func (r *AccountRepository) UpdateBalance(ctx context.Context, id int64, amount decimal.Decimal) error {
	query := `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, amount, id)
	return err
}

// TransferBetweenAccounts выполняет перевод средств между двумя счетами в рамках транзакции
func (r *AccountRepository) TransferBetweenAccounts(ctx context.Context, fromID, toID int64, amount decimal.Decimal) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Списание со счета отправителя с проверкой достаточности средств
	updateFromQuery := `
		UPDATE accounts
		SET balance = balance - $1
		WHERE id = $2 AND balance >= $1
		RETURNING balance
	`
	var newBalance decimal.Decimal
	err = tx.QueryRow(ctx, updateFromQuery, amount, fromID).Scan(&newBalance)
	if err != nil {
		return err
	}

	// Пополнение счета получателя
	updateToQuery := `
		UPDATE accounts
		SET balance = balance + $1
		WHERE id = $2
	`
	_, err = tx.Exec(ctx, updateToQuery, amount, toID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
