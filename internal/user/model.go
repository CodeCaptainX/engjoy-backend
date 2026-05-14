package user

import "time"

type User struct {
	ID           int64      `db:"id" json:"id"`
	UUID         string     `db:"uuid" json:"uuid"`
	StatusID     int64      `db:"status_id" json:"statusId"`
	Name         string     `db:"name" json:"name"`
	Email        string     `db:"email" json:"email"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Role         string     `db:"role" json:"role"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"lastLoginAt,omitempty"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deletedAt,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
