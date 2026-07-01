.PHONY: build run vet test tidy migrate-up migrate-down

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

vet:
	go vet ./...

test:
	go test ./...

tidy:
	go mod tidy

# Requires golang-migrate installed: https://github.com/golang-migrate/migrate
# DATABASE_URL must be set or pass DATABASE_URL=... on the command line.
migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1
