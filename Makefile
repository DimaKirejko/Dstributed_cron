-include .env
export

export PROJECT_ROOT=${shell pwd}

pg-up:
	@docker compose up -d dc-postgres
	@docker compose up -d port-forwarder

pg-down:
	@docker compose down dc-postgres
	@docker compose down port-forwarder

pg-cleanup:
	@read -p "CLEANUP ALL VOLUME?? [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down dc-postgres port-forwarder && \
		rm -rf ${PROJECT_ROOT}/out/pgdata && \
		echo "DB file is cleared"; \
	else \
		echo "Attempt to clear DB rejected"; \
	fi

pg-port-forward:
	@docker compose up -d port-forwarder

pg-port-close:
	@docker compose down port-forwarder

up_metrics:
	@docker compose up -d prometheus
	@docker compose up -d grafana

down_metrics:
	@docker compose down prometheus
	@docker compose down grafana

up_prometheus:
	@docker compose up -d prometheus

down_prometheus:
	@docker compose down prometheus

up_grafana:
	@docker compose up -d grafana

down_grafana:
	@docker compose down grafana

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "seq is required. example: make migrate-create seq=init"; \
		exit 1; \
	fi; \
	docker compose run --rm dc-postgres-migrate \
		create -ext sql -dir /migrations -seq "$(seq)"

migrate-up:
	@ make migrate-action action=up

migrate-down:
	@make migrate-action action=down

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "action is required. example: make migrate-action action=up"; \
		exit 1; \
	fi; \
	docker compose run --rm dc-postgres-migrate \
		-path /migrations \
		-database "postgres://${PG_USER}:${PG_PASS}@dc-postgres:5432/${PG_DB}?sslmode=disable" \
		$(action)

run-DC:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	export PG_HOST=localhost && \
	go mod tidy && \
	go run ${PROJECT_ROOT}/cmd/DC/main.go

DC-deploy:
	docker compose up -d --build dc

swagger-gen:
	@docker compose run --rm swagger \
		init \
		-g cmd/DC/main.go \
		-o docs \
		--parseInternal \
		--parseDependency

swagger-fast-setup:
	@docker run --rm -p 18080:8080 \
  		-e SWAGGER_JSON=/docs/swagger.json \
 		-v "$PWD/docs:/docs" \
  		swaggerapi/swagger-ui

test:
	GOCACHE=$${GOCACHE:-/tmp/go-build} go test ./...
