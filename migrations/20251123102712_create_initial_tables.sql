-- +goose Up

-- Подключаем расширение для генерации UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Тип для статуса PR
CREATE TYPE pr_status AS ENUM ('OPEN','MERGED');

-- Таблица teams
CREATE TABLE teams (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       name TEXT NOT NULL UNIQUE,
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Таблица users
CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       username TEXT NOT NULL UNIQUE,
                       team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
                       is_active BOOLEAN NOT NULL DEFAULT TRUE,
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

-- Таблица pull_requests
CREATE TABLE pull_requests (
                               id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                               pull_request_name TEXT NOT NULL,
                               author_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
                               status pr_status NOT NULL DEFAULT 'OPEN',
                               created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
                               merged_at TIMESTAMP WITH TIME ZONE NULL
);

-- Таблица pr_reviewers (many-to-many)
CREATE TABLE pr_reviewers (
                              pr_id UUID NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
                              reviewer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
                              assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
                              PRIMARY KEY (pr_id, reviewer_id)
);

-- Индексы для быстрых запросов
CREATE INDEX idx_users_team_id ON users(team_id);
CREATE INDEX idx_pr_author ON pull_requests(author_id);
CREATE INDEX idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);
CREATE INDEX idx_pr_reviewers_pr ON pr_reviewers(pr_id);

-- +goose Down

-- В откате удаляем таблицы и индексы в обратном порядке
DROP INDEX IF EXISTS idx_pr_reviewers_pr;
DROP INDEX IF EXISTS idx_pr_reviewers_reviewer;
DROP INDEX IF EXISTS idx_pr_author;
DROP INDEX IF EXISTS idx_users_team_id;

DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

-- Удаляем ENUM pr_status
DROP TYPE IF EXISTS pr_status;

-- Расширение uuid-ossp оставляем на месте
