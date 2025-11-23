package service

import (
	"avito-assignment/internal/model"
	"avito-assignment/internal/repository"
)

type StatisticsService struct {
	statsRepo *repository.StatisticsRepository
}

func NewStatisticsService(statsRepo *repository.StatisticsRepository) *StatisticsService {
	return &StatisticsService{statsRepo: statsRepo}
}

// GetReviewStats возвращает статистику по назначениям
func (s *StatisticsService) GetReviewStats() (*model.ReviewStats, error) {
	return s.statsRepo.GetReviewStats()
}

