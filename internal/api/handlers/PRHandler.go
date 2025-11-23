// Package handlers содержит HTTP обработчики для API endpoints.
// Обработчики отвечают за парсинг HTTP запросов, валидацию данных,
// вызов сервисов и формирование HTTP ответов.
package handlers

import (
	"avito-assignment/internal/model"
	"avito-assignment/internal/service"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// PRHandler обрабатывает HTTP запросы, связанные с Pull Requests.
// Делегирует бизнес-логику сервису PRService.
type PRHandler struct {
	Service *service.PRService // Сервис с бизнес-логикой для PR
}

// CreatePRRequest представляет запрос на создание Pull Request.
// Используется для десериализации JSON из HTTP запроса.
type CreatePRRequest struct {
	Title    string    `json:"pull_request_name"` // Название PR
	AuthorID uuid.UUID `json:"author_id"`         // ID автора PR
}

// ReassignReviewerRequest представляет запрос на переназначение ревьювера.
// Используется для десериализации JSON из HTTP запроса.
type ReassignReviewerRequest struct {
	ReviewerID uuid.UUID `json:"reviewer_id"` // ID ревьювера для замены
}

// CreatePR обрабатывает HTTP POST запрос на создание Pull Request.
// Endpoint: POST /api/v1/pull-requests
//
// Процесс:
//  1. Парсит JSON из тела запроса
//  2. Создает модель PR
//  3. Вызывает сервис для создания PR (автоматически назначаются ревьюверы)
//  4. Возвращает созданный PR с кодом 201 Created
//
// Ошибки:
//   - 400 Bad Request: неверный формат JSON
//   - 404 Not Found: автор не найден
//   - 500 Internal Server Error: внутренняя ошибка сервера
func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	// Парсим JSON из тела запроса
	var req CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Создаем модель PR из запроса
	pr := &model.PullRequest{
		Title:    req.Title,
		AuthorID: req.AuthorID,
	}

	// Вызываем сервис для создания PR
	// Сервис автоматически назначит до 2 ревьюверов из команды автора
	createdPR, err := h.Service.CreatePR(pr)
	if err != nil {
		// Обрабатываем специфичные ошибки
		if err.Error() == "author not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Возвращаем созданный PR с кодом 201 Created
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdPR)
}

func (h *PRHandler) GetPR(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["pull_request_id"]
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid UUID", http.StatusBadRequest)
		return
	}

	pr, err := h.Service.GetPRByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pr)
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	prIDStr := vars["pull_request_id"]
	prID, err := uuid.Parse(prIDStr)
	if err != nil {
		http.Error(w, "invalid PR UUID", http.StatusBadRequest)
		return
	}

	var req ReassignReviewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedPR, err := h.Service.ReassignReviewer(prID, req.ReviewerID)
	if err != nil {
		if err.Error() == "pull request not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err.Error() == "cannot reassign reviewers for merged PR" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error() == "reviewer not assigned to this PR" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err.Error() == "no available reviewers in the team" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedPR)
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	prIDStr := vars["pull_request_id"]
	prID, err := uuid.Parse(prIDStr)
	if err != nil {
		http.Error(w, "invalid UUID", http.StatusBadRequest)
		return
	}

	mergedPR, err := h.Service.MergePR(prID)
	if err != nil {
		if err.Error() == "pull request not found" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mergedPR)
}

func (h *PRHandler) GetAllPRs(w http.ResponseWriter, r *http.Request) {
	prs, err := h.Service.GetAllPRs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prs)
}
