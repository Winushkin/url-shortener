build:
	docker build -t url_shortener_app .

local:
	docker compose -f deployments/docker-compose.dev.yaml -f deployments/docker-compose.override.yaml --env-file .env up --build -d

prod:
	docker compose -f deployments/docker-compose.yaml --env-file .env up --build -d

up:
	docker compose -f deployments/docker-compose.yaml --env-file .env up -d

downLocal:
	docker compose -f deployments/docker-compose.dev.yaml --env-file .env down -v

downProd:
	docker compose -f deployments/docker-compose.dev.yaml --env-file .env down -v

testUsecase:
	docker compose -f deployments/docker-compose.test.yaml up --build -d
	go test -bench=. -benchmem ./internal/usecase/benchmarks
	docker compose -f deployments/docker-compose.test.yaml down -v

lint:
	golangci-lint run

test:
	go test ./...

check_fmt:
	@go fmt ./...
	@git diff --exit-code --quiet

clean_vols:
	sudo rm -r deployments/pgdata
	sudo rm -r deployments/kafka_data
	sudo rm -r deployments/redis_data
	sudo rm -r deployments/test_pgdata
	
