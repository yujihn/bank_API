package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yujihn/bank_API/internal/models"
)

// CardRepository реализует работу с таблицей карт в базе данных
type CardRepository struct {
	db *pgxpool.Pool // Пул соединений с базой данных
}

// NewCardRepository создает новый экземпляр репозитория для работы с картами
func NewCardRepository(db *pgxpool.Pool) *CardRepository {
	return &CardRepository{db: db}
}

// CreateCard создает новую карту с зашифрованными данными
func (r *CardRepository) CreateCard(ctx context.Context, userID int64, encryptedNumber, encryptedExpire []byte, cvvHash string) (*models.Card, error) {
	query := `
		INSERT INTO cards (user_id, card_number, expire, cvv_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, created_at
	`
	var card models.Card
	err := r.db.QueryRow(ctx, query, userID, encryptedNumber, encryptedExpire, cvvHash).Scan(
		&card.ID, &card.UserID, &card.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Устанавливаем зашифрованные данные
	card.CardNumber = encryptedNumber
	card.Expire = encryptedExpire
	card.CVVHash = cvvHash

	return &card, nil
}

// GetCardByID получает карту по ID
func (r *CardRepository) GetCardByID(ctx context.Context, cardID int64) (*models.Card, error) {
	query := `
		SELECT id, user_id, card_number, expire, cvv_hash, created_at
		FROM cards 
		WHERE id = $1
	`
	var card models.Card
	err := r.db.QueryRow(ctx, query, cardID).Scan(
		&card.ID, &card.UserID, &card.CardNumber, &card.Expire, &card.CVVHash, &card.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &card, nil
}

// GetCardsByUserID получает все карты пользователя по его ID
func (r *CardRepository) GetCardsByUserID(ctx context.Context, userID int64) ([]*models.Card, error) {
	query := `
		SELECT id, user_id, created_at
		FROM cards 
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []*models.Card
	for rows.Next() {
		var card models.Card
		if err := rows.Scan(&card.ID, &card.UserID, &card.CreatedAt); err != nil {
			return nil, err
		}
		cards = append(cards, &card)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return cards, nil
}

// IsCardExistsForUser проверяет, существует ли карта с указанным ID для пользователя
func (r *CardRepository) IsCardExistsForUser(ctx context.Context, cardID int64, userID int64) (bool, error) {
	query := `
		SELECT 1 FROM cards
		WHERE id = $1 AND user_id = $2
	`
	var exists int
	err := r.db.QueryRow(ctx, query, cardID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return true, nil
}
