package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidLogin = errors.New("invalid email or password")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		return LoginResponse{}, ErrInvalidLogin
	}

	user, err := s.repo.FindActiveByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return LoginResponse{}, ErrInvalidLogin
		}
		return LoginResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return LoginResponse{}, ErrInvalidLogin
	}

	now := time.Now().UTC()
	if err := s.repo.TouchLastLogin(ctx, user.ID, now); err != nil {
		return LoginResponse{}, err
	}
	user.LastLoginAt = &now

	token, err := randomToken(32)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		Token: token,
		User:  user,
	}, nil
}

func randomToken(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
