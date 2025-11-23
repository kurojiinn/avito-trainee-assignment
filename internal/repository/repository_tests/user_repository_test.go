package repository_tests

import (
	"avito-assignment/internal/model"
	"avito-assignment/internal/repository"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_CreateUser(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := repository.NewUserRepository(db)

	// Создаем команду для теста
	teamRepo := repository.NewTeamRepository(db)
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	t.Run("successful creation", func(t *testing.T) {
		user := &model.User{
			ID:       uuid.New(),
			Username: "testuser",
			TeamID:   team.ID,
			IsActive: true,
		}

		err := repo.CreateUser(user)
		assert.NoError(t, err)

		// Проверяем, что пользователь создан
		retrieved, err := repo.GetUserByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Username, retrieved.Username)
		assert.Equal(t, user.TeamID, retrieved.TeamID)
		assert.Equal(t, user.IsActive, retrieved.IsActive)
	})

	t.Run("duplicate username", func(t *testing.T) {
		user := &model.User{
			ID:       uuid.New(),
			Username: "testuser", // Дубликат
			TeamID:   team.ID,
			IsActive: true,
		}

		err := repo.CreateUser(user)
		assert.Error(t, err)
	})
}

func TestUserRepository_GetUserByID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := repository.NewUserRepository(db)

	// Создаем команду
	teamRepo := repository.NewTeamRepository(db)
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем пользователя
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = repo.CreateUser(user)
	require.NoError(t, err)

	t.Run("existing user", func(t *testing.T) {
		retrieved, err := repo.GetUserByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Username, retrieved.Username)
	})

	t.Run("non-existent user", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := repo.GetUserByID(nonExistentID)
		assert.Error(t, err)
	})
}

func TestUserRepository_Update(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := repository.NewUserRepository(db)

	// Создаем команду
	teamRepo := repository.NewTeamRepository(db)
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем пользователя
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = repo.CreateUser(user)
	require.NoError(t, err)

	t.Run("update user", func(t *testing.T) {
		user.Username = "updateduser"
		user.IsActive = false

		updated, err := repo.Update(user)
		require.NoError(t, err)
		assert.Equal(t, "updateduser", updated.Username)
		assert.False(t, updated.IsActive)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := repository.NewUserRepository(db)

	// Создаем команду
	teamRepo := repository.NewTeamRepository(db)
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем пользователя
	user := &model.User{
		ID:       uuid.New(),
		Username: "testuser",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = repo.CreateUser(user)
	require.NoError(t, err)

	t.Run("delete user", func(t *testing.T) {
		err := repo.Delete(user.ID)
		assert.NoError(t, err)

		// Проверяем, что пользователь удален
		_, err = repo.GetUserByID(user.ID)
		assert.Error(t, err)
	})
}

func TestUserRepository_GetActiveUsersByTeam(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	repo := repository.NewUserRepository(db)

	// Создаем команду
	teamRepo := repository.NewTeamRepository(db)
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем активных пользователей
	activeUser1 := &model.User{
		ID:       uuid.New(),
		Username: "active1",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = repo.CreateUser(activeUser1)
	require.NoError(t, err)

	activeUser2 := &model.User{
		ID:       uuid.New(),
		Username: "active2",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = repo.CreateUser(activeUser2)
	require.NoError(t, err)

	// Создаем неактивного пользователя
	inactiveUser := &model.User{
		ID:       uuid.New(),
		Username: "inactive",
		TeamID:   team.ID,
		IsActive: false,
	}
	err = repo.CreateUser(inactiveUser)
	require.NoError(t, err)

	t.Run("get active users excluding one", func(t *testing.T) {
		users, err := repo.GetActiveUsersByTeam(team.ID, activeUser1.ID)
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, activeUser2.ID, users[0].ID)
	})

	t.Run("get active users excluding all", func(t *testing.T) {
		users, err := repo.GetActiveUsersByTeamExcluding(team.ID, []uuid.UUID{activeUser1.ID, activeUser2.ID})
		require.NoError(t, err)
		assert.Len(t, users, 0)
	})
}

func TestUserRepository_GetPRsByReviewer(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	userRepo := repository.NewUserRepository(db)
	prRepo := repository.NewPRRepository(db)

	// Создаем команду
	teamRepo := repository.NewTeamRepository(db)
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем автора
	author := &model.User{
		ID:       uuid.New(),
		Username: "author",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = userRepo.CreateUser(author)
	require.NoError(t, err)

	// Создаем ревьювера
	reviewer := &model.User{
		ID:       uuid.New(),
		Username: "reviewer",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = userRepo.CreateUser(reviewer)
	require.NoError(t, err)

	// Создаем PR
	pr := &model.PullRequest{
		ID:        uuid.New(),
		Title:     "Test PR",
		AuthorID:  author.ID,
		Status:    model.OPEN,
		Reviewers: []uuid.UUID{reviewer.ID},
	}
	err = prRepo.Create(pr, []uuid.UUID{reviewer.ID})
	require.NoError(t, err)

	t.Run("get PRs by reviewer", func(t *testing.T) {
		prs, err := userRepo.GetPRsByReviewer(reviewer.ID)
		require.NoError(t, err)
		assert.Len(t, prs, 1)
		assert.Equal(t, pr.ID, prs[0].ID)
		assert.Contains(t, prs[0].Reviewers, reviewer.ID)
	})
}
