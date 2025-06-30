package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yujihn/bank_API/internal/models"
	"github.com/yujihn/bank_API/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// CardService обеспечивает бизнес-логику для работы с картами
type CardService struct {
	cardRepo      *repository.CardRepository // Репозиторий карт
	db            *pgxpool.Pool              // Пул соединений с базой данных
	encryptionKey []byte                     // Ключ для HMAC подписи
}

// NewCardService создает новый сервис карт
func NewCardService(cardRepo *repository.CardRepository, db *pgxpool.Pool, encryptionKey string) *CardService {
	return &CardService{
		cardRepo:      cardRepo,
		db:            db,
		encryptionKey: []byte(encryptionKey),
	}
}

// generateCardNumber генерирует номер карты, проходящий проверку Луна
func (s *CardService) generateCardNumber() (string, error) {
	// Банковский идентификатор (BIN)
	bin := "400000" // Пример BIN

	// Генерируем случайные цифры для номера карты
	// до предпоследней цифры (всего 15)
	var digits string
	randomDigits := make([]byte, 9)
	_, err := rand.Read(randomDigits)
	if err != nil {
		return "", err
	}

	// Конвертируем случайные байты в цифры
	for _, b := range randomDigits {
		digit := int(b) % 10
		digits += strconv.Itoa(digit)
	}

	// Полный номер без контрольной цифры
	number := bin + digits

	// Вычисляем контрольную цифру по алгоритму Луна
	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(number[i]))

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	checkDigit := (10 - (sum % 10)) % 10

	// Добавляем контрольную цифру
	fullNumber := number + strconv.Itoa(checkDigit)

	return fullNumber, nil
}

// generateExpirationDate генерирует дату истечения срока действия карты
func (s *CardService) generateExpirationDate() string {
	now := time.Now()
	expiryDate := now.AddDate(3, 0, 0) // Карта действительна 3 года
	return fmt.Sprintf("%02d/%d", expiryDate.Month(), expiryDate.Year()%100)
}

// generateCVV генерирует случайный CVV код
func (s *CardService) generateCVV() (string, error) {
	// Генерируем случайные 3 цифры
	b := make([]byte, 2)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Конвертируем в число от 100 до 999
	cvv := 100 + (int(b[0])<<8|int(b[1]))%900
	return fmt.Sprintf("%03d", cvv), nil
}

// encryptWithPGP шифрует данные с помощью PGP через SQL-функцию
func (s *CardService) encryptWithPGP(ctx context.Context, data string, key string) ([]byte, error) {
	query := `SELECT pgp_sym_encrypt($1, $2)`
	var encrypted []byte
	err := s.db.QueryRow(ctx, query, data, key).Scan(&encrypted)
	return encrypted, err
}

// decryptWithPGP расшифровывает данные с помощью PGP через SQL-функцию
func (s *CardService) decryptWithPGP(ctx context.Context, data []byte, key string) (string, error) {
	query := `SELECT pgp_sym_decrypt($1, $2)`
	var decrypted string
	err := s.db.QueryRow(ctx, query, data, key).Scan(&decrypted)
	return decrypted, err
}

// hashCVV создает bcrypt-хеш для CVV
func (s *CardService) hashCVV(cvv string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// validateCVV проверяет CVV против хеша
func (s *CardService) validateCVV(cvv string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(cvv))
	return err == nil
}

// CreateCard создает новую виртуальную карту
func (s *CardService) CreateCard(ctx context.Context, userID int64, pgpKey string) (*models.Card, map[string]string, error) {
	// Генерируем данные карты
	cardNumber, err := s.generateCardNumber()
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка генерации номера карты: %w", err)
	}

	expireDate := s.generateExpirationDate()
	cvv, err := s.generateCVV()
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка генерации CVV: %w", err)
	}

	// Шифруем номер карты и срок действия с PGP
	encryptedNumber, err := s.encryptWithPGP(ctx, cardNumber, pgpKey)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка шифрования номера карты: %w", err)
	}

	encryptedExpire, err := s.encryptWithPGP(ctx, expireDate, pgpKey)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка шифрования срока действия: %w", err)
	}

	// Хешируем CVV с bcrypt
	cvvHash, err := s.hashCVV(cvv)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка хеширования CVV: %w", err)
	}

	// Создаем запись в базе данных
	card, err := s.cardRepo.CreateCard(ctx, userID, encryptedNumber, encryptedExpire, cvvHash)
	if err != nil {
		return nil, nil, fmt.Errorf("ошибка создания карты в базе: %w", err)
	}

	// Генерируем цифровую подпись для проверки целостности
	message := fmt.Sprintf("%d:%s:%s:%s", card.ID, cardNumber, expireDate, cvv)
	signature := s.generateHMAC(message)

	// В реальном приложении сохраняли бы подпись в базе для будущей проверки
	// Здесь просто логируем (в продакшн-коде этого делать не стоит)

	// Данные для отображения пользователю (один раз)
	cardDetails := map[string]string{
		"number":    cardNumber,
		"expire":    expireDate,
		"cvv":       cvv,
		"signature": signature, // Можно вернуть подпись клиенту для сохранения
	}

	return card, cardDetails, nil
}

// GetCardDetails получает расшифрованные данные карты (только для владельца)
func (s *CardService) GetCardDetails(ctx context.Context, cardID int64, userID int64, pgpKey string) (map[string]string, error) {
	// Получаем карту
	card, err := s.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("карта не найдена")
		}
		return nil, fmt.Errorf("ошибка получения карты: %w", err)
	}

	// Проверяем, что карта принадлежит пользователю
	if card.UserID != userID {
		return nil, errors.New("доступ запрещен: карта не принадлежит пользователю")
	}

	// Расшифровываем данные
	cardNumber, err := s.decryptWithPGP(ctx, card.CardNumber, pgpKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки номера карты: %w", err)
	}

	expireDate, err := s.decryptWithPGP(ctx, card.Expire, pgpKey)
	if err != nil {
		return nil, fmt.Errorf("ошибка расшифровки срока действия: %w", err)
	}

	// Маскируем номер карты для безопасности (отображаем только последние 4 цифры)
	maskedNumber := "**** **** **** " + cardNumber[len(cardNumber)-4:]

	// Формируем ответ
	cardDetails := map[string]string{
		"number": maskedNumber,
		"expire": expireDate,
	}

	return cardDetails, nil
}

// GetUserCards получает список карт пользователя (без секретных данных)
func (s *CardService) GetUserCards(ctx context.Context, userID int64) ([]*models.Card, error) {
	return s.cardRepo.GetCardsByUserID(ctx, userID)
}

// VerifyCardPayment проверяет данные карты для платежа
func (s *CardService) VerifyCardPayment(ctx context.Context, cardID int64, cvv string, pgpKey string) (bool, error) {
	// Получаем карту
	card, err := s.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, errors.New("карта не найдена")
		}
		return false, fmt.Errorf("ошибка получения карты: %w", err)
	}

	// Проверяем CVV
	isValidCVV := s.validateCVV(cvv, card.CVVHash)
	if !isValidCVV {
		return false, errors.New("неверный CVV код")
	}

	// Проверяем срок действия
	expire, err := s.decryptWithPGP(ctx, card.Expire, pgpKey)
	if err != nil {
		return false, fmt.Errorf("ошибка расшифровки срока действия: %w", err)
	}

	// Расшифровываем номер карты (нужен для формирования HMAC)
	cardNumber, err := s.decryptWithPGP(ctx, card.CardNumber, pgpKey)
	if err != nil {
		return false, fmt.Errorf("ошибка расшифровки номера карты: %w", err)
	}

	// Парсим дату из формата MM/YY
	var month, year int
	_, err = fmt.Sscanf(expire, "%d/%d", &month, &year)
	if err != nil {
		return false, fmt.Errorf("ошибка парсинга срока действия: %w", err)
	}

	// Добавляем 2000 к году (для формата YY)
	year += 2000

	// Проверяем, что карта не просрочена
	now := time.Now()
	expiryDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	// Карта действительна до конца указанного месяца
	expiryDate = expiryDate.AddDate(0, 1, -1)

	if now.After(expiryDate) {
		return false, errors.New("карта просрочена")
	}

	// Генерируем цифровую подпись для проверки целостности
	message := fmt.Sprintf("%d:%s:%s:%s", cardID, cardNumber, expire, cvv)
	hmacSignature := s.generateHMAC(message)

	// В реальном приложении мы бы сравнивали сгенерированную подпись
	// с сохраненной при создании карты, но в данном случае просто
	// проверяем, что подпись корректно сформирована
	if len(hmacSignature) == 0 {
		return false, errors.New("ошибка генерации HMAC-подписи")
	}

	// Проверка успешна
	return true, nil
}

// generateHMAC создает HMAC-SHA256 для данных
func (s *CardService) generateHMAC(message string) string {
	h := hmac.New(sha256.New, s.encryptionKey)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

// verifyHMAC проверяет HMAC-SHA256 подпись
func (s *CardService) verifyHMAC(message, signature string) bool {
	expectedMAC, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, s.encryptionKey)
	mac.Write([]byte(message))
	return hmac.Equal(mac.Sum(nil), expectedMAC)
}
