package model

import (
	"time"

	"github.com/google/uuid"
)

// User представляет участника команды.
type User struct {
	ID       uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	TeamID   uuid.UUID `json:"team_id"`
	IsActive bool      `json:"is_active"`
}

// Team представляет команду пользователей.
type Team struct {
	ID      uuid.UUID `json:"team_id"`
	Name    string    `json:"team_name"`
	Members []User    `json:"members,omitempty"`
}

// PRStatus описывает статус Pull Request.
type PRStatus string

const (
	OPEN   PRStatus = "OPEN"
	MERGED PRStatus = "MERGED"
)

// PullRequest представляет Pull Request с назначенными ревью верами
type PullRequest struct {
	ID        uuid.UUID   `json:"pull_request_id"`
	Title     string      `json:"pull_request_name"`
	AuthorID  uuid.UUID   `json:"author_id"`
	Status    PRStatus    `json:"status"`
	Reviewers []uuid.UUID `json:"reviewers"`
	CreatedAt time.Time   `json:"createdAt"`
	MergedAt  *time.Time  `json:"mergedAt,omitempty"`
}
