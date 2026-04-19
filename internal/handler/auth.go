package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"flowmodel/internal/middleware"
	"flowmodel/internal/model"
	"flowmodel/internal/repository"
)

type AuthHandler struct {
	userRepo repository.UserRepository
	jwtKey   []byte
}

type RegisterRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	if req.Login == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Логин и пароль обязательны", nil)
		return
	}

	existing, err := h.userRepo.FindByLogin(r.Context(), req.Login)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}
	if existing != nil {
		WriteError(w, http.StatusBadRequest, "conflict", "Пользователь уже существует", nil)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка хэширования", nil)
		return
	}

	// По умолчанию создаем роль researcher
	user := &model.User{
		Login:        req.Login,
		PasswordHash: string(hash),
		Role:         "researcher",
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка создания пользователя", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": user.ID})
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
