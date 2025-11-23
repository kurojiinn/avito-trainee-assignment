package service_tests

import (
	"avito-assignment/internal/model"
	"avito-assignment/internal/repository"
	"avito-assignment/internal/repository/repository_tests"
	service2 "avito-assignment/internal/service"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupServiceTestDB(t *testing.T) *repository.UserRepository {
	t.Helper()
	db := repository_tests.SetupTestDB(t)
	return repository.NewUserRepository(db)
}

func TestUserService_CreateUser(t *testing.T) {
	userRepo := setupServiceTestDB(t)
	defer repository_tests.CleanupTestDB(t, userRepo.DB)

	service := service2.NewUserService(userRepo)

	// Создаем команду для теста
	teamRepo := repository.NewTeamRepository(userRepo.DB)
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		user := &model.User{
			Username: "testuser",
			TeamID:   team.ID,
			IsActive: true,
		}

		created, err := service.CreateUser(user)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, created.ID)
		assert.Equal(t, user.Username, created.Username)
		assert.Equal(t, user.TeamID, created.TeamID)
		assert.Equal(t, user.IsActive, created.IsActive)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	userRepo := setupServiceTestDB(t)
	defer repository_tests.CleanupTestDB(t, userRepo.DB)

	service := service2.NewUserService(userRepo)

	// Создаем команду
	teamRepo := repository.NewTeamRepository(userRepo.DB)
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем пользователя
	user := &model.User{
		Username: "testuser",
		TeamID:   team.ID,
		IsActive: true,
	}
	created, err := service.CreateUser(user)
	require.NoError(t, err)

	t.Run("existing user", func(t *testing.T) {
		retrieved, err := service.GetUserByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, retrieved.ID)
		assert.Equal(t, created.Username, retrieved.Username)
	})

	t.Run("non-existent user", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := service.GetUserByID(nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	userRepo := setupServiceTestDB(t)
	defer repository_tests.CleanupTestDB(t, userRepo.DB)

	service := service2.NewUserService(userRepo)

	// Создаем команду
	teamRepo := repository.NewTeamRepository(userRepo.DB)
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем пользователя
	user := &model.User{
		Username: "testuser",
		TeamID:   team.ID,
		IsActive: true,
	}
	created, err := service.CreateUser(user)
	require.NoError(t, err)

	t.Run("update user", func(t *testing.T) {
		created.Username = "updateduser"
		created.IsActive = false

		updated, err := service.UpdateUser(created)
		require.NoError(t, err)
		assert.Equal(t, "updateduser", updated.Username)
		assert.False(t, updated.IsActive)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	userRepo := setupServiceTestDB(t)
	defer repository_tests.CleanupTestDB(t, userRepo.DB)

	service := service2.NewUserService(userRepo)

	// Создаем команду
	teamRepo := repository.NewTeamRepository(userRepo.DB)
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем пользователя
	user := &model.User{
		Username: "testuser",
		TeamID:   team.ID,
		IsActive: true,
	}
	created, err := service.CreateUser(user)
	require.NoError(t, err)

	t.Run("delete user", func(t *testing.T) {
		err := service.DeleteUser(created.ID)
		assert.NoError(t, err)

		// Проверяем, что пользователь удален
		_, err = service.GetUserByID(created.ID)
		assert.Error(t, err)
	})
}

//// TODO:  исправить этот тест
//func TestUserService_GetAssignedPRs(t *testing.T) {
//	userRepo := setupServiceTestDB(t)
//	defer repository.CleanupTestDB(t, userRepo.DB)
//
//	service := NewUserService(userRepo)
//	prRepo := repository.NewPRRepository(userRepo.DB)
//	teamRepo := repository.NewTeamRepository(userRepo.DB)
//
//	// Создаем команду
//	team := &model.Team{
//		ID:   uuid.New(),
//		Name: "Test Team",
//	}
//	err := teamRepo.Create(team)
//	require.NoError(t, err)
//
//	// Создаем автора
//	author := &model.User{
//		Username: "author",
//		TeamID:   team.ID,
//		IsActive: true,
//	}
//	createdAuthor, err := service.CreateUser(author)
//	require.NoError(t, err)
//
//	// Создаем ревьювера
//	reviewer := &model.User{
//		Username: "reviewer",
//		TeamID:   team.ID,
//		IsActive: true,
//	}
//	createdReviewer, err := service.CreateUser(reviewer)
//	require.NoError(t, err)
//
//	// Создаем PR напрямую через репозиторий
//	pr := &model.PullRequest{
//		ID:       uuid.New(),
//		Title:    "Test PR",
//		AuthorID: createdAuthor.ID,
//		Status:   model.OPEN,
//	}
//	err = prRepo.Create(pr, []uuid.UUID{createdReviewer.ID})
//	require.NoError(t, err)
//
//	t.Run("get assigned PRs", func(t *testing.T) {
//		prs, err := service.GetAssignedPRs(createdReviewer.ID)
//		require.NoError(t, err)
//		assert.Len(t, prs, 1)
//		assert.Equal(t, pr.ID, prs[0].ID)
//	})
//}
