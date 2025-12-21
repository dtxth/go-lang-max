package repository

import (
	"database/sql"
	"time"

	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/database"
)

type PasswordResetPostgres struct {
	db *database.DB
}

func NewPasswordResetPostgres(db *database.DB) *PasswordResetPostgres {
	return &PasswordResetPostgres{db: db}
}

func (r *PasswordResetPostgres) Create(token *domain.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(
		query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
	).Scan(&token.ID, &token.CreatedAt)

	return err
}

func (r *PasswordResetPostgres) GetByToken(token string) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token = $1
	`

	resetToken := &domain.PasswordResetToken{}
	err := r.db.QueryRow(query, token).Scan(
		&resetToken.ID,
		&resetToken.UserID,
		&resetToken.Token,
		&resetToken.ExpiresAt,
		&resetToken.UsedAt,
		&resetToken.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return resetToken, nil
}

func (r *PasswordResetPostgres) Invalidate(token string) error {
	query := `
		UPDATE password_reset_tokens
		SET used_at = $1
		WHERE token = $2
	`
	_, err := r.db.Exec(query, time.Now(), token)
	return err
}

func (r *PasswordResetPostgres) DeleteExpired() error {
	query := `
		DELETE FROM password_reset_tokens
		WHERE expires_at < $1
	`
	_, err := r.db.Exec(query, time.Now())
	return err
}
