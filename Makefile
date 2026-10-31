.PHONY: all run test docker-up docker-down tidy clean

all: test run

docker-up:
	docker compose up -d

docker-down:
	docker compose down

tidy:
	go mod tidy

test:
	go test -v ./...

run:
	go run cmd/api/main.go

clean:
	rm -rf bin/
