package handlers

import (
	"avito-assignment/internal/service"
	"encoding/json"
	"net/http"
)

type StatisticsHandler struct {
	Service *service.StatisticsService
}

func (h *StatisticsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.Service.GetReviewStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(stats)
	if err != nil {
		return
	}
}
