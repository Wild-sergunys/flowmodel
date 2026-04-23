package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"flowmodel/internal/middleware"
	"flowmodel/internal/model"
	"flowmodel/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	repo repository.UserRepository
}

func NewUserHandler(repo repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := h.repo.FindAll(r.Context())
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	u, err := h.repo.FindByID(r.Context(), id)
	switch {
	case err != nil:
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	case u == nil:
		WriteError(w, http.StatusNotFound, "not_found", "Пользователь не найден", nil)
		return
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u)
	}
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	if req.Role != "admin" && req.Role != "researcher" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Роль должна быть admin или researcher", nil)
		return
	}

	user := &model.User{ID: id, Role: req.Role}
	if err := h.repo.Update(r.Context(), user); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка обновления", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь обновлён"})
}

func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}
	currentUserID := int(claims["user_id"].(float64))

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	if id == currentUserID {
		WriteError(w, http.StatusForbidden, "forbidden", "Нельзя удалить самого себя", nil)
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка удаления", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Пользователь удалён"})
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный формат запроса", nil)
		return
	}

	if req.Login == "" || req.Password == "" || req.Role == "" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Логин, пароль и роль обязательны", nil)
		return
	}

	if req.Role != "admin" && req.Role != "researcher" {
		WriteError(w, http.StatusBadRequest, "validation_error", "Роль должна быть admin или researcher", nil)
		return
	}

	existing, _ := h.repo.FindByLogin(r.Context(), req.Login)
	if existing != nil {
		WriteError(w, http.StatusBadRequest, "conflict", "Пользователь уже существует", nil)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := &model.User{Login: req.Login, PasswordHash: string(hash), Role: req.Role}

	if err := h.repo.Create(r.Context(), user); err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка создания", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"id": user.ID})
}
