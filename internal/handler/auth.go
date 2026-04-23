package handler

import (
	"encoding/json"
	"net/http"

	"flowmodel/internal/middleware"
	"flowmodel/internal/service"

	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	authService *service.AuthService
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
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

	user, token, err := h.authService.Login(r.Context(), req.Login, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Неверный логин или пароль", nil)
			return
		}
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}

	// Сбрасываем счётчик попыток при успешном входе
	if loginRateLimiter != nil {
		ip := middleware.GetIP(r)
		loginRateLimiter.Reset(ip)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token: token,
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
	user, err := h.authService.GetUser(r.Context(), userID)
	if err != nil || user == nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
