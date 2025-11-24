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
//
// Параметры:
//   - user: объект пользователя с заполненными полями
//   - ID должен быть сгенерирован заранее (uuid.New())
//   - Username должен быть уникальным
//   - TeamID может быть пустым (nil UUID)
//   - IsActive определяет, может ли пользователь быть ревьювером
//
// Возвращает:
//   - error: ошибка, если не удалось сохранить
//   - Может быть ошибка уникальности username
//   - Может быть ошибка внешнего ключа team_id
//
// Пример использования:
//
//	user := &model.User{
//	    ID: uuid.New(),
//	    Username: "john_doe",
//	    TeamID: teamID,
//	    IsActive: true,
//	}
//	err := repo.CreateUser(user)
func (r *UserRepository) CreateUser(user *model.User) error {
	query := `
		INSERT INTO users (id, username, team_id, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.DB.Exec(query, user.ID, user.Username, user.TeamID, user.IsActive, time.Now())
	return err
}

// GetUserByID возвращает пользователя по его уникальному идентификатору.
//
// Параметры:
//   - id: UUID пользователя
//
// Возвращает:
//   - *model.User: найденный пользователь
//   - error: ошибка, если пользователь не найден (sql.ErrNoRows)
//
// Пример использования:
//
//	user, err := repo.GetUserByID(userID)
//	if err != nil {
//	    // Пользователь не найден
//	}
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
//
// Параметры:
//   - user: объект пользователя с обновленными полями (ID должен быть заполнен)
//
// Возвращает:
//   - *model.User: обновленный пользователь
//   - error: ошибка, если пользователь не найден или произошла ошибка БД
//
// Пример использования:
//
//	user.Username = "new_username"
//	user.IsActive = false
//	updated, err := repo.Update(user)
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
//
// Используется для получения списка PR, которые пользователь должен отревьювить.
// Результаты отсортированы по дате создания (новые первыми).
//
// Параметры:
//   - userID: UUID пользователя-ревьювера
//
// Возвращает:
//   - []model.PullRequest: список PR с полной информацией, включая список всех ревьюверов
//   - error: ошибка выполнения запроса
//
// Пример использования:
//
//	prs, err := repo.GetPRsByReviewer(reviewerID)
//	for _, pr := range prs {
//	    // Обработка каждого PR
//	}
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
		// Загружаем список всех ревьюверов для этого PR
		// Это необходимо, так как в основном запросе мы получаем только информацию о PR,
		// а список ревьюверов хранится в отдельной таблице pr_reviewers
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
// исключая указанного пользователя.
//
// Используется при назначении ревьюверов на PR, чтобы исключить автора
// из списка кандидатов на ревью.
//
// Параметры:
//   - teamID: UUID команды
//   - excludeID: UUID пользователя, которого нужно исключить из результата
//     (обычно это автор PR)
//
// Возвращает:
//   - []model.User: список активных пользователей (is_active = true)
//     Результаты отсортированы по username
//   - error: ошибка выполнения запроса
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
	// Build query with exclusions
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
