package handler_tests

import (
	"avito-assignment/internal/api/handlers"
	"avito-assignment/internal/model"
	"avito-assignment/internal/repository"
	"avito-assignment/internal/repository/repository_tests"
	"avito-assignment/internal/service"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandlers(t *testing.T) (*handlers.UserHandler, *handlers.TeamHandler, *handlers.PRHandler, *repository.UserRepository) {
	t.Helper()

	testDB := repository_tests.SetupTestDB(t)
	userRepo := repository.NewUserRepository(testDB)
	teamRepo := repository.NewTeamRepository(testDB)
	prRepo := repository.NewPRRepository(testDB)

	userService := service.NewUserService(userRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)

	userHandler := &handlers.UserHandler{Service: userService}
	teamHandler := &handlers.TeamHandler{Service: teamService}
	prHandler := &handlers.PRHandler{Service: prService}

	return userHandler, teamHandler, prHandler, userRepo
}

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handlers.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}

func TestUserHandler_CreateUser(t *testing.T) {
	userHandler, teamHandler, _, userRepo := setupTestHandlers(t)
	defer repository_tests.CleanupTestDB(t, userRepo.DB)

	// Создаем команду
	team := &model.Team{
		Name: "Test Team",
	}
	teamBody, _ := json.Marshal(team)
	req := httptest.NewRequest("POST", "/api/v1/teams", bytes.NewBuffer(teamBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	teamHandler.CreateTeam(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var createdTeam model.Team
	json.Unmarshal(w.Body.Bytes(), &createdTeam)

	t.Run("successful creation", func(t *testing.T) {
		user := &model.User{
			Username: "testuser",
			TeamID:   createdTeam.ID,
			IsActive: true,
		}
		userBody, _ := json.Marshal(user)

		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		userHandler.CreateUser(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var createdUser model.User
		json.Unmarshal(w.Body.Bytes(), &createdUser)
		assert.Equal(t, user.Username, createdUser.Username)
		assert.NotEqual(t, uuid.Nil, createdUser.ID)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/users", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		userHandler.CreateUser(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestUserHandler_GetUser(t *testing.T) {
	userHandler, teamHandler, _, userRepo := setupTestHandlers(t)
	defer repository_tests.CleanupTestDB(t, userRepo.DB)

	// Создаем команду
	team := &model.Team{
		Name: "Test Team",
	}
	teamBody, _ := json.Marshal(team)
	req := httptest.NewRequest("POST", "/api/v1/teams", bytes.NewBuffer(teamBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	teamHandler.CreateTeam(w, req)
	var createdTeam model.Team
	json.Unmarshal(w.Body.Bytes(), &createdTeam)

	// Создаем пользователя
	user := &model.User{
		Username: "testuser",
		TeamID:   createdTeam.ID,
		IsActive: true,
	}
	userBody, _ := json.Marshal(user)
	req = httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	userHandler.CreateUser(w, req)
	var createdUser model.User
	json.Unmarshal(w.Body.Bytes(), &createdUser)

	t.Run("existing user", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/users/"+createdUser.ID.String(), nil)
		req = mux.SetURLVars(req, map[string]string{"user_id": createdUser.ID.String()})
		w := httptest.NewRecorder()

		userHandler.GetUser(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var retrievedUser model.User
		json.Unmarshal(w.Body.Bytes(), &retrievedUser)
		assert.Equal(t, createdUser.ID, retrievedUser.ID)
	})

	t.Run("invalid UUID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/users/invalid", nil)
		req = mux.SetURLVars(req, map[string]string{"user_id": "invalid"})
		w := httptest.NewRecorder()

		userHandler.GetUser(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("non-existent user", func(t *testing.T) {
		nonExistentID := uuid.New()
		req := httptest.NewRequest("GET", "/api/v1/users/"+nonExistentID.String(), nil)
		req = mux.SetURLVars(req, map[string]string{"user_id": nonExistentID.String()})
		w := httptest.NewRecorder()

		userHandler.GetUser(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestTeamHandler_CreateTeam(t *testing.T) {
	_, teamHandler, _, userRepo := setupTestHandlers(t)
	defer repository_tests.CleanupTestDB(t, userRepo.DB)

	t.Run("successful creation", func(t *testing.T) {
		team := &model.Team{
			Name: "Test Team",
		}
		teamBody, _ := json.Marshal(team)

		req := httptest.NewRequest("POST", "/api/v1/teams", bytes.NewBuffer(teamBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		teamHandler.CreateTeam(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var createdTeam model.Team
		json.Unmarshal(w.Body.Bytes(), &createdTeam)
		assert.Equal(t, team.Name, createdTeam.Name)
		assert.NotEqual(t, uuid.Nil, createdTeam.ID)
	})

	t.Run("duplicate name", func(t *testing.T) {
		team := &model.Team{
			Name: "Test Team",
		}
		teamBody, _ := json.Marshal(team)

		req := httptest.NewRequest("POST", "/api/v1/teams", bytes.NewBuffer(teamBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		teamHandler.CreateTeam(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestPRHandler_CreatePR(t *testing.T) {
	userHandler, teamHandler, prHandler, userRepo := setupTestHandlers(t)
	defer repository_tests.CleanupTestDB(t, userRepo.DB)

	// Создаем команду
	team := &model.Team{
		Name: "Test Team",
	}
	teamBody, _ := json.Marshal(team)
	req := httptest.NewRequest("POST", "/api/v1/teams", bytes.NewBuffer(teamBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	teamHandler.CreateTeam(w, req)
	var createdTeam model.Team
	json.Unmarshal(w.Body.Bytes(), &createdTeam)

	// Создаем автора
	author := &model.User{
		Username: "author",
		TeamID:   createdTeam.ID,
		IsActive: true,
	}
	userBody, _ := json.Marshal(author)
	req = httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	userHandler.CreateUser(w, req)
	var createdAuthor model.User
	json.Unmarshal(w.Body.Bytes(), &createdAuthor)

	// Создаем ревьюверов
	reviewer1 := &model.User{
		Username: "reviewer1",
		TeamID:   createdTeam.ID,
		IsActive: true,
	}
	userBody, _ = json.Marshal(reviewer1)
	req = httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	userHandler.CreateUser(w, req)

	reviewer2 := &model.User{
		Username: "reviewer2",
		TeamID:   createdTeam.ID,
		IsActive: true,
	}
	userBody, _ = json.Marshal(reviewer2)
	req = httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	userHandler.CreateUser(w, req)

	t.Run("successful creation", func(t *testing.T) {
		prReq := handlers.CreatePRRequest{
			Title:    "Test PR",
			AuthorID: createdAuthor.ID,
		}
		prBody, _ := json.Marshal(prReq)

		req := httptest.NewRequest("POST", "/api/v1/pull-requests", bytes.NewBuffer(prBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		prHandler.CreatePR(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var createdPR model.PullRequest
		json.Unmarshal(w.Body.Bytes(), &createdPR)
		assert.Equal(t, prReq.Title, createdPR.Title)
		assert.LessOrEqual(t, len(createdPR.Reviewers), 2)
	})

	t.Run("author not found", func(t *testing.T) {
		nonExistentAuthorID := uuid.New()
		prReq := handlers.CreatePRRequest{
			Title:    "Test PR",
			AuthorID: nonExistentAuthorID,
		}
		prBody, _ := json.Marshal(prReq)

		req := httptest.NewRequest("POST", "/api/v1/pull-requests", bytes.NewBuffer(prBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		prHandler.CreatePR(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestPRHandler_MergePR(t *testing.T) {
	userHandler, teamHandler, prHandler, userRepo := setupTestHandlers(t)
	defer repository_tests.CleanupTestDB(t, userRepo.DB)

	// Создаем команду
	team := &model.Team{
		Name: "Test Team",
	}
	teamBody, _ := json.Marshal(team)
	req := httptest.NewRequest("POST", "/api/v1/teams", bytes.NewBuffer(teamBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	teamHandler.CreateTeam(w, req)
	var createdTeam model.Team
	json.Unmarshal(w.Body.Bytes(), &createdTeam)

	// Создаем автора
	author := &model.User{
		Username: "author",
		TeamID:   createdTeam.ID,
		IsActive: true,
	}
	userBody, _ := json.Marshal(author)
	req = httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	userHandler.CreateUser(w, req)
	var createdAuthor model.User
	json.Unmarshal(w.Body.Bytes(), &createdAuthor)

	// Создаем ревьювера
	reviewer := &model.User{
		Username: "reviewer",
		TeamID:   createdTeam.ID,
		IsActive: true,
	}
	userBody, _ = json.Marshal(reviewer)
	req = httptest.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	userHandler.CreateUser(w, req)

	// Создаем PR
	prReq := handlers.CreatePRRequest{
		Title:    "Test PR",
		AuthorID: createdAuthor.ID,
	}
	prBody, _ := json.Marshal(prReq)
	req = httptest.NewRequest("POST", "/api/v1/pull-requests", bytes.NewBuffer(prBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	prHandler.CreatePR(w, req)
	var createdPR model.PullRequest
	json.Unmarshal(w.Body.Bytes(), &createdPR)

	t.Run("merge PR", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/pull-requests/"+createdPR.ID.String()+"/merge", nil)
		req = mux.SetURLVars(req, map[string]string{"pull_request_id": createdPR.ID.String()})
		w := httptest.NewRecorder()

		prHandler.MergePR(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var mergedPR model.PullRequest
		json.Unmarshal(w.Body.Bytes(), &mergedPR)
		assert.Equal(t, model.MERGED, mergedPR.Status)
		assert.NotNil(t, mergedPR.MergedAt)
	})

	t.Run("merge is idempotent", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/pull-requests/"+createdPR.ID.String()+"/merge", nil)
		req = mux.SetURLVars(req, map[string]string{"pull_request_id": createdPR.ID.String()})
		w := httptest.NewRecorder()

		prHandler.MergePR(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
