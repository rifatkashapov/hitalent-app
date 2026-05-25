GO ?= go
DOCKER_COMPOSE ?= docker compose
GOCACHE ?= /tmp/go-build-cache

DB_HOST ?= localhost
DB_PORT ?= 5432
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= postgres
POSTGRES_DB ?= hitalent

GOOSE_DRIVER ?= postgres
GOOSE_DBSTRING ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(POSTGRES_DB)?sslmode=disable
GOOSE_MIGRATION_DIR ?= internal/db/migrations
GOOSE_TABLE ?= public.schema_migrations

.PHONY: help test test-verbose fmt build up up-detached up_detached app db db_up migrate migrate-local migrate-status create-migrate create_migrate logs ps down down-all down_all

help:
	@echo "Available targets:"
	@echo "  make test             Run all Go tests"
	@echo "  make test-verbose     Run all Go tests with verbose output"
	@echo "  make fmt              Format Go files"
	@echo "  make build            Build the Go app locally"
	@echo "  make up               Build and run app stack"
	@echo "  make up-detached      Build and run app stack in background"
	@echo "  make db               Run only Postgres"
	@echo "  make migrate          Run migrations through docker compose"
	@echo "  make migrate-local    Run migrations from local goose"
	@echo "  make migrate-status   Show local goose migration status"
	@echo "  make create-migrate name=foo"
	@echo "  make logs             Follow compose logs"
	@echo "  make ps               Show compose services"
	@echo "  make down             Stop compose services"
	@echo "  make down-all         Stop services and remove volumes"
	@echo ""
	@echo "Database variables:"
	@echo "  DB_HOST=$(DB_HOST)"
	@echo "  DB_PORT=$(DB_PORT)"
	@echo "  POSTGRES_USER=$(POSTGRES_USER)"
	@echo "  POSTGRES_DB=$(POSTGRES_DB)"
	@echo "  GOOSE_TABLE=$(GOOSE_TABLE)"

test:
	GOCACHE=$(GOCACHE) $(GO) test ./...

test-verbose:
	GOCACHE=$(GOCACHE) $(GO) test -v ./...

fmt:
	$(GO) fmt ./...

build:
	$(GO) build ./cmd/hitalent-app

up:
	$(DOCKER_COMPOSE) up --build app

up-detached:
	$(DOCKER_COMPOSE) up --build -d app

up_detached: up-detached

app: up

db:
	$(DOCKER_COMPOSE) up -d db

db_up: db

migrate:
	$(DOCKER_COMPOSE) up --build migrate

migrate-local:
	goose -dir $(GOOSE_MIGRATION_DIR) -table $(GOOSE_TABLE) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" up

migrate-status:
	goose -dir $(GOOSE_MIGRATION_DIR) -table $(GOOSE_TABLE) $(GOOSE_DRIVER) "$(GOOSE_DBSTRING)" status

create-migrate:
	goose -dir $(GOOSE_MIGRATION_DIR) create $(name) sql

create_migrate: create-migrate

logs:
	$(DOCKER_COMPOSE) logs -f

ps:
	$(DOCKER_COMPOSE) ps

down:
	$(DOCKER_COMPOSE) down

down-all:
	$(DOCKER_COMPOSE) down -v

down_all: down-all
