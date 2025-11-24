// Package service содержит бизнес-логику приложения.
// Сервисы координируют работу репозиториев и реализуют бизнес-правила.
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
// Отвечает за создание PR, назначение ревьюверов, переназначение и объединение.
type PRService struct {
	prRepo   *repository.PRRepository   // Репозиторий для работы с PR
	userRepo *repository.UserRepository // Репозиторий для работы с пользователями
	teamRepo *repository.TeamRepository // Репозиторий для работы с командами
}

// NewPRService создает новый экземпляр PRService с заданными репозиториями.
func NewPRService(prRepo *repository.PRRepository, userRepo *repository.UserRepository, teamRepo *repository.TeamRepository) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}

// CreatePR создает новый Pull Request и автоматически назначает ревьюверов.
// Бизнес-правила:
//   - Автор должен существовать в системе
//   - Ревьюверы выбираются из команды автора
//   - Автор исключается из списка кандидатов
//   - Назначается до 2 активных ревьюверов
//   - Если доступных кандидатов меньше 2, назначается доступное количество (0 или 1)
//
// Параметры:
//   - pr: Pull Request для создания (должен содержать Title и AuthorID)
//
// Возвращает:
//   - Созданный PR с назначенными ревьюверами или ошибку
func (s *PRService) CreatePR(pr *model.PullRequest) (*model.PullRequest, error) {
	// Проверяем существование автора в системе
	author, err := s.userRepo.GetUserByID(pr.AuthorID)
	if err != nil {
		return nil, errors.New("Автор/команда не найдены")
	}

	// Получаем активных пользователей команды автора, исключая самого автора
	// Только активные пользователи (IsActive = true) могут быть назначены ревьюверами
	candidates, err := s.userRepo.GetActiveUsersByTeam(author.TeamID, pr.AuthorID)
	if err != nil {
		return nil, err
	}

	// Выбираем до 2 ревьюверов случайным образом из доступных кандидатов
	reviewers := s.selectRandomReviewers(candidates, 2)

	// Инициализируем поля PR
	pr.ID = uuid.New()        // Генерируем уникальный идентификатор
	pr.Status = model.OPEN    // Устанавливаем статус "открыт"
	pr.CreatedAt = time.Now() // Записываем время создания
	pr.Reviewers = reviewers  // Назначаем выбранных ревьюверов

	// Сохраняем PR в базу данных вместе с назначенными ревьюверами
	err = s.prRepo.Create(pr, reviewers)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

// GetPRByID возвращает Pull Request по его идентификатору.
// Включает список назначенных ревьюверов.
//
// Параметры:
//   - id: UUID Pull Request
//
// Возвращает:
//   - PR с полной информацией или ошибку, если PR не найден
func (s *PRService) GetPRByID(id uuid.UUID) (*model.PullRequest, error) {
	pr, err := s.prRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("pull request not found")
	}
	return pr, nil
}

// ReassignReviewer переназначает одного ревьювера на другого.
// Бизнес-правила:
//   - PR должен существовать и иметь статус OPEN
//   - Старый ревьювер должен быть назначен на этот PR
//   - Новый ревьювер выбирается из команды старого ревьювера
//   - Автор PR и старый ревьювер исключаются из кандидатов
//   - После объединения PR (MERGED) переназначение запрещено
//
// Параметры:
//   - prID: UUID Pull Request
//   - oldReviewerID: UUID ревьювера, которого нужно заменить
//
// Возвращает:
//   - Обновленный PR с новым ревьювером или ошибку
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
// Операция идемпотентна: повторный вызов не приводит к ошибке и возвращает актуальное состояние.
//
// Бизнес-правила:
//   - PR должен существовать
//   - Если PR уже MERGED, операция просто возвращает текущее состояние
//   - После объединения изменение ревьюверов запрещено
//
// Параметры:
//   - prID: UUID Pull Request для объединения
//
// Возвращает:
//   - PR со статусом MERGED и временем объединения или ошибку
func (s *PRService) MergePR(prID uuid.UUID) (*model.PullRequest, error) {
	// Проверяем существование PR
	_, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, errors.New("pull request not found")
	}

	// Выполняем merge (операция идемпотентна - можно вызывать многократно)
	// Если PR уже MERGED, метод просто вернет текущее состояние без изменений
	err = s.prRepo.Merge(prID)
	if err != nil {
		return nil, err
	}

	// Возвращаем обновленный PR с актуальным статусом
	return s.prRepo.GetByID(prID)
}

// GetAllPRs возвращает все Pull Requests из системы.
// Список отсортирован по времени создания (новые первыми).
//
// Возвращает:
//   - Список всех PR с назначенными ревьюверами или ошибку
func (s *PRService) GetAllPRs() ([]model.PullRequest, error) {
	return s.prRepo.GetAll()
}

// selectRandomReviewers выбирает случайных ревьюверов из списка кандидатов.
// Используется алгоритм Fisher-Yates для равномерного распределения.
//
// Параметры:
//   - candidates: список кандидатов на роль ревьювера
//   - maxCount: максимальное количество ревьюверов для выбора (обычно 2)
//
// Возвращает:
//   - Список UUID выбранных ревьюверов (от 0 до maxCount)
func (s *PRService) selectRandomReviewers(candidates []model.User, maxCount int) []uuid.UUID {
	// Если кандидатов нет, возвращаем пустой список
	if len(candidates) == 0 {
		return []uuid.UUID{}
	}

	// Определяем количество ревьюверов для выбора
	// Не может быть больше доступных кандидатов
	count := maxCount
	if len(candidates) < maxCount {
		count = len(candidates)
	}

	// Создаем копию списка кандидатов для перемешивания
	// Это необходимо, чтобы не изменять исходный список
	shuffled := make([]model.User, len(candidates))
	copy(shuffled, candidates)

	// Перемешиваем список используя алгоритм Fisher-Yates
	// Это обеспечивает равномерное распределение вероятностей
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// Выбираем первых count ревьюверов из перемешанного списка
	reviewers := make([]uuid.UUID, 0, count)
	for i := 0; i < count; i++ {
		reviewers = append(reviewers, shuffled[i].ID)
	}

	return reviewers
}
