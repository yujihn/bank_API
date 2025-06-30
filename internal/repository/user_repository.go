package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yujihn/bank_API/internal/models"
)

// ErrUserNotFound возвращается, когда пользователь не найден в базе данных
var ErrUserNotFound = errors.New("пользователь не найден")

// UserRepository интерфейс для работы с данными пользователей
type UserRepository interface {
	Create(ctx context.Context, user *models.User) (int64, error)       // Создает нового пользователя
	GetByEmail(ctx context.Context, email string) (*models.User, error) // Находит пользователя по email
	GetByID(ctx context.Context, id int64) (*models.User, error)        // Находит пользователя по ID
}

// UserRepositoryPgx реализует интерфейс UserRepository с помощью pgx
type UserRepositoryPgx struct {
	pool *pgxpool.Pool // Пул соединений с базой данных
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &UserRepositoryPgx{pool: pool}
}

// Create создает нового пользователя в базе данных
func (r *UserRepositoryPgx) Create(ctx context.Context, user *models.User) (int64, error) {
	var id int64

	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) 
         VALUES ($1, $2) 
         RETURNING id`,
		user.Email, user.Password).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// GetByEmail ищет пользователя по email
func (r *UserRepositoryPgx) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}

	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at 
         FROM users 
         WHERE email = $1`,
		email).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetByID ищет пользователя по ID
func (r *UserRepositoryPgx) GetByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}

	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at 
         FROM users 
         WHERE id = $1`,
		id).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}
