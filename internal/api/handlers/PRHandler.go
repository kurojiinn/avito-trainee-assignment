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
type PRHandler struct {
	Service *service.PRService
}

// CreatePRRequest представляет запрос на создание Pull Request.
type CreatePRRequest struct {
	Title    string    `json:"pull_request_name"`
	AuthorID uuid.UUID `json:"author_id"`
}

// ReassignReviewerRequest представляет запрос на переназначение ревьювера.
type ReassignReviewerRequest struct {
	ReviewerID uuid.UUID `json:"reviewer_id"`
}

// CreatePR обрабатывает HTTP POST запрос на создание Pull Request.
func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pr := &model.PullRequest{
		Title:    req.Title,
		AuthorID: req.AuthorID,
	}

	createdPR, err := h.Service.CreatePR(pr)
	if err != nil {
		if err.Error() == "author not found" {
			http.Error(w, "Автор/команда не найдены", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pullReq, errGetting := h.Service.GetPRByID(createdPR.ID)
	if errGetting != nil {
		http.Error(w, errGetting.Error(), http.StatusInternalServerError)
		return
	}
	if pullReq != nil {
		http.Error(w, "PR уже существует", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode("PR создан")
	if err != nil {
		return
	}
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
	err = json.NewEncoder(w).Encode(pr)
	if err != nil {
		return
	}
}

func (h *PRHandler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Читаем JSON
	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 2. Проверяем, что пришли нужные поля
	if req.PullRequestID == "" || req.OldUserID == "" {
		http.Error(w, `{"error":"pull_request_id and old_user_id are required"}`, http.StatusBadRequest)
		return
	}

	// 3. Парсим UUID
	prID, err := uuid.Parse(req.PullRequestID)
	if err != nil {
		http.Error(w, `{"error":"invalid pull_request_id (must be UUID)"}`, http.StatusBadRequest)
		return
	}

	oldUserID, err := uuid.Parse(req.OldUserID)
	if err != nil {
		http.Error(w, `{"error":"invalid old_user_id (must be UUID)"}`, http.StatusBadRequest)
		return
	}

	// 4. Вызываем доменную логику
	updatedPR, replacedBy, err := h.Service.ReassignReviewer(prID, oldUserID)
	if err != nil {
		switch err.Error() {
		case "pull request not found", "user not found":
			w.WriteHeader(http.StatusNotFound)
			err = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"code":    "NOT_FOUND",
					"message": "PR или пользователь не найден\n",
				},
			})
			if err != nil {
				return
			}
			return

		case "cannot reassign reviewers for merged PR":
			w.WriteHeader(http.StatusConflict)
			err = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"code":    "PR_MERGED",
					"message": "Нельзя менять после MERGED",
				},
			})
			if err != nil {
				return
			}
			return

		case "reviewer not assigned to this PR":
			w.WriteHeader(http.StatusConflict)
			err = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"code":    "NOT_ASSIGNED",
					"message": "Пользователь не был назначен ревьювером",
				},
			})
			if err != nil {
				return
			}
			return

		case "no available reviewers in the team":
			w.WriteHeader(http.StatusConflict)
			err = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"code":    "NO_CANDIDATE",
					"message": "Нет доступных кандидатов",
				},
			})
			if err != nil {
				return
			}
			return
		}

		// fallback
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Возвращаем успешный ответ
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Переназначение выполнено",
		"pr":          updatedPR,
		"replaced_by": replacedBy,
	})
	if err != nil {
		return
	}
}

func (h *PRHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Читаем JSON в структуру
	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	// 2. Проверяем, что значение пришло
	if req.PullRequestID == "" {
		http.Error(w, `{"error": "pull_request_id is required"}`, http.StatusBadRequest)
		return
	}

	// 3. Парсим UUID
	prID, err := uuid.Parse(req.PullRequestID)
	if err != nil {
		http.Error(w, `{"error": "invalid pull_request_id (must be UUID)"}`, http.StatusBadRequest)
		return
	}

	// 4. Вызываем сервис
	_, errMerged := h.Service.MergePR(prID)
	if errMerged != nil {
		if errMerged.Error() == "pull request not found" {
			w.WriteHeader(http.StatusNotFound)
			err = json.NewEncoder(w).Encode(map[string]string{
				"error": "PR не найден",
			})
			if err != nil {
				return
			}
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		err = json.NewEncoder(w).Encode(map[string]string{
			"error": errMerged.Error(),
		})
		if err != nil {
			return
		}
		return
	}

	// 5. Возвращаем PR
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"pr": "PR в состоянии MERGED",
	})
	if err != nil {
		return
	}
}

func (h *PRHandler) GetAllPRs(w http.ResponseWriter, r *http.Request) {
	prs, err := h.Service.GetAllPRs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(prs)
	if err != nil {
		return
	}
}
