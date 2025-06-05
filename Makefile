run:
	cd cmd/api && go run .

test:
	go test ./... -v

up:
	docker-compose up --build -d

down:
	docker-compose down

down-v:
	docker-compose down -v

gensql:
	cd sqlboiler_config && sqlboiler psql -o ../internal/infrastructure/database/models