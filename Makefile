# Testing commands
test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

test-unit:
	go test ./internal/... -v

test-integration:
	go test ./... -tags=integration -v

# Mock generation
install-mockgen:
	go install go.uber.org/mock/mockgen@latest

generate-repo-mocks:
	mockgen -source=internal/domain/interfaces/repository.go -destination=mocks/mock_repository.go -package=mocks

generate-controller-mocks:
	mockgen -source=internal/domain/interfaces/controller.go -destination=mocks/mock_controller.go -package=mocks

generate-mocks: generate-repo-mocks generate-controller-mocks

# Clean generated files
clean-mocks:
	rm -rf mocks/

# Regenerate all mocks
regen-mocks: clean-mocks generate-mocks

# Development with hot reload
dev:
	docker-compose up --build

# Production build and run
up:
	docker-compose up --build -d

down:
	docker-compose down

down-v:
	docker-compose down -v

# Clean everything and restart
restart: down-v dev

# Run locally (requires local postgres)
run:
	go run ./cmd/api

# Install air for local development
install-air:
	go install github.com/air-verse/air@latest

# Run with air locally
air:
	air -c .air.toml

gensql:
	cd sqlboiler_config && sqlboiler psql -o ../internal/infrastructure/database/models

# Migration commands
migrate-up:
	migrate -path db/migrations -database "postgres://postgres:password@localhost:5432/assignment?sslmode=disable" up

migrate-down:
	migrate -path db/migrations -database "postgres://postgres:password@localhost:5432/assignment?sslmode=disable" down

migrate-create:
	migrate create -ext sql -dir db/migrations -seq $(name)

migrate-force:
	migrate -path db/migrations -database "postgres://postgres:password@localhost:5432/assignment?sslmode=disable" force $(version)