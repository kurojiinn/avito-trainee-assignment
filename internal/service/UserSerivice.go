package service

import (
	"avito-assignment/internal/model"
	"avito-assignment/internal/repository"
	"errors"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// Создать пользователя
func (s *UserService) CreateUser(userC *model.User) (*model.User, error) {

	user := &model.User{
		ID:       uuid.New(),
		Username: userC.Username,
		TeamID:   userC.TeamID,
		IsActive: userC.IsActive,
	}
	err := s.userRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Получить пользователя по ID
func (s *UserService) GetUserByID(id uuid.UUID) (*model.User, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserService) UpdateUser(user *model.User) (*model.User, error) {
	return s.userRepo.Update(user)
}

// Удаление пользователя
func (s *UserService) DeleteUser(id uuid.UUID) error {
	return s.userRepo.Delete(id)
}

// Получение PR'ов, где пользователь назначен ревьювером
func (s *UserService) GetAssignedPRs(userID uuid.UUID) ([]model.PullRequest, error) {
	return s.userRepo.GetPRsByReviewer(userID)
}
