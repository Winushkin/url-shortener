build:
	docker compose -f deployments/docker-compose.yaml --env-file .env up --build
