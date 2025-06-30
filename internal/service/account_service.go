package service

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"
	"github.com/yujihn/bank_API/internal/models/account"
	"github.com/yujihn/bank_API/internal/models/transaction"
	"github.com/yujihn/bank_API/internal/repository"
)

var (
	ErrInsufficientFunds = errors.New("недостаточно средств")                    // Ошибка при недостатке средств на счете
	ErrSameAccount       = errors.New("нельзя переводить деньги на тот же счет") // Ошибка при попытке перевода на тот же счет
	ErrNegativeAmount    = errors.New("сумма не может быть отрицательной")       // Ошибка при отрицательной сумме
)

type AccountService struct {
	accountRepo     *repository.AccountRepository     // Репозиторий для работы со счетами
	transactionRepo *repository.TransactionRepository // Репозиторий для работы с транзакциями
}

// NewAccountService создает новый сервис для работы со счетами
func NewAccountService(accountRepo *repository.AccountRepository, transactionRepo *repository.TransactionRepository) *AccountService {
	return &AccountService{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

// CreateAccount создает новый счет для пользователя
func (s *AccountService) CreateAccount(ctx context.Context, userID int64, currency account.Currency) (*account.Account, error) {
	return s.accountRepo.CreateAccount(ctx, userID, currency)
}

// GetAccountByID получает счет по ID с проверкой владения
func (s *AccountService) GetAccountByID(ctx context.Context, id int64, userID int64) (*account.Account, error) {
	acc, err := s.accountRepo.GetAccountByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Проверка, принадлежит ли счет пользователю
	if acc.UserID != userID {
		return nil, errors.New("счет не принадлежит пользователю")
	}

	return acc, nil
}

// GetAccountsByUserID получает все счета пользователя
func (s *AccountService) GetAccountsByUserID(ctx context.Context, userID int64) ([]*account.Account, error) {
	return s.accountRepo.GetAccountsByUserID(ctx, userID)
}

// UpdateBalance пополняет или снимает средства со счета
func (s *AccountService) UpdateBalance(ctx context.Context, id int64, userID int64, amount decimal.Decimal) error {
	// Проверка суммы: нулевая сумма недопустима
	if amount.Equal(decimal.Zero) {
		return errors.New("сумма должна быть отлична от нуля")
	}

	// Получение счета с проверкой владения
	acc, err := s.GetAccountByID(ctx, id, userID)
	if err != nil {
		return err
	}

	// Если это списание, проверяем достаточность средств
	if amount.LessThan(decimal.Zero) && acc.Balance.Add(amount).LessThan(decimal.Zero) {
		return ErrInsufficientFunds
	}

	// Определяем тип транзакции: пополнение или списание
	txType := transaction.WITHDRAWAL // Списание
	if amount.GreaterThan(decimal.Zero) {
		txType = transaction.DEPOSIT // Пополнение
	}

	// Обновляем баланс в базе данных
	err = s.accountRepo.UpdateBalance(ctx, id, amount)
	if err != nil {
		return err
	}

	// Записываем транзакцию
	absAmount := amount.Abs()
	_, err = s.transactionRepo.CreateTransaction(ctx, id, absAmount, txType, transaction.COMPLETED)

	return err
}

// Transfer переводит деньги между счетами
func (s *AccountService) Transfer(ctx context.Context, fromID, toID int64, userID int64, amount decimal.Decimal) error {
	// Проверки
	if fromID == toID {
		return ErrSameAccount
	}

	if amount.LessThanOrEqual(decimal.Zero) {
		return ErrNegativeAmount
	}

	// Проверка владения счетом отправителя
	fromAcc, err := s.GetAccountByID(ctx, fromID, userID)
	if err != nil {
		return err
	}

	// Проверка достаточности средств
	if fromAcc.Balance.LessThan(amount) {
		return ErrInsufficientFunds
	}

	// Проверка существования счета получателя
	_, err = s.accountRepo.GetAccountByID(ctx, toID)
	if err != nil {
		return err
	}

	// Выполнение перевода в базе данных
	err = s.accountRepo.TransferBetweenAccounts(ctx, fromID, toID, amount)
	if err != nil {
		return err
	}

	// Запись транзакций для обоих счетов
	_, err = s.transactionRepo.CreateTransaction(ctx, fromID, amount, transaction.DEPOSIT, transaction.COMPLETED)
	if err != nil {
		return err
	}

	_, err = s.transactionRepo.CreateTransaction(ctx, toID, amount, transaction.WITHDRAWAL, transaction.COMPLETED)
	return err
}

// GetTransactionsByAccountID получает историю транзакций для конкретного счета
func (s *AccountService) GetTransactionsByAccountID(ctx context.Context, accountID int64, userID int64) ([]*transaction.Transaction, error) {
	// Проверка владения счетом
	_, err := s.GetAccountByID(ctx, accountID, userID)
	if err != nil {
		return nil, err
	}

	return s.transactionRepo.GetTransactionsByAccountID(ctx, accountID)
}

// GetTransactionsByUserID получает все транзакции пользователя по его ID
func (s *AccountService) GetTransactionsByUserID(ctx context.Context, userID int64) ([]*transaction.Transaction, error) {
	return s.transactionRepo.GetTransactionsByUserID(ctx, userID)
}
