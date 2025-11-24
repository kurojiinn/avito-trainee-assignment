package model

import (
	"time"

	"github.com/google/uuid"
)

// User представляет участника команды.
type User struct {
	Username string    `json:"username"`
	ID       uuid.UUID `json:"user_id"`
	TeamID   uuid.UUID `json:"team_id"`
	IsActive bool      `json:"is_active"`
}

// Team представляет команду пользователей.
type Team struct {
	Name    string    `json:"team_name"`
	Members []User    `json:"members,omitempty"`
	ID      uuid.UUID `json:"team_id"`
}

// PRStatus описывает статус Pull Request.
type PRStatus string

const (
	OPEN   PRStatus = "OPEN"
	MERGED PRStatus = "MERGED"
)

// PullRequest представляет Pull Request с назначенными ревьюверами.
type PullRequest struct {
	Title     string      `json:"pull_request_name"`
	ID        uuid.UUID   `json:"pull_request_id"`
	AuthorID  uuid.UUID   `json:"author_id"`
	Reviewers []uuid.UUID `json:"reviewers"`
	Status    PRStatus    `json:"status"`
	CreatedAt time.Time   `json:"createdAt"`
	MergedAt  *time.Time  `json:"mergedAt,omitempty"`
}
