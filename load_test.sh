#!/bin/bash

# Скрипт для нагрузочного тестирования API
# Требует установленного Apache Bench (ab) или wrk

BASE_URL="${BASE_URL:-http://localhost:8080}"
CONCURRENT_USERS="${CONCURRENT_USERS:-5}"
TOTAL_REQUESTS="${TOTAL_REQUESTS:-100}"

echo "========================================="
echo "Нагрузочное тестирование API"
echo "========================================="
echo "Base URL: $BASE_URL"
echo "Concurrent users: $CONCURRENT_USERS"
echo "Total requests: $TOTAL_REQUESTS"
echo ""

# Проверяем доступность сервиса
echo "Проверка доступности сервиса..."
if ! curl -s -f "$BASE_URL/health" > /dev/null; then
    echo "ОШИБКА: Сервис недоступен по адресу $BASE_URL"
    exit 1
fi
echo "✓ Сервис доступен"
echo ""

# Создаем тестовые данные
echo "Создание тестовых данных..."
TEAM_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/teams" \
    -H "Content-Type: application/json" \
    -d '{"team_name": "Load Test Team"}')
TEAM_ID=$(echo $TEAM_RESPONSE | grep -o '"team_id":"[^"]*"' | cut -d'"' -f4)
echo "✓ Создана команда: $TEAM_ID"

AUTHOR_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/users" \
    -H "Content-Type: application/json" \
    -d "{\"username\": \"load_author\", \"team_id\": \"$TEAM_ID\", \"is_active\": true}")
AUTHOR_ID=$(echo $AUTHOR_RESPONSE | grep -o '"user_id":"[^"]*"' | cut -d'"' -f4)
echo "✓ Создан автор: $AUTHOR_ID"
echo ""

# Тест 1: Health Check
echo "========================================="
echo "Тест 1: Health Check"
echo "========================================="
if command -v ab &> /dev/null; then
    ab -n $TOTAL_REQUESTS -c $CONCURRENT_USERS "$BASE_URL/health" | grep -E "(Requests per second|Time per request|Failed requests)"
elif command -v wrk &> /dev/null; then
    wrk -t$CONCURRENT_USERS -c$CONCURRENT_USERS -d10s "$BASE_URL/health"
else
    echo "Установите Apache Bench (ab) или wrk для нагрузочного тестирования"
fi
echo ""

# Тест 2: Получение статистики
echo "========================================="
echo "Тест 2: Получение статистики"
echo "========================================="
if command -v ab &> /dev/null; then
    ab -n $TOTAL_REQUESTS -c $CONCURRENT_USERS "$BASE_URL/api/v1/statistics" | grep -E "(Requests per second|Time per request|Failed requests)"
elif command -v wrk &> /dev/null; then
    wrk -t$CONCURRENT_USERS -c$CONCURRENT_USERS -d10s "$BASE_URL/api/v1/statistics"
else
    echo "Установите Apache Bench (ab) или wrk для нагрузочного тестирования"
fi
echo ""

# Тест 3: Создание PR
echo "========================================="
echo "Тест 3: Создание PR"
echo "========================================="
PR_DATA="{\"pull_request_name\": \"Load Test PR\", \"author_id\": \"$AUTHOR_ID\"}"
if command -v ab &> /dev/null; then
    echo "$PR_DATA" | ab -n $TOTAL_REQUESTS -c $CONCURRENT_USERS -p - -T application/json "$BASE_URL/api/v1/pull-requests" | grep -E "(Requests per second|Time per request|Failed requests)"
elif command -v wrk &> /dev/null; then
    echo "$PR_DATA" > /tmp/pr_data.json
    wrk -t$CONCURRENT_USERS -c$CONCURRENT_USERS -d10s -s /tmp/pr_data.json "$BASE_URL/api/v1/pull-requests"
    rm /tmp/pr_data.json
else
    echo "Установите Apache Bench (ab) или wrk для нагрузочного тестирования"
fi
echo ""

echo "========================================="
echo "Нагрузочное тестирование завершено"
echo "========================================="

