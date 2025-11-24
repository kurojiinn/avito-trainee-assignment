// Package repository предоставляет функции для работы с базой данных.
package repository

import (
	"avito-assignment/internal/model"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// UserRepository предоставляет методы для работы с пользователями в базе данных.
type UserRepository struct {
	DB *sql.DB
}

// NewUserRepository создает новый экземпляр UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUser сохраняет нового пользователя в базе данных.
func (r *UserRepository) CreateUser(user *model.User) error {
	query := `
		INSERT INTO users (id, username, team_id, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.DB.Exec(query, user.ID, user.Username, user.TeamID, user.IsActive, time.Now())
	return err
}

// GetUserByID возвращает пользователя по его уникальному идентификатору.
func (r *UserRepository) GetUserByID(id uuid.UUID) (*model.User, error) {
	query := `
		SELECT id, username, team_id, is_active
		FROM users
		WHERE id = $1
	`
	row := r.DB.QueryRow(query, id)
	var u model.User
	err := row.Scan(&u.ID, &u.Username, &u.TeamID, &u.IsActive)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Update обновляет пользователя.
func (r *UserRepository) Update(user *model.User) (*model.User, error) {
	query := `
		UPDATE users
		SET username=$1, team_id=$2, is_active=$3
		WHERE id=$4
		RETURNING id, username, team_id, is_active
	`
	row := r.DB.QueryRow(query, user.Username, user.TeamID, user.IsActive, user.ID)
	var u model.User
	err := row.Scan(&u.ID, &u.Username, &u.TeamID, &u.IsActive)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Delete удаляет пользователя
func (r *UserRepository) Delete(id uuid.UUID) error {
	_, err := r.DB.Exec("DELETE FROM users WHERE id=$1", id)
	return err
}

// GetPRsByReviewer возвращает список всех Pull Request, где указанный пользователь
// назначен ревьювером.
func (r *UserRepository) GetPRsByReviewer(userID uuid.UUID) ([]model.PullRequest, error) {
	query := `
		SELECT pr.id, pr.pull_request_name, pr.author_id, pr.status, pr.created_at, pr.merged_at
		FROM pull_requests pr
		JOIN pr_reviewers rr ON rr.pr_id = pr.id
		WHERE rr.reviewer_id = $1
		ORDER BY pr.created_at DESC
	`
	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			return
		}
	}()

	var prs []model.PullRequest
	for rows.Next() {
		var pr model.PullRequest
		err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
		if err != nil {
			return nil, err
		}

		reviewersQuery := `SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $1`
		reviewerRows, err := r.DB.Query(reviewersQuery, pr.ID)
		if err != nil {
			return nil, err
		}

		var reviewers []uuid.UUID
		for reviewerRows.Next() {
			var reviewerID uuid.UUID
			if err := reviewerRows.Scan(&reviewerID); err != nil {
				err = reviewerRows.Close()
				if err != nil {
					return nil, err
				}
				return nil, err
			}
			reviewers = append(reviewers, reviewerID)
		}
		err = reviewerRows.Close()
		if err != nil {
			return nil, err
		}
		pr.Reviewers = reviewers
		prs = append(prs, pr)
		if err := reviewerRows.Err(); err != nil {
			return nil, err
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return prs, nil
}

// GetActiveUsersByTeam возвращает список активных пользователей команды,
func (r *UserRepository) GetActiveUsersByTeam(teamID, excludeID uuid.UUID) ([]model.User, error) {
	query := `
		SELECT id, username, team_id, is_active
		FROM users
		WHERE team_id = $1 AND is_active = true AND id != $2
		ORDER BY username
	`
	rows, err := r.DB.Query(query, teamID, excludeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		err := rows.Scan(&u.ID, &u.Username, &u.TeamID, &u.IsActive)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// DeactivateTeamMembers массово деактивирует всех пользователей команды
func (r *UserRepository) DeactivateTeamMembers(teamID uuid.UUID) (int, error) {
	query := `
		UPDATE users
		SET is_active = false
		WHERE team_id = $1 AND is_active = true
	`
	result, err := r.DB.Exec(query, teamID)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}
func (r *UserRepository) GetUsersByTeam(teamID uuid.UUID) ([]model.User, error) {
	query := `
			SELECT id, username, team_id, is_active
			FROM users
			WHERE team_id = $1 AND is_active = true
			ORDER BY username
		`
	rows, err := r.DB.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		err := rows.Scan(&u.ID, &u.Username, &u.TeamID, &u.IsActive)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

// GetActiveUsersByTeamExcluding возвращает активных пользователей команды, исключая несколько пользователей
func (r *UserRepository) GetActiveUsersByTeamExcluding(teamID uuid.UUID, excludeIDs []uuid.UUID) ([]model.User, error) {
	if len(excludeIDs) == 0 {
		query := `
			SELECT id, username, team_id, is_active
			FROM users
			WHERE team_id = $1 AND is_active = true
			ORDER BY username
		`
		rows, err := r.DB.Query(query, teamID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var users []model.User
		for rows.Next() {
			var u model.User
			err := rows.Scan(&u.ID, &u.Username, &u.TeamID, &u.IsActive)
			if err != nil {
				return nil, err
			}
			users = append(users, u)
		}
		return users, nil
	}
	query := `
		SELECT id, username, team_id, is_active
		FROM users
		WHERE team_id = $1 AND is_active = true
	`
	args := []interface{}{teamID}
	excludeConditions := make([]string, len(excludeIDs))
	for i, excludeID := range excludeIDs {
		excludeConditions[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, excludeID)
	}
	if len(excludeConditions) > 0 {
		query += " AND id NOT IN (" + strings.Join(excludeConditions, ", ") + ")"
	}
	query += ` ORDER BY username`

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			panic(err)
		}
	}()

	var users []model.User
	for rows.Next() {
		var u model.User
		err := rows.Scan(&u.ID, &u.Username, &u.TeamID, &u.IsActive)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
