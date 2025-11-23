package service_tests

import (
	"avito-assignment/internal/model"
	"avito-assignment/internal/repository"
	"avito-assignment/internal/repository/repository_tests"
	"avito-assignment/internal/service"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPRServiceTestDB(t *testing.T) (*repository.PRRepository, *repository.UserRepository, *repository.TeamRepository) {
	t.Helper()
	db := repository_tests.SetupTestDB(t)
	return repository.NewPRRepository(db), repository.NewUserRepository(db), repository.NewTeamRepository(db)
}

func TestPRService_CreatePR(t *testing.T) {
	prRepo, userRepo, teamRepo := setupPRServiceTestDB(t)
	defer repository_tests.CleanupTestDB(t, prRepo.DB)

	userService := service.NewUserService(userRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)

	// Создаем команду
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем автора
	author := &model.User{
		Username: "author",
		TeamID:   team.ID,
		IsActive: true,
	}
	createdAuthor, err := userService.CreateUser(author)
	require.NoError(t, err)

	// Создаем активных пользователей в команде
	reviewer1 := &model.User{
		Username: "reviewer1",
		TeamID:   team.ID,
		IsActive: true,
	}
	_, err = userService.CreateUser(reviewer1)
	require.NoError(t, err)

	reviewer2 := &model.User{
		Username: "reviewer2",
		TeamID:   team.ID,
		IsActive: true,
	}
	_, err = userService.CreateUser(reviewer2)
	require.NoError(t, err)

	t.Run("create PR with automatic reviewers", func(t *testing.T) {
		pr := &model.PullRequest{
			Title:    "Test PR",
			AuthorID: createdAuthor.ID,
		}

		created, err := prService.CreatePR(pr)
		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, created.ID)
		assert.Equal(t, model.OPEN, created.Status)
		assert.LessOrEqual(t, len(created.Reviewers), 2)
		assert.Greater(t, len(created.Reviewers), 0)
		// Автор не должен быть в списке ревьюверов
		for _, reviewerID := range created.Reviewers {
			assert.NotEqual(t, createdAuthor.ID, reviewerID)
		}
	})

	t.Run("create PR with no available reviewers", func(t *testing.T) {
		// Создаем автора в отдельной команде без других участников
		emptyTeam := &model.Team{
			ID:   uuid.New(),
			Name: "Empty Team",
		}
		err := teamRepo.Create(emptyTeam)
		require.NoError(t, err)

		lonelyAuthor := &model.User{
			Username: "lonely",
			TeamID:   emptyTeam.ID,
			IsActive: true,
		}
		createdLonely, err := userService.CreateUser(lonelyAuthor)
		require.NoError(t, err)

		pr := &model.PullRequest{
			Title:    "Lonely PR",
			AuthorID: createdLonely.ID,
		}

		created, err := prService.CreatePR(pr)
		require.NoError(t, err)
		assert.Len(t, created.Reviewers, 0)
	})

	t.Run("create PR with author not found", func(t *testing.T) {
		nonExistentAuthorID := uuid.New()
		pr := &model.PullRequest{
			Title:    "Test PR",
			AuthorID: nonExistentAuthorID,
		}

		_, err := prService.CreatePR(pr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "author not found")
	})
}

func TestPRService_ReassignReviewer(t *testing.T) {
	prRepo, userRepo, teamRepo := setupPRServiceTestDB(t)
	defer repository_tests.CleanupTestDB(t, prRepo.DB)

	userService := service.NewUserService(userRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)

	// Создаем команду
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем автора
	author := &model.User{
		Username: "author",
		TeamID:   team.ID,
		IsActive: true,
	}
	createdAuthor, err := userService.CreateUser(author)
	require.NoError(t, err)

	// Создаем ревьюверов (нужно минимум 3, чтобы можно было переназначить)
	oldReviewer := &model.User{
		Username: "old_reviewer",
		TeamID:   team.ID,
		IsActive: true,
	}
	createdOldReviewer, err := userService.CreateUser(oldReviewer)
	require.NoError(t, err)

	newReviewer := &model.User{
		Username: "new_reviewer",
		TeamID:   team.ID,
		IsActive: true,
	}
	_, err = userService.CreateUser(newReviewer)
	require.NoError(t, err)

	// Создаем еще одного ревьювера для гарантии, что будет доступен кандидат
	thirdReviewer := &model.User{
		Username: "third_reviewer",
		TeamID:   team.ID,
		IsActive: true,
	}
	_, err = userService.CreateUser(thirdReviewer)
	require.NoError(t, err)

	t.Run("reassign reviewer", func(t *testing.T) {
		// Создаем отдельный PR с известным ревьювером для теста переназначения
		// Это гарантирует, что у нас есть контролируемая ситуация
		testPR := &model.PullRequest{
			ID:        uuid.New(),
			Title:     "Test Reassign PR",
			AuthorID:  createdAuthor.ID,
			Status:    model.OPEN,
			CreatedAt: time.Now(),
		}
		// Назначаем только одного ревьювера для простоты теста
		err = prRepo.Create(testPR, []uuid.UUID{createdOldReviewer.ID})
		require.NoError(t, err)

		// Переназначаем ревьювера
		updated, err := prService.ReassignReviewer(testPR.ID, createdOldReviewer.ID)
		require.NoError(t, err)

		// Проверяем, что старый ревьювер удален
		assert.NotContains(t, updated.Reviewers, createdOldReviewer.ID, "Старый ревьювер должен быть удален")

		// Проверяем, что новый ревьювер назначен
		assert.Greater(t, len(updated.Reviewers), 0, "Должен быть хотя бы один ревьювер")

		// Проверяем, что новый ревьювер не совпадает со старым
		for _, reviewerID := range updated.Reviewers {
			assert.NotEqual(t, createdOldReviewer.ID, reviewerID, "Новый ревьювер не должен совпадать со старым")
		}

		// Проверяем, что ревьюверы не дублируются
		reviewerMap := make(map[uuid.UUID]bool)
		for _, reviewerID := range updated.Reviewers {
			assert.False(t, reviewerMap[reviewerID], "Ревьюверы не должны дублироваться")
			reviewerMap[reviewerID] = true
		}
	})

	t.Run("cannot reassign merged PR", func(t *testing.T) {
		// Создаем новый PR и мержим его
		pr := &model.PullRequest{
			Title:    "Merged PR",
			AuthorID: createdAuthor.ID,
		}
		mergedPR, err := prService.CreatePR(pr)
		require.NoError(t, err)

		_, err = prService.MergePR(mergedPR.ID)
		require.NoError(t, err)

		// Пытаемся переназначить ревьювера
		if len(mergedPR.Reviewers) > 0 {
			_, err = prService.ReassignReviewer(mergedPR.ID, mergedPR.Reviewers[0])
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "merged PR")
		}
	})
}

func TestPRService_MergePR(t *testing.T) {
	prRepo, userRepo, teamRepo := setupPRServiceTestDB(t)
	defer repository_tests.CleanupTestDB(t, prRepo.DB)

	userService := service.NewUserService(userRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)

	// Создаем команду
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем автора
	author := &model.User{
		Username: "author",
		TeamID:   team.ID,
		IsActive: true,
	}
	createdAuthor, err := userService.CreateUser(author)
	require.NoError(t, err)

	// Создаем ревьювера
	reviewer := &model.User{
		Username: "reviewer",
		TeamID:   team.ID,
		IsActive: true,
	}
	_, err = userService.CreateUser(reviewer)
	require.NoError(t, err)

	// Создаем PR
	pr := &model.PullRequest{
		Title:    "Test PR",
		AuthorID: createdAuthor.ID,
	}
	createdPR, err := prService.CreatePR(pr)
	require.NoError(t, err)

	t.Run("merge PR", func(t *testing.T) {
		merged, err := prService.MergePR(createdPR.ID)
		require.NoError(t, err)
		assert.Equal(t, model.MERGED, merged.Status)
		assert.NotNil(t, merged.MergedAt)
	})

	t.Run("merge is idempotent", func(t *testing.T) {
		// Пытаемся мержить уже мерженый PR
		merged, err := prService.MergePR(createdPR.ID)
		require.NoError(t, err)
		assert.Equal(t, model.MERGED, merged.Status)
	})
}

func TestPRService_GetPRByID(t *testing.T) {
	prRepo, userRepo, teamRepo := setupPRServiceTestDB(t)
	defer repository_tests.CleanupTestDB(t, prRepo.DB)

	userService := service.NewUserService(userRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)

	// Создаем команду
	team := &model.Team{
		ID:   uuid.New(),
		Name: "Test Team",
	}
	err := teamRepo.Create(team)
	require.NoError(t, err)

	// Создаем автора
	author := &model.User{
		Username: "author",
		TeamID:   team.ID,
		IsActive: true,
	}
	createdAuthor, err := userService.CreateUser(author)
	require.NoError(t, err)

	// Создаем ревьювера
	reviewer := &model.User{
		Username: "reviewer",
		TeamID:   team.ID,
		IsActive: true,
	}
	_, err = userService.CreateUser(reviewer)
	require.NoError(t, err)

	// Создаем PR
	pr := &model.PullRequest{
		Title:    "Test PR",
		AuthorID: createdAuthor.ID,
	}
	createdPR, err := prService.CreatePR(pr)
	require.NoError(t, err)

	t.Run("get existing PR", func(t *testing.T) {
		retrieved, err := prService.GetPRByID(createdPR.ID)
		require.NoError(t, err)
		assert.Equal(t, createdPR.ID, retrieved.ID)
		assert.Equal(t, createdPR.Title, retrieved.Title)
	})

	t.Run("get non-existent PR", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := prService.GetPRByID(nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pull request not found")
	})
}
