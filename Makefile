compose:
	docker compose up -d

migrate:
	migrate -path migrations -database "postgres://auth_user:auth_pass@localhost:5432/auth_db?sslmode=disable" up

build:
	go build cmd/auth/main.go

run:
	./main --config=./config/local.yaml