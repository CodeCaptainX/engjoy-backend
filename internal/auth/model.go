package auth

import "sentenceminer/internal/user"

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string    `json:"token"`
	User  user.User `json:"user"`
}
