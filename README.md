
## API Endpoints

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

## Примеры использования

### Создание команды
```bash
  POST http://localhost:8080/api/v1/team/add \
  -H "Content-Type: application/json" \
  -d '
  {
  "team_name": "transactional",
  "members": [
    {
      "username": "Ivan Petrov",
      "is_active": true
    },
  ]
}'
```

### Создание пользователя
```bash
  POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "team_id": "<team_id>",
    "is_active": true
  }'
```
## Makefile команды

- `make build` - Собрать приложение
- `make run` - Запустить приложение локально
- `make docker-build` - Собрать Docker образ
- `make docker-up` - Запустить сервисы через Docker Compose
- `make docker-down` - Остановить сервисы
- `make docker-logs` - Показать логи API сервиса
- `make migrate-up` - Применить миграции
- `make migrate-down` - Откатить последнюю миграцию
- `make migrate-status` - Показать статус миграций
- `make migrate-create name=<name>` - Создать новую миграцию
- `make install-deps` - Установить зависимости
- `make fmt` - Форматировать код
- `make vet` - Проверить код с помощью go vet
- `make clean` - Удалить артефакты сборки


## Архитектура

Проект следует принципам чистой архитектуры:

```
cmd/api/              # Точка входа приложения
internal/
  api/handlers/       # HTTP обработчики
  config/             # Конфигурация
  db/                 # Подключение к БД
  model/              # Модели данных
  repository/         # Слой доступа к данным
  service/            # Бизнес-логика
migrations/           # SQL миграции
```


## Вопросы и решения

### Выбор библиотеки для миграций
Использована библиотека `goose` для миграций, так как она проста в использовании и хорошо интегрируется с PostgreSQL.

Добавлен linter(golangci lint)