MOCKUK_DATABASE_URL ?= postgres://mockuk:mockuk@localhost:5433/mockuk?sslmode=disable

.PHONY: gen db test run

gen:
	go tool oapi-codegen -config oapi-codegen.yaml ../../backend/contracts/uk.yaml

db:
	docker compose -f ../../docker-compose.yml up -d --wait mock-uk-db

test:
	MOCKUK_DATABASE_URL="$(MOCKUK_DATABASE_URL)" go test -p 1 ./...

run:
	DATABASE_URL="$(MOCKUK_DATABASE_URL)" MOCK_API_TOKEN=demo-uk-token go run ./cmd/mockuk
