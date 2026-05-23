package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"sentenceminer/config"
	"sentenceminer/internal/user"
	"sentenceminer/pkg/utils"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Service struct {
	oauthConfig *oauth2.Config
	userRepo    *user.Repository
	jwtSecret   string
}

func NewService(cfg config.AppConfig, userRepo *user.Repository) *Service {
	return &Service{
		oauthConfig: &oauth2.Config{
			RedirectURL:  cfg.GoogleRedirectURL,
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
			Endpoint:     google.Endpoint,
		},
		userRepo:  userRepo,
		jwtSecret: cfg.JWTSecret,
	}
}

var ErrInvalidLogin = errors.New("invalid email or password")

func (s *Service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)
	if email == "" || password == "" {
		return LoginResponse{}, ErrInvalidLogin
	}

	u, err := s.userRepo.FindActiveByEmail(email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return LoginResponse{}, ErrInvalidLogin
		}
		return LoginResponse{}, err
	}

	if u.PasswordHash == nil {
		return LoginResponse{}, ErrInvalidLogin
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*u.PasswordHash), []byte(password)); err != nil {
		return LoginResponse{}, ErrInvalidLogin
	}

	now := time.Now().UTC()
	if err := s.userRepo.TouchLastLogin(u.ID, now); err != nil {
		return LoginResponse{}, err
	}
	u.LastLoginAt = &now

	sessionID, err := randomToken(32)
	if err != nil {
		return LoginResponse{}, err
	}

	if err := s.userRepo.UpdateLoginSession(u.ID, sessionID); err != nil {
		return LoginResponse{}, err
	}
	u.LoginSession = &sessionID

	token, err := utils.GenerateToken(u.ID, s.jwtSecret, 72*time.Hour)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		Token: token,
		User:  u,
	}, nil
}

func (s *Service) GetGoogleLoginURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state)
}

func (s *Service) HandleGoogleCallback(ctx context.Context, code string) (LoginResponse, error) {
	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("code exchange failed: %w", err)
	}

	googleUser, err := s.getGoogleUserInfo(ctx, token)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to get user info: %w", err)
	}

	// 1. Find user by google_id
	u, err := s.userRepo.FindByGoogleID(googleUser.ID)
	if err != nil && !errors.Is(err, user.ErrUserNotFound) {
		return LoginResponse{}, err
	}

	if errors.Is(err, user.ErrUserNotFound) {
		// 2. If not found by google_id, try finding by email
		u, err = s.userRepo.FindActiveByEmail(googleUser.Email)
		if err != nil && !errors.Is(err, user.ErrUserNotFound) {
			return LoginResponse{}, err
		}

		if errors.Is(err, user.ErrUserNotFound) {
			// 3. Create new user if not found at all
			u = user.User{
				Name:     googleUser.Name,
				Email:    googleUser.Email,
				GoogleID: &googleUser.ID,
				RoleID:   2, // Default role
				StatusID: 1, // Active
			}
			if err := s.userRepo.Create(&u); err != nil {
				return LoginResponse{}, fmt.Errorf("failed to create user: %w", err)
			}
		} else {
			// 4. Update existing email-only user with google_id
			// Note: We might need an Update method in repo, but for now let's keep it simple.
			// Ideally we should link the accounts.
		}
	}

	now := time.Now().UTC()
	if err := s.userRepo.TouchLastLogin(u.ID, now); err != nil {
		return LoginResponse{}, err
	}
	u.LastLoginAt = &now

	sessionID, err := randomToken(32)
	if err != nil {
		return LoginResponse{}, err
	}

	if err := s.userRepo.UpdateLoginSession(u.ID, sessionID); err != nil {
		return LoginResponse{}, err
	}
	u.LoginSession = &sessionID

	jwtToken, err := utils.GenerateToken(u.ID, s.jwtSecret, 72*time.Hour)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		Token: jwtToken,
		User:  u,
	}, nil
}

func (s *Service) getGoogleUserInfo(ctx context.Context, token *oauth2.Token) (GoogleUser, error) {
	client := s.oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return GoogleUser{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GoogleUser{}, fmt.Errorf("google api returned status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return GoogleUser{}, err
	}

	var googleUser GoogleUser
	if err := json.Unmarshal(data, &googleUser); err != nil {
		return GoogleUser{}, err
	}

	return googleUser, nil
}

func randomToken(size int) (string, error) {
	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
