package main

import (
	"avito-assignment/internal/api/handlers"
	"avito-assignment/internal/config"
	"avito-assignment/internal/db"
	"avito-assignment/internal/repository"
	"avito-assignment/internal/service"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	// Загрузка конфигураций из переменных окружения
	cfg := config.LoadConfig()

	dbConn := db.Connect(&cfg.DB)
	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Инициализация репозиториев (слой доступа к данным)
	userRepo := repository.NewUserRepository(dbConn)
	teamRepo := repository.NewTeamRepository(dbConn)
	prRepo := repository.NewPRRepository(dbConn)
	statsRepo := repository.NewStatisticsRepository(dbConn)

	// Инициализация сервисов
	userService := service.NewUserService(userRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)
	statsService := service.NewStatisticsService(statsRepo)

	// Инициализация HTTP обработчиков
	userHandler := &handlers.UserHandler{Service: userService}
	teamHandler := &handlers.TeamHandler{Service: teamService, PRService: prService}
	prHandler := &handlers.PRHandler{Service: prService}
	statsHandler := &handlers.StatisticsHandler{Service: statsService}

	r := mux.NewRouter()

	// Health check endpoint - проверка работоспособности сервиса
	r.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// User endpoints - управление пользователями
	r.HandleFunc("/api/v1/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/api/v1/users/{user_id}", userHandler.GetUser).Methods("GET")
	r.HandleFunc("/api/v1/users/{user_id}", userHandler.UpdateUser).Methods("PUT")
	r.HandleFunc("/api/v1/users/{user_id}", userHandler.DeleteUser).Methods("DELETE")
	r.HandleFunc("/api/v1/users/{user_id}/pull-requests", userHandler.GetUserPRs).Methods("GET")

	// Team endpoints - управление командами
	r.HandleFunc("/api/v1/team/add", teamHandler.CreateTeam).Methods("POST")
	r.HandleFunc("/api/v1/team/{team_id}", teamHandler.GetTeam).Methods("GET")
	r.HandleFunc("/api/v1/team/{team_id}", teamHandler.UpdateTeam).Methods("PUT")
	r.HandleFunc("/api/v1/team/{team_id}", teamHandler.DeleteTeam).Methods("DELETE")
	r.HandleFunc("/api/v1/team/{team_id}/deactivate-members", teamHandler.DeactivateTeamMembers).Methods("POST")

	// PR endpoints - управление Pull Requests
	r.HandleFunc("/api/v1/pull-request/create", prHandler.CreatePR).Methods("POST")
	r.HandleFunc("/api/v1/pull-request", prHandler.GetAllPRs).Methods("GET")
	r.HandleFunc("/api/v1/pull-request/{pull_request_id}", prHandler.GetPR).Methods("GET")
	r.HandleFunc("/api/v1/pull-request/reassign", prHandler.ReassignReviewer).Methods("POST")
	r.HandleFunc("/api/v1/pull-request/merge", prHandler.MergePR).Methods("POST")

	// Statistics endpoint - статистика по назначениям
	r.HandleFunc("/api/v1/statistics", statsHandler.GetStatistics).Methods("GET")

	// Запуск HTTP сервера
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	log.Printf("Starting server at %s", addr)
	// Запуск сервера
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
