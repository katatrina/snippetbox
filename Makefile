run:
	go run ./cmd/web

# You need to provide password for user "postgres" when prompting
createdb:
	createdb -U postgres -O postgres -W snippetbox

# You need to provide password for user "postgres" when prompting
dropdb:
	dropdb -U postgres -W snippetbox

create-migrate:
	migrate create -ext sql -dir ./internal/db/migrations/ -seq -digits 6 init_db

migrateup:
	migrate -verbose -database postgres://postgres:12345@localhost:5432/snippetbox?sslmode=disable -path internal/db/migrations up

migratedown:
	migrate -verbose -database postgres://postgres:12345@localhost:5432/snippetbox?sslmode=disable -path internal/db/migrations down

sqlc:
	docker run --rm -v $(CURDIR):/src -w /src sqlc/sqlc generate

.PHONY: run createdb dropdb create-migrate migrateup migratedown sqlc