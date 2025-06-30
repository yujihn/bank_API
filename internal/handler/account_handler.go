package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/yujihn/bank_API/internal/dto"
	"github.com/yujihn/bank_API/internal/middleware"
	"github.com/yujihn/bank_API/internal/models/account"
	"github.com/yujihn/bank_API/internal/service"
)

type AccountHandler struct {
	accountService *service.AccountService
	logger         *logrus.Logger
}

func NewAccountHandler(accountService *service.AccountService, logger *logrus.Logger) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
		logger:         logger,
	}
}

// CreateAccount обрабатывает запрос на создание нового счета
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста (установлен middleware)
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.logger.Errorf("Ошибка получения userID из контекста: %v", err)
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	// Декодируем запрос
	var req dto.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Ошибка декодирования запроса: %v", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Проверяем валютный код (пока только RUB)
	if req.Currency != account.RUB {
		h.logger.Warnf("Попытка создать счет в неподдерживаемой валюте: %s", req.Currency)
		http.Error(w, "Поддерживается только валюта RUB", http.StatusBadRequest)
		return
	}

	// Создаем счет
	newAccount, err := h.accountService.CreateAccount(r.Context(), userID, req.Currency)
	if err != nil {
		h.logger.Errorf("Ошибка создания счета: %v", err)
		http.Error(w, "Не удалось создать счет", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	resp := dto.AccountResponse{
		ID:        newAccount.ID,
		UserID:    newAccount.UserID,
		Balance:   newAccount.Balance,
		Currency:  newAccount.Currency,
		CreatedAt: newAccount.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("Ошибка кодирования ответа: %v", err)
	}
}

// GetAccounts обработчик для получения списка счетов пользователя
func (h *AccountHandler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.logger.Errorf("Ошибка получения userID из контекста: %v", err)
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	// Получаем счета
	accounts, err := h.accountService.GetAccountsByUserID(r.Context(), userID)
	if err != nil {
		h.logger.Errorf("Ошибка получения счетов: %v", err)
		http.Error(w, "Не удалось получить счета", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	resp := dto.AccountsListResponse{
		Accounts: make([]dto.AccountResponse, 0, len(accounts)),
	}

	for _, acc := range accounts {
		resp.Accounts = append(resp.Accounts, dto.AccountResponse{
			ID:        acc.ID,
			UserID:    acc.UserID,
			Balance:   acc.Balance,
			Currency:  acc.Currency,
			CreatedAt: acc.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("Ошибка кодирования ответа: %v", err)
	}
}

// UpdateBalance обработчик для пополнения/списания средств
func (h *AccountHandler) UpdateBalance(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.logger.Errorf("Ошибка получения userID из контекста: %v", err)
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	// Получаем ID счета из URL
	vars := mux.Vars(r)
	accountID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Warnf("Неверный формат ID счета: %v", err)
		http.Error(w, "Неверный ID счета", http.StatusBadRequest)
		return
	}

	// Декодируем запрос
	var req dto.UpdateBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Ошибка декодирования запроса: %v", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Обновляем баланс
	err = h.accountService.UpdateBalance(r.Context(), accountID, userID, req.Amount)
	if err != nil {
		// Определяем тип ошибки для возврата подходящего HTTP-статуса
		switch {
		case errors.Is(err, service.ErrInsufficientFunds):
			h.logger.Warnf("Недостаточно средств для операции: %v", err)
			http.Error(w, "Недостаточно средств", http.StatusBadRequest)
		default:
			h.logger.Errorf("Ошибка обновления баланса: %v", err)
			http.Error(w, "Не удалось обновить баланс", http.StatusInternalServerError)
		}
		return
	}

	// Получаем обновленный счет для ответа
	updatedAccount, err := h.accountService.GetAccountByID(r.Context(), accountID, userID)
	if err != nil {
		h.logger.Errorf("Ошибка получения обновленного счета: %v", err)
		http.Error(w, "Операция выполнена, но не удалось получить данные счета", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	resp := dto.AccountResponse{
		ID:        updatedAccount.ID,
		UserID:    updatedAccount.UserID,
		Balance:   updatedAccount.Balance,
		Currency:  updatedAccount.Currency,
		CreatedAt: updatedAccount.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("Ошибка кодирования ответа: %v", err)
	}
}

// Transfer обработчик для перевода между счетами
func (h *AccountHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.logger.Errorf("Ошибка получения userID из контекста: %v", err)
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	// Декодируем запрос
	var req dto.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Errorf("Ошибка декодирования запроса: %v", err)
		http.Error(w, "Неверный формат запроса", http.StatusBadRequest)
		return
	}

	// Выполняем перевод
	err = h.accountService.Transfer(r.Context(), req.FromAccountID, req.ToAccountID, userID, req.Amount)
	if err != nil {
		// Определяем тип ошибки
		switch {
		case errors.Is(err, service.ErrInsufficientFunds):
			h.logger.Warnf("Недостаточно средств для перевода: %v", err)
			http.Error(w, "Недостаточно средств", http.StatusBadRequest)
		case errors.Is(err, service.ErrSameAccount):
			h.logger.Warnf("Попытка перевода на тот же счет: %v", err)
			http.Error(w, "Нельзя переводить на тот же счет", http.StatusBadRequest)
		case errors.Is(err, service.ErrNegativeAmount):
			h.logger.Warnf("Попытка перевода отрицательной суммы: %v", err)
			http.Error(w, "Сумма перевода должна быть положительной", http.StatusBadRequest)
		default:
			h.logger.Errorf("Ошибка выполнения перевода: %v", err)
			http.Error(w, "Не удалось выполнить перевод", http.StatusInternalServerError)
		}
		return
	}

	// Отправляем успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "success"}); err != nil {
		h.logger.Errorf("Ошибка кодирования ответа: %v", err)
	}
}

// GetTransactions обработчик для получения списка транзакций по счету
func (h *AccountHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	// Получаем userID из контекста
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		h.logger.Errorf("Ошибка получения userID из контекста: %v", err)
		http.Error(w, "Ошибка авторизации", http.StatusUnauthorized)
		return
	}

	// Получаем ID счета из URL
	vars := mux.Vars(r)
	accountID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		h.logger.Warnf("Неверный формат ID счета: %v", err)
		http.Error(w, "Неверный ID счета", http.StatusBadRequest)
		return
	}

	// Получаем транзакции
	transactions, err := h.accountService.GetTransactionsByAccountID(r.Context(), accountID, userID)
	if err != nil {
		h.logger.Errorf("Ошибка получения транзакций: %v", err)
		http.Error(w, "Не удалось получить транзакции", http.StatusInternalServerError)
		return
	}

	// Формируем ответ
	resp := dto.TransactionListResponse{
		Transactions: make([]dto.TransactionResponse, 0, len(transactions)),
	}

	for _, tx := range transactions {
		resp.Transactions = append(resp.Transactions, dto.TransactionResponse{
			ID:        tx.ID,
			AccountID: tx.AccountID,
			Amount:    tx.Amount,
			Type:      tx.Type,
			Status:    tx.Status,
			CreatedAt: tx.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		h.logger.Errorf("Ошибка кодирования ответа: %v", err)
	}
}
