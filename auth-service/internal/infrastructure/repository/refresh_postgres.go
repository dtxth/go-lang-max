package repository

import (
	"database/sql"
	"time"
)

type RefreshPostgres struct {
	db *sql.DB
}

func NewRefreshPostgres(db *sql.DB) *RefreshPostgres {
	return &RefreshPostgres{db: db}
}

func (r *RefreshPostgres) Save(jti string, userID int64, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO refresh_tokens (jti, user_id, expires_at) VALUES ($1, $2, $3)`,
		jti, userID, expiresAt,
	)
	return err
}

func (r *RefreshPostgres) IsValid(jti string) (bool, error) {
	var revoked bool
	var expiresAt time.Time
	err := r.db.QueryRow(
		`SELECT revoked, expires_at FROM refresh_tokens WHERE jti = $1`,
		jti,
	).Scan(&revoked, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if revoked {
		return false, nil
	}
	if time.Now().After(expiresAt) {
		return false, nil
	}
	return true, nil
}

func (r *RefreshPostgres) Revoke(jti string) error {
	_, err := r.db.Exec(`UPDATE refresh_tokens SET revoked = TRUE WHERE jti = $1`, jti)
	return err
}

func (r *RefreshPostgres) RevokeAllForUser(userID int64) error {
	_, err := r.db.Exec(`UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1`, userID)
	return err
}
