.PHONY: test all cover up down info clean

ifneq (,$(wildcard .env))
include .env
export
endif

MAIN       := ./cmd/main.go
DC         := docker compose
DC_TEST    := -f tests/docker-compose_test.yml
DB_SERVICE := postgres_container
INTERNAL := ./internal
TESTS := ./services/ ./repository/ ./handlers/ ./validation/


all: test

test: test_integration test_unit

test_unit:
	@cd $(INTERNAL) && go test $(TESTS) -cover


test_integration:
	@$(DC) $(DC_TEST) down -v
	$(DC) $(DC_TEST) up --build -d
	@sleep 2
	go test ./tests -timeout 20s
	@$(DC) $(DC_TEST) down -v

cover:
	@cd $(INTERNAL) && go test $(TESTS) -coverprofile=cover.out && go tool cover -html=cover.out -o ../cover.html
	@rm -f $(INTERNAL)/cover.out

up:
	$(DC) up --build

down:
	$(DC) down

up_database:
	$(DC) up db -d

down_database:
	$(DC) down db -d


info:
	@$(DC) ps -a
	@echo ---------------------------------------------------------------------------
	@docker ps -a


migrate_init:
	@docker exec -i postgres_container psql -U $(DB_USER) -d $(DB_NAME) < ./sql_scripts/init.sql

migrate_drop:
	@docker exec -i postgres_container psql -U $(DB_USER) -d $(DB_NAME) < ./sql_scripts/drop/drop.sql

migrate_clear:
	@docker exec -i postgres_container psql -U $(DB_USER) -d $(DB_NAME) < ./sql_scripts/drop/clear.sql



clean: clean_cover clean_test_cont clean_container 

clean_cover:
	rm -f cover.html

clean_test_cont:
	$(DC) $(DC_TEST) down -v

clean_container:
	$(DC) down -v