# Сервис назначения ревьюеров для Pull Request'ов

Микросервис для автоматического назначения ревьюеров на Pull Request'ы (PR), а также управления командами и участниками.

## Описание

Сервис предоставляет HTTP API для:
- Управления пользователями (создание, обновление, удаление, получение)
- Управления командами (создание, обновление, удаление, получение)
- Создания PR с автоматическим назначением ревьюеров (до 2 активных ревьюверов из команды автора)
- Переназначения ревьюверов (из команды заменяемого ревьювера)
- Объединения PR (идемпотентная операция)
- Получения списка PR, назначенных конкретному пользователю

## Технологический стек

- **Язык**: Go 1.21
- **База данных**: PostgreSQL 15
- **HTTP роутер**: Gorilla Mux
- **Миграции**: Goose
- **Контейнеризация**: Docker & Docker Compose

## Быстрый старт

### Требования

- Docker и Docker Compose
- Go 1.21+ (для локальной разработки)

### Запуск с Docker Compose

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd avito-trainee-assignment
```

2. Запустите сервис:
```bash
docker-compose up
```

Сервис будет доступен на `http://localhost:8080`

Миграции применяются автоматически при запуске через отдельный сервис `migrate`.



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
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Content-Type: application/json" \
  -d '
  {
  "team_name": "Transactional",
  "members": [
    {
      "name": "Ivan Petrov",
      "email": "ivan@example.com",
      "role": "developer"
    },
    {
      "name": "Anna Smirnova",
      "email": "anna@example.com",
      "role": "tester"
    }
  ]
}'
```

### Создание пользователя
```bash
curl -X POST http://localhost:8080/api/v1/users \
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



## Дополнительные задания

### ✅ Эндпоинт статистики

Реализован эндпоинт `/api/v1/statistics`, который возвращает:
- Общее количество назначений ревьюверов
- Статистику назначений по пользователям (топ-20)
- Статистику назначений по PR (топ-20)
- Общую статистику по PR (всего, открытых, мерженых)
- Среднее количество ревьюверов на PR

**Пример запроса:**
```bash
curl http://localhost:8080/api/v1/statistics
```

### ✅ Массовая деактивация пользователей команды

Реализован эндпоинт `POST /api/v1/teams/{team_id}/deactivate-members`, который:
- Массово деактивирует всех активных пользователей команды
- Безопасно переназначает ревьюверов для открытых PR
- Возвращает количество деактивированных пользователей

**Пример запроса:**
```bash
curl -X POST http://localhost:8080/api/v1/teams/{team_id}/deactivate-members
```

**Оптимизация:** Операция оптимизирована для укладывания в 100 мс для средних объёмов данных:
- Используются эффективные SQL-запросы с JOIN
- Минимизировано количество обращений к БД
- Используются транзакции для атомарности


### ✅ Конфигурация линтера

Настроен `golangci-lint` с конфигурацией в `.golangci.yml`:
- Включены основные линтеры: errcheck, govet, golint, gosec, staticcheck
- Настроены правила для тестовых файлов
- Оптимизированы настройки для проекта

**Установка и запуск:**
```bash
# Установка golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Запуск линтера
golangci-lint run
```

