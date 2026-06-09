build:
	docker build -t url_shortener_app .

buildup:
	docker compose -f deployments/docker-compose.yaml --env-file .env up --build -d

down:
	docker compose -f deployments/docker-compose.yaml --env-file .env down -v

testUsecase:
	docker compose -f deployments/docker-compose.test.yaml up --build -d
	go test -bench=. -benchmem ./internal/usecase/benchmarks


lint:
	golangci-lint run

test:
	go test ./...

check_fmt:
	@go fmt ./...
	@git diff --exit-code --quiet