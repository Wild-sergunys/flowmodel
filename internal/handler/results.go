package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Wild-sergunys/flowmodel/internal/middleware"
	"github.com/Wild-sergunys/flowmodel/internal/model"
	"github.com/Wild-sergunys/flowmodel/internal/repository"
	"github.com/Wild-sergunys/flowmodel/internal/service"
)

type ResultsHandler struct {
	calcRepo          repository.CalculationRepository
	materialRepo      repository.MaterialRepository
	paramRepo         repository.ParameterRepository
	materialParamRepo repository.MaterialParameterRepository
}

func NewResultsHandler(calcRepo repository.CalculationRepository, materialRepo repository.MaterialRepository) *ResultsHandler {
	return &ResultsHandler{
		calcRepo:     calcRepo,
		materialRepo: materialRepo,
	}
}

func (h *ResultsHandler) SetParamRepo(paramRepo repository.ParameterRepository) {
	h.paramRepo = paramRepo
}

func (h *ResultsHandler) SetMaterialParamRepo(materialParamRepo repository.MaterialParameterRepository) {
	h.materialParamRepo = materialParamRepo
}

func getUserID(r *http.Request) (int, bool) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(jwt.MapClaims)
	if !ok {
		return 0, false
	}
	userID := int(claims["user_id"].(float64))
	return userID, true
}

func (h *ResultsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}

	calcs, err := h.calcRepo.FindByUserID(r.Context(), userID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}

	if calcs == nil {
		calcs = []model.Calculation{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calcs)
}

func (h *ResultsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	calc, err := h.calcRepo.FindByID(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}
	if calc == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Расчёт не найден", nil)
		return
	}

	if calc.UserID != userID {
		WriteError(w, http.StatusForbidden, "forbidden", "Нет доступа к этому расчёту", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(calc)
}

func (h *ResultsHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	calc, err := h.calcRepo.FindByID(r.Context(), id)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка сервера", nil)
		return
	}
	if calc == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Расчёт не найден", nil)
		return
	}

	if calc.UserID != userID {
		WriteError(w, http.StatusForbidden, "forbidden", "Нет доступа к этому расчёту", nil)
		return
	}

	report := map[string]interface{}{
		"id":          calc.ID,
		"user_id":     calc.UserID,
		"material_id": calc.MaterialID,
		"created_at":  calc.CreatedAt,
		"input":       json.RawMessage(calc.InputJSON),
		"result":      json.RawMessage(calc.ResultJSON),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func (h *ResultsHandler) Download(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserID(r)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "unauthorized", "Не авторизован", nil)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "validation_error", "Неверный ID", nil)
		return
	}

	calc, err := h.calcRepo.FindByID(r.Context(), id)
	if err != nil || calc == nil {
		WriteError(w, http.StatusNotFound, "not_found", "Расчёт не найден", nil)
		return
	}

	if calc.UserID != userID {
		WriteError(w, http.StatusForbidden, "forbidden", "Нет доступа к этому расчёту", nil)
		return
	}

	// Получаем информацию о материале
	materialName := "неизвестный"
	if material, err := h.materialRepo.FindByID(r.Context(), calc.MaterialID); err == nil && material != nil {
		materialName = material.Name
	}

	// Получаем параметры материала
	materialParams := make(map[string]interface{})
	if h.materialParamRepo != nil {
		params, err := h.materialParamRepo.FindByMaterialID(r.Context(), calc.MaterialID)
		if err == nil {
			for k, v := range params {
				materialParams[k] = v
			}
		}
	}

	var inputData, resultData map[string]interface{}
	json.Unmarshal([]byte(calc.InputJSON), &inputData)
	json.Unmarshal([]byte(calc.ResultJSON), &resultData)

	calcData := map[string]interface{}{
		"id":              calc.ID,
		"material_id":     calc.MaterialID,
		"material_name":   materialName,
		"material_params": materialParams,
		"created_at":      calc.CreatedAt,
		"input":           inputData,
		"result":          resultData,
	}

	excelData, err := service.GenerateExcel(calcData)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "internal_error", "Ошибка генерации отчёта", nil)
		return
	}

	filename := fmt.Sprintf("report_%d.xlsx", id)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Write(excelData)
}
