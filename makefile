# Makefile
include .env

MIGRATE_PATH=migrations
MIGRATE_DATABASE="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable"


build:
	go build -o main ./cmd

run: build migrateup sqlc_generate # Include database and sqlc setup
	./main

test:
	go test ./...

migrateup:
	migrate -path $(MIGRATE_PATH) -database $(MIGRATE_DATABASE) up

migratedown:
	migrate -path $(MIGRATE_PATH) -database $(MIGRATE_DATABASE) down

migrateforce:
	migrate -path $(MIGRATE_PATH) -database $(MIGRATE_DATABASE) force $(MIGRATE_VERSION)


sqlcgenerate:
	sqlc generate
# Linting (using golangci-lint)
lint:
	golangci-lint run ./...



dockerbuild:
	docker compose build

dockerup: dockerbuild
	docker compose up

dockerupdetached: dockerbuild
	docker compose up -d

dockerdown:
	docker compose down

dockerps:
	docker compose ps

dockerlogs:
	docker compose logs -f

dockerrestart:
	docker compose restart

clean: dockerdown
	go clean
	rm main
