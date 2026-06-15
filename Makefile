build:
	docker build -t url_shortener_app .

buildup:
	docker compose -f deployments/docker-compose.yaml --env-file .env up --build -d

down:
	docker compose -f deployments/docker-compose.yaml --env-file .env down -v
	rm -r deployments/data/kafka
	rm -r deployments/data/redis
	rm -r deployments/data/postgres

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
	