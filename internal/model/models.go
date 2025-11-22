package model

import (
	"time"

	"github.com/google/uuid"
)

// User представляет участника команды.
type User struct {
	ID       uuid.UUID
	Name     string
	TeamID   uuid.UUID
	IsActive bool
}

// Team представляет команду пользователей.
type Team struct {
	ID      uuid.UUID
	Name    string
	Members []User
}

// PRStatus описывает статус Pull Request.
type PRStatus string

const (
	OPEN   PRStatus = "OPEN"
	MERGED PRStatus = "MERGED"
)

// PR представляет Pull Request с назначенными ревьюверами.
type PR struct {
	ID        uuid.UUID   `json:"pull_request_id"`
	Title     string      `json:"pull_request_name"`
	AuthorID  uuid.UUID   `json:"author_id"`
	Status    PRStatus    `json:"status"`
	Reviewers []uuid.UUID `json:"reviewers"`
	CreatedAt time.Time   `json:"createdAt"`
	MergedAt  *time.Time  `json:"mergedAt,omitempty"`
}
