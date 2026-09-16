DATABASE_URL ?= postgres://ukapp:ukapp@localhost:5432/ukapp?sslmode=disable

.PHONY: gen gen-api gen-db lint test run db

## gen: openapi.yaml → gen/api; schema.sql → Postgres → gen/db (jet). Требует docker compose up.
gen: gen-api gen-db

gen-api:
	go tool oapi-codegen -config oapi-codegen.yaml openapi.yaml

gen-db: db
	psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -q -f api/schema.sql
	rm -rf gen/db
	go tool jet -source=postgres -dsn="$(DATABASE_URL)" -schema=public -path=./gen/db

db:
	docker compose up -d --wait db

lint:
	golangci-lint run ./...

test:
	go test ./...

run:
	go run ./cmd/api
