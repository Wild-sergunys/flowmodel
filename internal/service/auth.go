package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"flowmodel/internal/model"
	"flowmodel/internal/repository"
)

type AuthService struct {
	userRepo repository.UserRepository
	jwtKey   []byte
}

func NewAuthService(userRepo repository.UserRepository, jwtKey string) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtKey:   []byte(jwtKey),
	}
}

func (s *AuthService) Login(ctx context.Context, login, password string) (*model.User, string, error) {
	user, err := s.userRepo.FindByLogin(ctx, login)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"login":   user.Login,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		return nil, "", err
	}

	return user, tokenString, nil
}

func (s *AuthService) GetUser(ctx context.Context, userID int) (*model.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

var (
	ErrInvalidCredentials = &AuthError{Message: "Неверный логин или пароль"}
)

type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
