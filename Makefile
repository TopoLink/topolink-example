.PHONY: build test run consumer topolink verify compose-up compose-down

build:
	go build ./...

test:
	go test -race ./...

run:
	go run ./cmd/api

consumer:
	go run ./cmd/consumer

topolink:
	go run github.com/topolink/topolink/cmd/topolink@v0.2.0 -repo . -format markdown

verify:
	./scripts/verify.sh

compose-up:
	docker compose -f deploy/docker-compose.yml up --build

compose-down:
	docker compose -f deploy/docker-compose.yml down -v
