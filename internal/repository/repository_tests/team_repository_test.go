package repository_tests

import (
	"avito-assignment/internal/model"
	"avito-assignment/internal/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeamRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := repository.NewTeamRepository(db)

	t.Run("successful creation", func(t *testing.T) {
		team := &model.Team{
			ID:   uuid.New(),
			Name: "Test Team",
		}

		err := repo.Create(team)
		assert.NoError(t, err)

		// Проверяем, что команда создана
		retrieved, err := repo.GetByID(team.ID)
		require.NoError(t, err)
		assert.Equal(t, team.Name, retrieved.Name)
	})

	t.Run("duplicate name", func(t *testing.T) {
		team := &model.Team{
			ID:   uuid.New(),
			Name: "Test Team", // Дубликат
		}

		err := repo.Create(team)
		assert.Error(t, err)
	})
}

func TestTeamRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := repository.NewTeamRepository(db)

	// Создаем команду
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := repo.Create(team)
	require.NoError(t, err)

	t.Run("existing team", func(t *testing.T) {
		retrieved, err := repo.GetByID(team.ID)
		require.NoError(t, err)
		assert.Equal(t, team.ID, retrieved.ID)
		assert.Equal(t, team.Name, retrieved.Name)
	})

	t.Run("non-existent team", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := repo.GetByID(nonExistentID)
		assert.Error(t, err)
	})
}

func TestTeamRepository_GetByName(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := repository.NewTeamRepository(db)

	// Создаем команду
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := repo.Create(team)
	require.NoError(t, err)

	t.Run("existing team by name", func(t *testing.T) {
		retrieved, err := repo.GetByName("Test Team")
		require.NoError(t, err)
		assert.Equal(t, team.ID, retrieved.ID)
		assert.Equal(t, team.Name, retrieved.Name)
	})

	t.Run("non-existent team by name", func(t *testing.T) {
		_, err := repo.GetByName("Non-existent Team")
		assert.Error(t, err)
	})
}

func TestTeamRepository_GetMembers(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	teamRepo := repository.NewTeamRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Создаем команду
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем пользователей в команде
	user1 := &model.User{
		ID:       uuid.New(),
		Username: "user1",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = userRepo.CreateUser(user1)
	require.NoError(t, err)

	user2 := &model.User{
		ID:       uuid.New(),
		Username: "user2",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = userRepo.CreateUser(user2)
	require.NoError(t, err)

	t.Run("get team members", func(t *testing.T) {
		members, err := teamRepo.GetMembers(team.ID)
		require.NoError(t, err)
		assert.Len(t, members, 2)
	})
}

func TestTeamRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := repository.NewTeamRepository(db)

	// Создаем команду
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := repo.Create(team)
	require.NoError(t, err)

	t.Run("update team", func(t *testing.T) {
		team.Name = "Updated Team"
		err := repo.Update(team)
		assert.NoError(t, err)

		// Проверяем обновление
		retrieved, err := repo.GetByID(team.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Team", retrieved.Name)
	})
}

func TestTeamRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := repository.NewTeamRepository(db)

	// Создаем команду
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := repo.Create(team)
	require.NoError(t, err)

	t.Run("delete team", func(t *testing.T) {
		err := repo.Delete(team.ID)
		assert.NoError(t, err)

		// Проверяем, что команда удалена
		_, err = repo.GetByID(team.ID)
		assert.Error(t, err)
	})
}
