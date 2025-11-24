package service

import (
	"avito-assignment/internal/model"
	"avito-assignment/internal/repository"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// PRService реализует бизнес-логику для работы с Pull Requests.
type PRService struct {
	prRepo   *repository.PRRepository
	userRepo *repository.UserRepository
	teamRepo *repository.TeamRepository
}

func NewPRService(prRepo *repository.PRRepository, userRepo *repository.UserRepository, teamRepo *repository.TeamRepository) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

// CreatePR создает новый Pull Request и автоматически назначает ревьюверов.
func (s *PRService) CreatePR(pr *model.PullRequest) (*model.PullRequest, error) {
	author, err := s.userRepo.GetUserByID(pr.AuthorID)
	if err != nil {
		return nil, errors.New("Автор/команда не найдены")
	}

	candidates, err := s.userRepo.GetActiveUsersByTeam(author.TeamID, pr.AuthorID)
	if err != nil {
		return nil, err
	}

	reviewers := s.selectRandomReviewers(candidates, 2)

	pr.ID = uuid.New()
	pr.Status = model.OPEN
	pr.CreatedAt = time.Now()
	pr.Reviewers = reviewers

	err = s.prRepo.Create(pr, reviewers)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// GetPRByID возвращает Pull Request по его идентификатору.
func (s *PRService) GetPRByID(id uuid.UUID) (*model.PullRequest, error) {
	pr, err := s.prRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("pull request not found")
	}
	return pr, nil
}

// ReassignReviewer переназначает одного ревьювера на другого.
func (s *PRService) ReassignReviewer(
	prID uuid.UUID,
	oldReviewerID uuid.UUID,
) (*model.PullRequest, uuid.UUID, error) {
	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, uuid.Nil, errors.New("pull request not found")
	}

	if pr.Status == model.MERGED {
		return nil, uuid.Nil, errors.New("cannot reassign reviewers for merged PR")
	}

	found := false
	for _, reviewerID := range pr.Reviewers {
		if reviewerID == oldReviewerID {
			found = true
			break
		}
	}
	if !found {
		return nil, uuid.Nil, errors.New("reviewer not assigned to this PR")
	}

	oldReviewer, err := s.userRepo.GetUserByID(oldReviewerID)
	if err != nil {
		return nil, uuid.Nil, errors.New("old reviewer not found")
	}

	excludeIDs := []uuid.UUID{pr.AuthorID, oldReviewerID}
	for _, reviewerID := range pr.Reviewers {
		if reviewerID != oldReviewerID {
			excludeIDs = append(excludeIDs, reviewerID)
		}
	}

	candidates, err := s.userRepo.GetActiveUsersByTeamExcluding(
		oldReviewer.TeamID,
		excludeIDs,
	)
	if err != nil {
		return nil, uuid.Nil, err
	}
	if len(candidates) == 0 {
		return nil, uuid.Nil, errors.New("no available reviewers in the team")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	newReviewerID := candidates[rng.Intn(len(candidates))].ID

	err = s.prRepo.ReassignReviewer(prID, oldReviewerID, newReviewerID)
	if err != nil {
		return nil, uuid.Nil, err
	}

	updated, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, uuid.Nil, err
	}

	return updated, newReviewerID, nil
}

// MergePR переводит Pull Request в статус MERGED.
func (s *PRService) MergePR(prID uuid.UUID) (*model.PullRequest, error) {
	_, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, errors.New("pull request not found")
	}

	err = s.prRepo.Merge(prID)
	if err != nil {
		return nil, err
	}

	return s.prRepo.GetByID(prID)
}

// GetAllPRs возвращает все Pull Requests из системы.
func (s *PRService) GetAllPRs() ([]model.PullRequest, error) {
	return s.prRepo.GetAll()
}

// selectRandomReviewers выбирает случайных ревьюверов из списка кандидатов.
func (s *PRService) selectRandomReviewers(candidates []model.User, maxCount int) []uuid.UUID {
	if len(candidates) == 0 {
		return []uuid.UUID{}
	}

	count := maxCount
	if len(candidates) < maxCount {
		count = len(candidates)
	}

	shuffled := make([]model.User, len(candidates))
	copy(shuffled, candidates)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	reviewers := make([]uuid.UUID, 0, count)
	for i := 0; i < count; i++ {
		reviewers = append(reviewers, shuffled[i].ID)
	}

	return reviewers
}
