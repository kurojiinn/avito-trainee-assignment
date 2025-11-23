package repository_tests

import (
	"avito-assignment/internal/model"
	"avito-assignment/internal/repository"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPRRepository_Create(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	prRepo := repository.NewPRRepository(db)
	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)

	// Создаем команду
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

	// Создаем ревьюверов
	reviewer1 := &model.User{
		ID:       uuid.New(),
		Username: "reviewer1",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = userRepo.CreateUser(reviewer1)
	require.NoError(t, err)

	reviewer2 := &model.User{
		ID:       uuid.New(),
		Username: "reviewer2",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = userRepo.CreateUser(reviewer2)
	require.NoError(t, err)

	t.Run("create PR with reviewers", func(t *testing.T) {
		pr := &model.PullRequest{
			ID:        uuid.New(),
			Title:     "Test PR",
			AuthorID:  author.ID,
			Status:    model.OPEN,
			CreatedAt: time.Now(),
		}

		err := prRepo.Create(pr, []uuid.UUID{reviewer1.ID, reviewer2.ID})
		assert.NoError(t, err)

		// Проверяем, что PR создан
		retrieved, err := prRepo.GetByID(pr.ID)
		require.NoError(t, err)
		assert.Equal(t, pr.Title, retrieved.Title)
		assert.Len(t, retrieved.Reviewers, 2)
		assert.Contains(t, retrieved.Reviewers, reviewer1.ID)
		assert.Contains(t, retrieved.Reviewers, reviewer2.ID)
	})
}

func TestPRRepository_GetByID(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	prRepo := repository.NewPRRepository(db)
	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)

	// Создаем команду
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
		CreatedAt: time.Now(),
	}
	err = prRepo.Create(pr, []uuid.UUID{reviewer.ID})
	require.NoError(t, err)

	t.Run("get existing PR", func(t *testing.T) {
		retrieved, err := prRepo.GetByID(pr.ID)
		require.NoError(t, err)
		assert.Equal(t, pr.ID, retrieved.ID)
		assert.Equal(t, pr.Title, retrieved.Title)
		assert.Len(t, retrieved.Reviewers, 1)
	})

	t.Run("get non-existent PR", func(t *testing.T) {
		nonExistentID := uuid.New()
		_, err := prRepo.GetByID(nonExistentID)
		assert.Error(t, err)
	})
}

func TestPRRepository_ReassignReviewer(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	prRepo := repository.NewPRRepository(db)
	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)

	// Создаем команду
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

	// Создаем ревьюверов
	oldReviewer := &model.User{
		ID:       uuid.New(),
		Username: "old_reviewer",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = userRepo.CreateUser(oldReviewer)
	require.NoError(t, err)

	newReviewer := &model.User{
		ID:       uuid.New(),
		Username: "new_reviewer",
		TeamID:   team.ID,
		IsActive: true,
	}
	err = userRepo.CreateUser(newReviewer)
	require.NoError(t, err)

	// Создаем PR
	pr := &model.PullRequest{
		ID:        uuid.New(),
		Title:     "Test PR",
		AuthorID:  author.ID,
		Status:    model.OPEN,
		CreatedAt: time.Now(),
	}
	err = prRepo.Create(pr, []uuid.UUID{oldReviewer.ID})
	require.NoError(t, err)

	t.Run("reassign reviewer", func(t *testing.T) {
		err := prRepo.ReassignReviewer(pr.ID, oldReviewer.ID, newReviewer.ID)
		assert.NoError(t, err)

		// Проверяем переназначение
		retrieved, err := prRepo.GetByID(pr.ID)
		require.NoError(t, err)
		assert.NotContains(t, retrieved.Reviewers, oldReviewer.ID)
		assert.Contains(t, retrieved.Reviewers, newReviewer.ID)
	})

	t.Run("cannot reassign merged PR", func(t *testing.T) {
		// Создаем новый PR и мержим его
		mergedPR := &model.PullRequest{
			ID:        uuid.New(),
			Title:     "Merged PR",
			AuthorID:  author.ID,
			Status:    model.OPEN,
			CreatedAt: time.Now(),
		}
		err = prRepo.Create(mergedPR, []uuid.UUID{oldReviewer.ID})
		require.NoError(t, err)

		err = prRepo.Merge(mergedPR.ID)
		require.NoError(t, err)

		// Пытаемся переназначить ревьювера в мерженом PR
		err = prRepo.ReassignReviewer(mergedPR.ID, oldReviewer.ID, newReviewer.ID)
		assert.Error(t, err)
	})
}

func TestPRRepository_Merge(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	prRepo := repository.NewPRRepository(db)
	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)

	// Создаем команду
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
		CreatedAt: time.Now(),
	}
	err = prRepo.Create(pr, []uuid.UUID{reviewer.ID})
	require.NoError(t, err)

	t.Run("merge PR", func(t *testing.T) {
		err := prRepo.Merge(pr.ID)
		assert.NoError(t, err)

		// Проверяем статус
		retrieved, err := prRepo.GetByID(pr.ID)
		require.NoError(t, err)
		assert.Equal(t, model.MERGED, retrieved.Status)
		assert.NotNil(t, retrieved.MergedAt)
	})

	t.Run("merge is idempotent", func(t *testing.T) {
		// Пытаемся мержить уже мерженый PR
		err := prRepo.Merge(pr.ID)
		assert.NoError(t, err)

		// Проверяем, что статус не изменился
		retrieved, err := prRepo.GetByID(pr.ID)
		require.NoError(t, err)
		assert.Equal(t, model.MERGED, retrieved.Status)
	})
}

// TODO: исправить этот тест
func TestPRRepository_GetAll(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	prRepo := repository.NewPRRepository(db)
	userRepo := repository.NewUserRepository(db)
	teamRepo := repository.NewTeamRepository(db)

	// Создаем команду
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

	// Создаем несколько PR
	pr1 := &model.PullRequest{
		ID:        uuid.New(),
		Title:     "PR 1",
		AuthorID:  author.ID,
		Status:    model.OPEN,
		CreatedAt: time.Now(),
	}
	err = prRepo.Create(pr1, []uuid.UUID{reviewer.ID})
	require.NoError(t, err)

	pr2 := &model.PullRequest{
		ID:        uuid.New(),
		Title:     "PR 2",
		AuthorID:  author.ID,
		Status:    model.OPEN,
		CreatedAt: time.Now(),
	}
	err = prRepo.Create(pr2, []uuid.UUID{reviewer.ID})
	require.NoError(t, err)

	t.Run("get all PRs", func(t *testing.T) {
		prs, err := prRepo.GetAll()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(prs), 2)
	})
}
