package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindActiveByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := r.db.QueryRowxContext(
		ctx,
		`SELECT id, uuid, status_id, name, email, password_hash, role, last_login_at, created_at, updated_at, deleted_at
		FROM tbl_users
		WHERE LOWER(email) = LOWER($1) AND status_id = 1 AND deleted_at IS NULL`,
		strings.TrimSpace(email),
	).StructScan(&user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *Repository) TouchLastLogin(ctx context.Context, id int64, at time.Time) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE tbl_users
		SET last_login_at = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`,
		id,
		at,
	)
	return err
}
