build:
	docker build -t url_shortener_app .

buildup:
	docker compose -f deployments/docker-compose.yaml --env-file .env up --build -d

down:
	docker compose -f deployments/docker-compose.yaml --env-file .env down
