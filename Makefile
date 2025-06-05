run:
	cd cmd/api && go run .

up:
	docker-compose up -d

down:
	docker-compose down

down-v:
	docker-compose down -v

gensql:
	cd sqlboiler_config && sqlboiler psql -o ../internal/infrastructure/database/models