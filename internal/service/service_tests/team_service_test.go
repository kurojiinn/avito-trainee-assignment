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

func setupTeamServiceTestDB(t *testing.T) (*repository.TeamRepository, *repository.UserRepository) {
	t.Helper()
	db := repository_tests.SetupTestDB(t)
	return repository.NewTeamRepository(db), repository.NewUserRepository(db)
}

func TestTeamService_CreateTeam(t *testing.T) {
	teamRepo, userRepo := setupTeamServiceTestDB(t)
	defer repository_tests.CleanupTestDB(t, teamRepo.DB)

	service := service2.NewTeamService(teamRepo, userRepo)

	t.Run("successful creation", func(t *testing.T) {
		team := &model.Team{
			Name: "Test Team",
		}

		created, err := service.CreateTeam(team)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, created.ID)
		assert.Equal(t, team.Name, created.Name)
	})

	t.Run("duplicate name", func(t *testing.T) {
		team := &model.Team{
			Name: "Test Team",
		}

		_, err := service.CreateTeam(team)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestTeamService_GetTeamByID(t *testing.T) {
	teamRepo, userRepo := setupTeamServiceTestDB(t)
	defer repository_tests.CleanupTestDB(t, teamRepo.DB)

	service := service2.NewTeamService(teamRepo, userRepo)

	// Создаем команду
	team := &model.Team{
		Name: "Test Team",
	}
	created, err := service.CreateTeam(team)
	require.NoError(t, err)

	t.Run("existing team", func(t *testing.T) {
		retrieved, err := service.GetTeamByID(created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, retrieved.ID)
		assert.Equal(t, created.Name, retrieved.Name)
	})

	t.Run("non-existent team", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := service.GetTeamByID(nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "team not found")
	})
}

func TestTeamService_UpdateTeam(t *testing.T) {
	teamRepo, userRepo := setupTeamServiceTestDB(t)
	defer repository_tests.CleanupTestDB(t, teamRepo.DB)

	service := service2.NewTeamService(teamRepo, userRepo)

	// Создаем команду
	team := &model.Team{
		Name: "Test Team",
	}
	created, err := service.CreateTeam(team)
	require.NoError(t, err)

	t.Run("update team", func(t *testing.T) {
		created.Name = "Updated Team"
		updated, err := service.UpdateTeam(created)
		require.NoError(t, err)
		assert.Equal(t, "Updated Team", updated.Name)
	})
}

func TestTeamService_DeleteTeam(t *testing.T) {
	teamRepo, userRepo := setupTeamServiceTestDB(t)
	defer repository_tests.CleanupTestDB(t, teamRepo.DB)

	service := service2.NewTeamService(teamRepo, userRepo)

	// Создаем команду
	team := &model.Team{
		Name: "Test Team",
	}
	created, err := service.CreateTeam(team)
	require.NoError(t, err)

	t.Run("delete team", func(t *testing.T) {
		err := service.DeleteTeam(created.ID)
		assert.NoError(t, err)

		// Проверяем, что команда удалена
		_, err = service.GetTeamByID(created.ID)
		assert.Error(t, err)
	})
}
