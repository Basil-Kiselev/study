# Переменные
APP_NAME=account
BINARY_PATH=../bin/$(APP_NAME)
DOCKER_COMPOSE=../docker/docker-compose.yml

.PHONY: up down build run test clean

# Запустить базу данных
up:
	docker compose -f $(DOCKER_COMPOSE) up -d

# Остановить базу данных
down:
	docker compose -f $(DOCKER_COMPOSE) down

# Собрать Go приложение
build:
	go build -o $(BINARY_PATH) ./cmd/main.go

# Запустить приложение
run: build
	$(BINARY_PATH)

# Запустить тесты
test:
	go test ./...

# Очистить бинарные файлы
clean:
	rm -f $(BINARY_PATH)
