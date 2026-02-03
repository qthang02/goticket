.PHONY: run up down logs

run:
	go run cmd/api/main.go

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f
