package model

// ReviewStats представляет статистику по назначениям ревьюверов
type ReviewStats struct {
	TotalAssignments     int                      `json:"total_assignments"`
	AssignmentsByUser    []UserAssignmentStats    `json:"assignments_by_user"`
	AssignmentsByPR      []PRAssignmentStats      `json:"assignments_by_pr"`
	TotalPRs             int                      `json:"total_prs"`
	OpenPRs              int                      `json:"open_prs"`
	MergedPRs            int                      `json:"merged_prs"`
	AverageReviewersPerPR float64                 `json:"average_reviewers_per_pr"`
}

// UserAssignmentStats представляет статистику назначений для пользователя
type UserAssignmentStats struct {
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	Assignments int    `json:"assignments"`
}

// PRAssignmentStats представляет статистику назначений для PR
type PRAssignmentStats struct {
	PRID        string `json:"pr_id"`
	PRTitle     string `json:"pr_title"`
	ReviewersCount int `json:"reviewers_count"`
	Status      string `json:"status"`
}

