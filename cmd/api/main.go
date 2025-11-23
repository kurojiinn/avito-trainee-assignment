// Package main - точка входа приложения.
// Инициализирует все компоненты системы и запускает HTTP сервер.
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
	_ "github.com/lib/pq" // Драйвер PostgreSQL
)

// main инициализирует и запускает HTTP сервер.
// Последовательность инициализации:
// 1. Загрузка конфигурации из переменных окружения
// 2. Подключение к базе данных PostgreSQL
// 3. Создание репозиториев для работы с данными
// 4. Создание сервисов с бизнес-логикой
// 5. Создание HTTP обработчиков
// 6. Настройка маршрутизации
// 7. Запуск HTTP сервера на порту 8080 (или из переменной окружения PORT)
func main() {
	// 1. Загружаем конфигурацию из переменных окружения
	// Используются значения по умолчанию, если переменные не установлены
	cfg := config.LoadConfig()

	// 2. Подключаемся к базе данных PostgreSQL
	// Строка подключения формируется из конфигурации
	dbConn := db.Connect(&cfg.DB)
	// Закрываем соединение при завершении программы
	defer dbConn.Close()

	// 3. Инициализируем репозитории (слой доступа к данным)
	// Репозитории отвечают за выполнение SQL запросов к БД
	userRepo := repository.NewUserRepository(dbConn)
	teamRepo := repository.NewTeamRepository(dbConn)
	prRepo := repository.NewPRRepository(dbConn)
	statsRepo := repository.NewStatisticsRepository(dbConn)

	// 4. Инициализируем сервисы (слой бизнес-логики)
	// Сервисы содержат бизнес-правила и координируют работу репозиториев
	userService := service.NewUserService(userRepo)
	teamService := service.NewTeamService(teamRepo, userRepo)
	prService := service.NewPRService(prRepo, userRepo, teamRepo)
	statsService := service.NewStatisticsService(statsRepo)

	// 5. Инициализируем HTTP обработчики (слой представления)
	// Обработчики принимают HTTP запросы, валидируют данные и вызывают сервисы
	userHandler := &handlers.UserHandler{Service: userService}
	teamHandler := &handlers.TeamHandler{Service: teamService, PRService: prService}
	prHandler := &handlers.PRHandler{Service: prService}
	statsHandler := &handlers.StatisticsHandler{Service: statsService}

	// 6. Настраиваем маршрутизацию HTTP запросов
	// Используется библиотека gorilla/mux для гибкой маршрутизации
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
	r.HandleFunc("/api/v1/teams", teamHandler.CreateTeam).Methods("POST")
	r.HandleFunc("/api/v1/teams/{team_id}", teamHandler.GetTeam).Methods("GET")
	r.HandleFunc("/api/v1/teams/{team_id}", teamHandler.UpdateTeam).Methods("PUT")
	r.HandleFunc("/api/v1/teams/{team_id}", teamHandler.DeleteTeam).Methods("DELETE")
	r.HandleFunc("/api/v1/teams/{team_id}/deactivate-members", teamHandler.DeactivateTeamMembers).Methods("POST")

	// PR endpoints - управление Pull Requests
	r.HandleFunc("/api/v1/pull-requests", prHandler.CreatePR).Methods("POST")
	r.HandleFunc("/api/v1/pull-requests", prHandler.GetAllPRs).Methods("GET")
	r.HandleFunc("/api/v1/pull-requests/{pull_request_id}", prHandler.GetPR).Methods("GET")
	r.HandleFunc("/api/v1/pull-requests/{pull_request_id}/reassign", prHandler.ReassignReviewer).Methods("POST")
	r.HandleFunc("/api/v1/pull-requests/{pull_request_id}/merge", prHandler.MergePR).Methods("POST")

	// Statistics endpoint - статистика по назначениям
	r.HandleFunc("/api/v1/statistics", statsHandler.GetStatistics).Methods("GET")

	// 7. Запускаем HTTP сервер
	// Порт по умолчанию: 8080
	// Можно переопределить через переменную окружения PORT
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	log.Printf("Starting server at %s", addr)
	// Запускаем сервер и обрабатываем ошибки
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
