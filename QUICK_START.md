# Быстрый старт для разработчиков

Этот документ поможет быстро разобраться в коде проекта и начать разработку.

## Архитектура в двух словах

Проект использует **чистую архитектуру** с тремя основными слоями:

1. **Handlers** - принимают HTTP запросы, валидируют данные
2. **Services** - содержат бизнес-логику
3. **Repositories** - работают с базой данных

```
HTTP Request → Handler → Service → Repository → Database
                ↓         ↓         ↓
              Validate  Business  SQL
                        Logic     Queries
```

## Как работает создание PR

Вот полный путь запроса при создании PR:

### 1. HTTP Handler (`PRHandler.CreatePR`)

```go
// Принимает JSON: {"pull_request_name": "...", "author_id": "..."}
// Валидирует данные
// Вызывает сервис
```

### 2. Service (`PRService.CreatePR`)

```go
// Проверяет существование автора
// Получает команду автора
// Находит активных пользователей команды (исключая автора)
// Выбирает до 2 ревьюверов случайным образом
// Создает PR через репозиторий
```

### 3. Repository (`PRRepository.Create`)

```go
// Начинает транзакцию
// Вставляет PR в таблицу pull_requests
// Вставляет ревьюверов в таблицу pr_reviewers
// Коммитит транзакцию
```

## Основные концепции

### 1. Модели данных

Все модели в `internal/model/`:
- `User` - пользователь
- `Team` - команда
- `PullRequest` - PR
- `ReviewStats` - статистика

### 2. Репозитории

Репозитории в `internal/repository/`:
- Выполняют SQL запросы
- Работают с транзакциями
- Преобразуют данные БД в модели

**Пример:**
```go
func (r *UserRepository) GetUserByID(id uuid.UUID) (*model.User, error) {
    // SQL запрос
    // Преобразование в модель
    // Возврат результата
}
```

### 3. Сервисы

Сервисы в `internal/service/`:
- Содержат бизнес-логику
- Координируют работу репозиториев
- Проверяют бизнес-правила

**Пример:**
```go
func (s *PRService) CreatePR(pr *model.PullRequest) (*model.PullRequest, error) {
    // 1. Проверка автора
    // 2. Получение кандидатов
    // 3. Выбор ревьюверов
    // 4. Создание через репозиторий
}
```

### 4. Хэндлеры

Хэндлеры в `internal/api/handlers/`:
- Принимают HTTP запросы
- Валидируют входные данные
- Вызывают сервисы
- Возвращают HTTP ответы

**Пример:**
```go
func (h *PRHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
    // 1. Парсинг JSON
    // 2. Валидация
    // 3. Вызов сервиса
    // 4. Возврат результата или ошибки
}
```

## Работа с базой данных

### Таблицы

- `teams` - команды
- `users` - пользователи
- `pull_requests` - PR
- `pr_reviewers` - связь PR и ревьюверов (many-to-many)

### Транзакции

Критические операции выполняются в транзакциях:
```go
tx, _ := db.Begin()
// ... операции ...
tx.Commit() // или tx.Rollback() при ошибке
```

## Обработка ошибок

**В хэндлерах:**
- 400 - неверный запрос (невалидный JSON, UUID)
- 404 - ресурс не найден
- 409 - конфликт (дубликат имени)
- 500 - внутренняя ошибка

**В сервисах:**
- Возвращают понятные текстовые ошибки
- Проверяют бизнес-правила

**В репозиториях:**
- Возвращают ошибки БД
- Используют транзакции для атомарности

## Как добавить новый функционал

**Пример: Добавить эндпоинт для получения статистики по команде**

1. **Модель** (если нужна новая структура):
```go
// internal/model/statistics.go
type TeamStats struct {
    TeamID uuid.UUID
    // ...
}
```

2. **Репозиторий** (SQL запросы):
```go
// internal/repository/team_repository.go
func (r *TeamRepository) GetTeamStats(teamID uuid.UUID) (*model.TeamStats, error) {
    // SQL запрос
}
```

3. **Сервис** (бизнес-логика):
```go
// internal/service/team_service.go
func (s *TeamService) GetTeamStats(teamID uuid.UUID) (*model.TeamStats, error) {
    // Вызов репозитория + бизнес-логика
}
```

4. **Хэндлер** (HTTP):
```go
// internal/api/handlers/team_handler.go
func (h *TeamHandler) GetTeamStats(w http.ResponseWriter, r *http.Request) {
    // Валидация, вызов сервиса, возврат результата
}
```

5. **Маршрут** (в `main.go`):
```go
r.HandleFunc("/api/v1/teams/{team_id}/stats", teamHandler.GetTeamStats).Methods("GET")
```

## Полезные команды

```bash
# Запуск сервера
make run

# Тесты
make test

# Линтер
make lint

# Миграции
make migrate-up
make migrate-down
```

## Частые вопросы

**Q: Где находится бизнес-логика?**
A: В `internal/service/`

**Q: Как добавить новую таблицу?**
A: Создать миграцию в `migrations/` и применить через `make migrate-up`

**Q: Как тестировать?**
A: Используйте `internal/repository/test_helpers.go` для создания тестовой БД

**Q: Где валидируются данные?**
A: В хэндлерах (UUID, JSON) и сервисах (бизнес-правила)
