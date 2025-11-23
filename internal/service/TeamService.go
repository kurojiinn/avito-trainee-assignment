package service

import (
	"avito-assignment/internal/model"
	"avito-assignment/internal/repository"
	"errors"

	"github.com/google/uuid"
)

type TeamService struct {
	teamRepo *repository.TeamRepository
	userRepo *repository.UserRepository
}

func NewTeamService(teamRepo *repository.TeamRepository, userRepo *repository.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

// CreateTeam создает команду
func (s *TeamService) CreateTeam(team *model.Team) (*model.Team, error) {
	// Проверяем уникальность имени
	existing, _ := s.teamRepo.GetByName(team.Name)
	if existing != nil {
		return nil, errors.New("team with this name already exists")
	}

	team.ID = uuid.New()
	err := s.teamRepo.Create(team)
	if err != nil {
		return nil, err
	}
	return team, nil
}

// GetTeamByID возвращает команду по ID
func (s *TeamService) GetTeamByID(id uuid.UUID) (*model.Team, error) {
	team, err := s.teamRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("team not found")
	}

	// Загружаем участников
	members, err := s.teamRepo.GetMembers(id)
	if err == nil {
		team.Members = members
	}

	return team, nil
}

// UpdateTeam обновляет команду
func (s *TeamService) UpdateTeam(team *model.Team) (*model.Team, error) {
	// Проверяем существование
	_, err := s.teamRepo.GetByID(team.ID)
	if err != nil {
		return nil, errors.New("team not found")
	}

	// Проверяем уникальность имени (если изменилось)
	existing, _ := s.teamRepo.GetByName(team.Name)
	if existing != nil && existing.ID != team.ID {
		return nil, errors.New("team with this name already exists")
	}

	err = s.teamRepo.Update(team)
	if err != nil {
		return nil, err
	}
	return team, nil
}

// DeleteTeam удаляет команду
func (s *TeamService) DeleteTeam(id uuid.UUID) error {
	return s.teamRepo.Delete(id)
}

//ДОП задание

// DeactivateTeamMembers массово деактивирует всех пользователей команды
// и безопасно переназначает ревьюверов для открытых PR
func (s *TeamService) DeactivateTeamMembers(teamID uuid.UUID, prService *PRService) (int, error) {
	// Проверяем существование команды
	_, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		return 0, errors.New("team not found")
	}

	// Получаем всех активных пользователей команды
	activeUsers, err := s.userRepo.GetActiveUsersByTeam(teamID, uuid.Nil)
	if err != nil {
		return 0, err
	}

	if len(activeUsers) == 0 {
		return 0, nil
	}

	// Получаем все открытые PR
	allPRs, err := prService.GetAllPRs()
	if err != nil {
		return 0, err
	}

	// Находим открытые PR с ревьюверами из деактивируемой команды
	userIDMap := make(map[uuid.UUID]bool)
	for _, user := range activeUsers {
		userIDMap[user.ID] = true
	}

	// Переназначаем ревьюверов для открытых PR
	for _, pr := range allPRs {
		if pr.Status != model.OPEN {
			continue
		}

		// Находим ревьюверов из деактивируемой команды
		for _, reviewerID := range pr.Reviewers {
			if userIDMap[reviewerID] {
				// Используем существующий метод переназначения
				// Он автоматически найдет замену из команды ревьювера
				_, err := prService.ReassignReviewer(pr.ID, reviewerID)
				if err != nil {
					// Если не удалось переназначить (нет доступных ревьюверов),
					// просто пропускаем - ревьювер будет деактивирован
					continue
				}
			}
		}
	}

	// Деактивируем пользователей
	deactivatedCount, err := s.userRepo.DeactivateTeamMembers(teamID)
	if err != nil {
		return 0, err
	}

	return deactivatedCount, nil
}
