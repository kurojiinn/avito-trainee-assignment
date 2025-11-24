package repository

import (
	"avito-assignment/internal/model"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type TeamRepository struct {
	DB *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{DB: db}
}

// Create создает команду
func (r *TeamRepository) Create(team *model.Team) error {
	query := `
		INSERT INTO teams (id, name, created_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.DB.Exec(query, team.ID, team.Name, time.Now())
	return err
}

// GetByID возвращает команду по ID
func (r *TeamRepository) GetByID(id uuid.UUID) (*model.Team, error) {
	query := `
		SELECT id, name
		FROM teams
		WHERE id = $1
	`
	row := r.DB.QueryRow(query, id)
	var team model.Team
	err := row.Scan(&team.ID, &team.Name)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetByName возвращает команду по имени
func (r *TeamRepository) GetByName(name string) (*model.Team, error) {
	query := `
		SELECT id, name
		FROM teams
		WHERE name = $1
	`
	row := r.DB.QueryRow(query, name)
	var team model.Team
	err := row.Scan(&team.ID, &team.Name)
	if err != nil {
		return nil, err
	}
	return &team, nil
}

// GetMembers возвращает всех участников команды
func (r *TeamRepository) GetMembers(teamID uuid.UUID) ([]model.User, error) {
	query := `
		SELECT id, username, team_id, is_active
		FROM users
		WHERE team_id = $1
		ORDER BY username
	`
	rows, err := r.DB.Query(query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []model.User
	for rows.Next() {
		var u model.User
		err := rows.Scan(&u.ID, &u.Username, &u.TeamID, &u.IsActive)
		if err != nil {
			return nil, err
		}
		members = append(members, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

// Update обновляет команду
func (r *TeamRepository) Update(team *model.Team) error {
	query := `
		UPDATE teams
		SET name = $1
		WHERE id = $2
	`
	_, err := r.DB.Exec(query, team.Name, team.ID)
	return err
}

// Delete удаляет команду
func (r *TeamRepository) Delete(id uuid.UUID) error {
	_, err := r.DB.Exec("DELETE FROM teams WHERE id = $1", id)
	return err
}
