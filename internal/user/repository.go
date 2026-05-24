package user

import (
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

func (r *Repository) FindActiveByEmail(email string) (User, error) {
	var user User
	err := r.db.QueryRowx(
		`SELECT id, uuid, status_id, name, email, password_hash, google_id, role_id, last_login_at, created_at, updated_at, deleted_at
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

func (r *Repository) FindByGoogleID(googleID string) (User, error) {
	var user User
	err := r.db.QueryRowx(
		`SELECT id, uuid, status_id, name, email, password_hash, google_id, role_id, last_login_at, created_at, updated_at, deleted_at
		FROM tbl_users
		WHERE google_id = $1 AND status_id = 1 AND deleted_at IS NULL`,
		googleID,
	).StructScan(&user)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *Repository) Create(user *User) error {
	query := `
		INSERT INTO tbl_users (name, email, password_hash, google_id, role_id, status_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, uuid, created_at, updated_at`

	return r.db.QueryRowx(
		query,
		user.Name,
		strings.ToLower(user.Email),
		user.PasswordHash,
		user.GoogleID,
		user.RoleID,
		user.StatusID,
	).Scan(&user.ID, &user.UUID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *Repository) TouchLastLogin(id int64, at time.Time) error {
	_, err := r.db.Exec(
		`UPDATE tbl_users
		SET last_login_at = $2, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL`,
		id,
		at,
	)
	return err
}

func (r *Repository) UpdateLoginSession(id int64, session string) error {
	_, err := r.db.Exec(
		`UPDATE tbl_users
		 SET login_session = $2, updated_at = NOW()
		 WHERE id = $1 AND deleted_at IS NULL`,
		id,
		session,
	)
	return err
}

func (r *Repository) AddFavorite(userID int64, sentenceUUID string) error {
	_, err := r.db.Exec(
		`INSERT INTO tbl_users_sentences_favourites (user_id, sentence_id)
		 SELECT $1, id FROM tbl_sentences WHERE uuid = $2
		 ON CONFLICT (user_id, sentence_id) DO NOTHING`,
		userID,
		sentenceUUID,
	)
	return err
}

func (r *Repository) RemoveFavorite(userID int64, sentenceUUID string) error {
	_, err := r.db.Exec(
		`UPDATE tbl_users_sentences_favourites
		 SET deleted_at = NOW()
		 WHERE user_id = $1 
		 AND sentence_id = (SELECT id FROM tbl_sentences WHERE uuid = $2) 
		 AND deleted_at IS NULL`,
		userID,
		sentenceUUID,
	)
	return err
}

func (r *Repository) ListFavorites(userID int64) ([]string, error) {
	var uuids []string
	err := r.db.Select(
		&uuids,
		`SELECT s.uuid
		 FROM tbl_users_sentences_favourites f
		 JOIN tbl_sentences s ON f.sentence_id = s.id
		 WHERE f.user_id = $1 AND f.deleted_at IS NULL
		 ORDER BY f.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return uuids, nil
}
