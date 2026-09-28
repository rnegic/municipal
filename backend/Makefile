DATABASE_URL ?= postgres://ukapp:ukapp@localhost:5432/ukapp?sslmode=disable

.PHONY: gen gen-api gen-db lint test run db
gen: gen-api gen-db

gen-api:
	go tool oapi-codegen -config oapi-codegen.yaml openapi.yaml
	go tool oapi-codegen -config oapi-codegen-uk.yaml contracts/uk.yaml

gen-db: db
	psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -q -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
	psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -q -f internal/repository/schema.sql
	rm -rf gen/db
	go tool jet -source=postgres -dsn="$(DATABASE_URL)" -schema=public -path=./gen/db

db:
	docker compose -f ../docker-compose.yml up -d --wait db

lint:
	golangci-lint run ./...

test:
	go test -p 1 ./...

run:
	go run ./cmd/api
