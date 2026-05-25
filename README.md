# hitalent

`hitalent` is a small Go HTTP API for managing company departments and employees.

The service supports:

- department creation;
- department tree lookup with configurable depth;
- department update and delete;
- employee creation inside a department;
- PostgreSQL persistence;
- schema migrations with `goose`;
- Docker Compose startup for app, database, and migrations;
- handler tests with `httptest` and `testify`.

## Stack

- Go 1.26
- PostgreSQL 16
- GORM
- Goose migrations
- Docker / Docker Compose

## Quick Start

The default configuration does not require a local `.env` file. Docker Compose falls back to:

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=hitalent
GOOSE_TABLE=public.schema_migrations
```

Run the full stack:

```bash
docker compose up --build app
```

This command starts PostgreSQL, runs migrations through the `migrate` service, and then starts the API.

The API will be available at:

```text
http://localhost:8080
```

To run in the background:

```bash
docker compose up --build -d app
```

Check service status:

```bash
docker compose ps
```

Follow logs:

```bash
docker compose logs -f
```

Stop services:

```bash
docker compose down
```

Stop services and remove the database volume:

```bash
docker compose down -v
```

## Environment

`.env` is optional and should not be committed. Use it only when you want to override local values.

Create it from the example:

```bash
cp .env.example .env
```

Then edit values if needed.

## Migrations

Migrations live in:

```text
internal/db/migrations
```

Docker Compose runs migrations automatically before `app` starts.

You can also run only the database and migrations:

```bash
docker compose up --build db migrate
```

Note: without `-d`, this command stays attached to the long-running `db` logs after `migrate` exits.

For local Goose usage:

```bash
goose -dir internal/db/migrations -table public.schema_migrations postgres "postgres://postgres:postgres@localhost:5432/hitalent?sslmode=disable" up
```

Create a new migration:

```bash
make create-migrate name=add_example_table
```

## Tests

Run all tests:

```bash
make test
```

Verbose mode:

```bash
make test-verbose
```

The current tests cover HTTP handlers using `httptest` and `testify` mocks.

## Makefile

Useful commands:

```bash
make help
make test
make fmt
make build
make up
make up-detached
make db
make migrate
make migrate-local
make migrate-status
make logs
make ps
make down
make down-all
```

## API

Create a root department:

```bash
curl -X POST http://localhost:8080/departments \
  -H 'Content-Type: application/json' \
  -d '{"name":"Engineering"}'
```

Create a child department:

```bash
curl -X POST http://localhost:8080/departments \
  -H 'Content-Type: application/json' \
  -d '{"name":"Backend","parent_id":1}'
```

Get a department tree:

```bash
curl 'http://localhost:8080/departments/1?depth=2&include_employees=true'
```

Update a department:

```bash
curl -X PATCH http://localhost:8080/departments/1/update \
  -H 'Content-Type: application/json' \
  -d '{"name":"Platform","parent_id":null}'
```

Delete a department with cascade mode:

```bash
curl -X DELETE 'http://localhost:8080/departments/1/delete?mode=cascade'
```

Delete a department and reassign its employees:

```bash
curl -X DELETE 'http://localhost:8080/departments/1/delete?mode=reassign&reassign_to_department_id=2'
```

Create an employee:

```bash
curl -X POST http://localhost:8080/departments/1/employees \
  -H 'Content-Type: application/json' \
  -d '{"full_name":"Ada Lovelace","position":"Engineer"}'
```

## Project Layout

```text
cmd/hitalent-app/              application entrypoint
internal/app/hitalent-app/     app wiring, router, database connection
internal/department/           department business logic
internal/department/api/       department HTTP handlers and tests
internal/employee/             employee business logic
internal/employee/api/         employee HTTP handlers and tests
internal/db/migrations/        goose SQL migrations
internal/model/                request and database models
internal/errors/               domain errors
```
