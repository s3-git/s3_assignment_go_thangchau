run:
	cd cmd/api && go run .

test:
	go test ./... -v

up:
	docker-compose up --build -d

rebuild-api:
	docker-compose up --build -d --force-recreate api

down:
	docker-compose down

down-v:
	docker-compose down -v

gensql:
	cd sqlboiler_config && sqlboiler psql -o ../internal/infrastructure/database/models