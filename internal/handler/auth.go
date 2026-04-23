package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"flowmodel/internal/middleware"
	"flowmodel/internal/repository"
)

type AuthHandler struct {
	userRepo repository.UserRepository
	jwtKey   []byte
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

func NewAuthHandler(userRepo repository.UserRepository, jwtKey string) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
		jwtKey:   []byte(jwtKey),
	}
}

var loginRateLimiter *middleware.RateLimiter

func SetLoginRateLimiter(rl *middleware.RateLimiter) {
	loginRateLimiter = rl
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	if req.Login == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Логин и пароль обязательны", nil)
		return
	}

	user, err := h.userRepo.FindByLogin(r.Context(), req.Login)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}
	if user == nil {
		WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Неверный логин или пароль", nil)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Неверный логин или пароль", nil)
		return
	}

	// Сбрасываем счётчик попыток при успешном входе
	if loginRateLimiter != nil {
		ip := middleware.GetIP(r)
		loginRateLimiter.Reset(ip)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"login":   user.Login,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(h.jwtKey)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка создания токена", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token: tokenString,
		Role:  user.Role,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Выход выполнен успешно"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}

	userID := int(claims["user_id"].(float64))
	user, err := h.userRepo.FindByID(r.Context(), userID)
	if err != nil || user == nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
