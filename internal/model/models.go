// Package model содержит доменные модели данных.
// Эти структуры представляют бизнес-сущности приложения и не зависят от других слоев.
package model

import (
	"time"

	"github.com/google/uuid"
)

// User представляет участника команды.
// Пользователь может быть назначен ревьювером на Pull Request, если IsActive = true.
type User struct {
	// ID - уникальный идентификатор пользователя (UUID)
	ID uuid.UUID `json:"user_id"`
	
	// Username - имя пользователя (должно быть уникальным)
	Username string `json:"username"`
	
	// TeamID - идентификатор команды, к которой принадлежит пользователь
	// Может быть пустым (nil UUID), если пользователь не в команде
	TeamID uuid.UUID `json:"team_id"`
	
	// IsActive - флаг активности пользователя
	// Если false, пользователь не может быть назначен ревьювером
	IsActive bool `json:"is_active"`
}

// Team представляет команду пользователей.
// Команда - это группа пользователей, которые могут ревьювить PR друг друга.
type Team struct {
	// ID - уникальный идентификатор команды (UUID)
	ID uuid.UUID `json:"team_id"`
	
	// Name - название команды (должно быть уникальным)
	Name string `json:"team_name"`
	
	// Members - список участников команды
	// Заполняется при получении команды через API (опциональное поле)
	Members []User `json:"members,omitempty"`
}

// PRStatus описывает статус Pull Request.
// Используется для контроля жизненного цикла PR.
type PRStatus string

const (
	// OPEN - PR открыт, можно изменять список ревьюверов
	OPEN PRStatus = "OPEN"
	
	// MERGED - PR объединен, изменения ревьюверов запрещены
	MERGED PRStatus = "MERGED"
)

// PullRequest представляет Pull Request с назначенными ревьюверами.
// PR создается автором и автоматически получает до 2 активных ревьюверов из команды автора.
type PullRequest struct {
	// ID - уникальный идентификатор PR (UUID)
	ID uuid.UUID `json:"pull_request_id"`
	
	// Title - название Pull Request
	Title string `json:"pull_request_name"`
	
	// AuthorID - идентификатор автора PR
	// Автор не может быть назначен ревьювером своего PR
	AuthorID uuid.UUID `json:"author_id"`
	
	// Status - текущий статус PR (OPEN или MERGED)
	Status PRStatus `json:"status"`
	
	// Reviewers - список идентификаторов назначенных ревьюверов
	// Максимум 2 ревьювера, минимум 0 (если нет доступных кандидатов)
	Reviewers []uuid.UUID `json:"reviewers"`
	
	// CreatedAt - время создания PR
	CreatedAt time.Time `json:"createdAt"`
	
	// MergedAt - время объединения PR (nil, если PR еще не мержен)
	MergedAt *time.Time `json:"mergedAt,omitempty"`
}
