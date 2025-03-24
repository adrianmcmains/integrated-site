package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/yourusername/integrated-site/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO auth.users (email, password_hash, full_name, role, avatar_url)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.Role,
		user.AvatarURL,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, avatar_url, created_at, updated_at
		FROM auth.users
		WHERE id = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, avatar_url, created_at, updated_at
		FROM auth.users
		WHERE email = $1
	`

	var user models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.AvatarURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE auth.users
		SET email = $1, full_name = $2, role = $3, avatar_url = $4
		WHERE id = $5
		RETURNING updated_at
	`

	return r.db.QueryRow(ctx, query,
		user.Email,
		user.FullName,
		user.Role,
		user.AvatarURL,
		user.ID,
	).Scan(&user.UpdatedAt)
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	query := `
		UPDATE auth.users
		SET password_hash = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, passwordHash, id)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM auth.users
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, avatar_url, created_at, updated_at
		FROM auth.users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.FullName,
			&user.Role,
			&user.AvatarURL,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) Count(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM auth.users
	`

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}