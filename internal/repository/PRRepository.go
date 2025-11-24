package repository

import (
	"avito-assignment/internal/model"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type PRRepository struct {
	DB *sql.DB
}

func NewPRRepository(db *sql.DB) *PRRepository {
	return &PRRepository{DB: db}
}

// Create создает PR и назначает ревьюверов
func (r *PRRepository) Create(pr *model.PullRequest, reviewers []uuid.UUID) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil {
			return
		}
	}()

	query := `
		INSERT INTO pull_requests (id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(query, pr.ID, pr.Title, pr.AuthorID, pr.Status, pr.CreatedAt)
	if err != nil {
		return err
	}

	for _, reviewerID := range reviewers {
		reviewerQuery := `
			INSERT INTO pr_reviewers (pr_id, reviewer_id, assigned_at)
			VALUES ($1, $2, $3)
		`
		_, err = tx.Exec(reviewerQuery, pr.ID, reviewerID, time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByID возвращает PR по ID с ревьюверами
func (r *PRRepository) GetByID(id uuid.UUID) (*model.PullRequest, error) {
	query := `
		SELECT id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1
	`
	row := r.DB.QueryRow(query, id)
	var pr model.PullRequest
	err := row.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		return nil, err
	}

	reviewersQuery := `
		SELECT reviewer_id
		FROM pr_reviewers
		WHERE pr_id = $1
		ORDER BY assigned_at
	`
	reviewerRows, errQuery := r.DB.Query(reviewersQuery, id)
	if errQuery != nil {
		return nil, errQuery
	}
	defer func() {
		err = reviewerRows.Close()
		if err != nil {
			return
		}
	}()
	if err := reviewerRows.Err(); err != nil {
		return nil, err
	}
	var reviewers []uuid.UUID
	for reviewerRows.Next() {
		var reviewerID uuid.UUID
		if err := reviewerRows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, reviewerID)
	}
	pr.Reviewers = reviewers

	return &pr, nil
}

// Update обновляет PR
func (r *PRRepository) Update(pr *model.PullRequest) error {
	query := `
		UPDATE pull_requests
		SET pull_request_name = $1, status = $2, merged_at = $3
		WHERE id = $4
	`
	_, err := r.DB.Exec(query, pr.Title, pr.Status, pr.MergedAt, pr.ID)
	return err
}

// ReassignReviewer заменяет одного ревьювера на другого в указанном PR.
func (r *PRRepository) ReassignReviewer(prID, oldReviewerID, newReviewerID uuid.UUID) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil {
			return
		}
	}()

	var status string
	var reviewerExists bool
	checkQuery := `
		SELECT pr.status, EXISTS(
			SELECT 1 FROM pr_reviewers 
			WHERE pr_id = $1 AND reviewer_id = $2
		)
		FROM pull_requests pr
		WHERE pr.id = $1
	`
	err = tx.QueryRow(checkQuery, prID, oldReviewerID).Scan(&status, &reviewerExists)
	if err != nil {
		return err
	}
	if status == "MERGED" {
		return sql.ErrNoRows
	}
	if !reviewerExists {
		return sql.ErrNoRows
	}

	deleteQuery := `
		DELETE FROM pr_reviewers
		WHERE pr_id = $1 AND reviewer_id = $2
	`
	_, err = tx.Exec(deleteQuery, prID, oldReviewerID)
	if err != nil {
		return err
	}

	insertQuery := `
		INSERT INTO pr_reviewers (pr_id, reviewer_id, assigned_at)
		VALUES ($1, $2, $3)
	`
	_, err = tx.Exec(insertQuery, prID, newReviewerID, time.Now())
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Merge переводит PR в статус MERGED (идемпотентно)
func (r *PRRepository) Merge(prID uuid.UUID) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Rollback()
		if err != nil {
			return
		}
	}()

	var currentStatus string
	var mergedAt *time.Time
	statusQuery := `SELECT status, merged_at FROM pull_requests WHERE id = $1`
	err = tx.QueryRow(statusQuery, prID).Scan(&currentStatus, &mergedAt)
	if err != nil {
		return err
	}

	if currentStatus == "MERGED" {
		return tx.Commit()
	}

	now := time.Now()
	updateQuery := `
		UPDATE pull_requests
		SET status = 'MERGED', merged_at = $1
		WHERE id = $2
	`
	_, err = tx.Exec(updateQuery, now, prID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetAll возвращает все PR
func (r *PRRepository) GetAll() ([]model.PullRequest, error) {
	query := `
		SELECT id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		ORDER BY created_at DESC
	`
	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			return
		}
	}()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var prs []model.PullRequest
	for rows.Next() {
		var pr model.PullRequest
		err = rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
		if err != nil {
			return nil, err
		}

		reviewersQuery := `SELECT reviewer_id FROM pr_reviewers WHERE pr_id = $1 ORDER BY assigned_at`
		reviewerRows, errQuery := r.DB.Query(reviewersQuery, pr.ID)
		if errQuery != nil {
			return nil, errQuery
		}
		var reviewers []uuid.UUID
		for reviewerRows.Next() {
			var reviewerID uuid.UUID
			if err = reviewerRows.Scan(&reviewerID); err != nil {
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
		if err := reviewerRows.Err(); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	return prs, nil
}
