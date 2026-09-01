include .env
export

MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONY: test
test:
	@go test -v ./...

.PHONY: migrate-create
migrate-create:
	@migrate create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" up

.PHONY: migrate-down
migrate-down:
	@migrate -path=$(MIGRATIONS_PATH) -database="$(DB_URL)" down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: seed
seed:
	@go run cmd/migrate/seed/main.go

.PHONY: truncate
truncate:
	@psql "$(DB_URL)" -c "TRUNCATE TABLE comments, posts, users RESTART IDENTITY CASCADE;"

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt

.PHONY: watch
watch:
	@air

.PHONY: docker-up
docker-up:
	@docker compose up --build

.PHONY: docker-down
docker-down:
	@docker compose down

.PHONY: test-concurrency
test-concurrency:
	@go run scripts/test_concurrency.go

# Catch-all: lets you pass extra args to migrate-create / migrate-down
# (e.g. `make migrate-create create_users`) without Make complaining
# "No rule to make target 'create_users'".
%:
	@: