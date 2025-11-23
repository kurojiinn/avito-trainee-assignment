package repository

import (
	"avito-assignment/internal/model"
	"database/sql"

	"github.com/google/uuid"
)

type StatisticsRepository struct {
	DB *sql.DB
}

func NewStatisticsRepository(db *sql.DB) *StatisticsRepository {
	return &StatisticsRepository{DB: db}
}

// GetReviewStats возвращает общую статистику по назначениям
func (r *StatisticsRepository) GetReviewStats() (*model.ReviewStats, error) {
	stats := &model.ReviewStats{}

	// Общее количество назначений
	err := r.DB.QueryRow(`
		SELECT COUNT(*) FROM pr_reviewers
	`).Scan(&stats.TotalAssignments)
	if err != nil {
		return nil, err
	}

	// Статистика по пользователям
	userStatsQuery := `
		SELECT 
			u.id,
			u.username,
			COUNT(pr.reviewer_id) as assignments
		FROM users u
		LEFT JOIN pr_reviewers pr ON pr.reviewer_id = u.id
		GROUP BY u.id, u.username
		ORDER BY assignments DESC
		LIMIT 20
	`
	rows, err := r.DB.Query(userStatsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats.AssignmentsByUser = make([]model.UserAssignmentStats, 0)
	for rows.Next() {
		var userStat model.UserAssignmentStats
		var userID uuid.UUID
		err := rows.Scan(&userID, &userStat.Username, &userStat.Assignments)
		if err != nil {
			return nil, err
		}
		userStat.UserID = userID.String()
		stats.AssignmentsByUser = append(stats.AssignmentsByUser, userStat)
	}

	// Статистика по PR
	prStatsQuery := `
		SELECT 
			pr.id,
			pr.pull_request_name,
			COUNT(prr.reviewer_id) as reviewers_count,
			pr.status::text
		FROM pull_requests pr
		LEFT JOIN pr_reviewers prr ON prr.pr_id = pr.id
		GROUP BY pr.id, pr.pull_request_name, pr.status
		ORDER BY reviewers_count DESC
		LIMIT 20
	`
	rows, err = r.DB.Query(prStatsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats.AssignmentsByPR = make([]model.PRAssignmentStats, 0)
	for rows.Next() {
		var prStat model.PRAssignmentStats
		var prID uuid.UUID
		err := rows.Scan(&prID, &prStat.PRTitle, &prStat.ReviewersCount, &prStat.Status)
		if err != nil {
			return nil, err
		}
		prStat.PRID = prID.String()
		stats.AssignmentsByPR = append(stats.AssignmentsByPR, prStat)
	}

	// Общая статистика по PR
	err = r.DB.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'OPEN') as open,
			COUNT(*) FILTER (WHERE status = 'MERGED') as merged
		FROM pull_requests
	`).Scan(&stats.TotalPRs, &stats.OpenPRs, &stats.MergedPRs)
	if err != nil {
		return nil, err
	}

	// Среднее количество ревьюверов на PR
	if stats.TotalPRs > 0 {
		err = r.DB.QueryRow(`
			SELECT AVG(reviewer_count)::float
			FROM (
				SELECT pr.id, COUNT(prr.reviewer_id) as reviewer_count
				FROM pull_requests pr
				LEFT JOIN pr_reviewers prr ON prr.pr_id = pr.id
				GROUP BY pr.id
			) as pr_reviewer_counts
		`).Scan(&stats.AverageReviewersPerPR)
		if err != nil {
			return nil, err
		}
	}

	return stats, nil
}

