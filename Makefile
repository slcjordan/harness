.SECONDARY:

MIGRATE_PATH=db/migrations
PORT=53443

# overrideable
DEV_NAMESPACE?=harness-$(shell git rev-parse --abbrev-ref HEAD)
NETWORK?=$(shell docker-compose --project-name harness-grpcui config --format json | jq -r '.networks.backend.name')
HUGO_VERSION?=reg-git-non-root-0.136.5
PGHOST?=${DEV_NAMESPACE}-db
PGPASSWORD?=changeme
PGPORT?=5432
PGUSER?=user
POSTGRES_VERSION?=16.9
PGDATABASE?=postgres
DB_CONN_STRING?=postgresql://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/${PGDATABASE}?sslmode=disable
# OPEN_BROWSER?=open -a "Google Chrome"
OPEN_BROWSER=google-chrome
PGADMIN_DEFAULT_EMAIL?=admin@${PGHOST}.com
PGADMIN_DEFAULT_PASSWORD?=${PGPASSWORD}
PGADMIN_VERSION?=9.5.0
GOMIGRATE_VERSION?=v4.18.3
PROMPT_MIGRATION_NAME?=$(shell bash -c 'read -p "Descriptive Filename Part (e.g. sphincs_private_ica): " migration; echo $$migration')
MIGRATE_VERSION_FILE?=$(shell ls "${PWD}/${MIGRATE_PATH}" | sort -t '_' -k 1 -n | tail -n 1)
MIGRATE_VERSION?=$(shell echo "${MIGRATE_VERSION_FILE}" | sed -E 's|([0-9]*)_.*.sql|\1|')
SQLC_VERSION?=1.29.0
HARNESS_SLACK_CLIENT_ID?=secret
HARNESS_SLACK_CLIENT_SECRET?=secret
HARNESS_SLACK_REDIRECT_URL?=https://localhost:${PORT}/slack/oauth/callback
TEMPORALIO_VERSION?=1.28
TEMPORALIO_HOST?=${DEV_NAMESPACE}-temporalio
TEMPORALIO_DB?=temporalio
DUMP_FILENAME?=dump.sql
HARNESS_PORT?=$(shell docker-compose --project-name ${DEV_NAMESPACE} port harness ${PORT})
WORKFLOW_SERVER_PORT?=$(shell docker-compose --project-name ${DEV_NAMESPACE} port workflow-ui 8080)
POSTGRES_CONTAINER_ID?=$(shell docker-compose --project-name ${DEV_NAMESPACE} port -q postgres)


.PHONY: start-all
start-all: hugo-build
	DOCKER_BUILDKIT=1 docker-compose \
		--project-name ${DEV_NAMESPACE} \
		--progress plain \
		up \
			--build \
			--detach

.PHONY: down-all
down-all: hugo-build
	docker-compose \
		--project-name ${DEV_NAMESPACE} \
		down \
			--remove-orphans

.PHONY: debug
debug:
	echo ${POSTGRES_CONTAINER_ID}

.PHONY: open-harness
open-harness:
	${OPEN_BROWSER} ${HARNESS_PORT}/app/pages/portal-tester/

.PHONY: open-workflow
open-workflow:
	${OPEN_BROWSER} ${WORKFLOW_SERVER_PORT}

.PHONY: hugo-build
hugo-build:
	docker run \
		--rm \
		--interactive \
		--tty \
		--volume ${PWD}/ui/:/ui \
		hugomods/hugo:${HUGO_VERSION} \
			hugo --source /ui

.PHONY: temporalio-start
temporalio-start: wait-postgres
	docker run \
		--detach \
		--tty \
		--name ${TEMPORALIO_HOST} \
		--network '${NETWORK}' \
		--env DB=postgres12 \
		--env DBNAME=${TEMPORALIO_DB} \
		--env DB_PORT=${PGPORT} \
		--env POSTGRES_SEEDS=${PGHOST} \
		--env POSTGRES_USER=${PGUSER} \
		--env POSTGRES_PWD=${PGPASSWORD} \
		--publish 53233:7233 \
		--publish 53090:9090 \
		temporalio/auto-setup:${TEMPORALIO_VERSION}

.PHONY: temporalio-wait
temporalio-wait: temporalio-start
	until curl "http://localhost:53090/health"; \
	do \
			sleep 3; \
	done
	${OPEN_BROWSER} "http://localhost:${PGADMIN_PORT}"


.PHONY: go-sqlc
go-sqlc: postgres-schema-dump ## Generate go code from sqlc.
	- rm -f db/sqlc/*.sql.go
	- rm -f db/sqlc/models.go
	- rm -f db/sqlc/db.go
	docker run \
		--interactive \
		--tty \
		--rm \
		--volume ${PWD}:/repo \
		--workdir /repo/db/sqlc \
	sqlc/sqlc:${SQLC_VERSION} generate
	sudo chown -R $(shell id -u):$(shell id -g) db/sqlc
	chmod 440 db/sqlc/*.sql.go
	chmod 440 db/sqlc/models.go
	chmod 440 db/sqlc/db.go

.PHONY: postgres-dump
postgres-dump: wait-postgres
	docker run \
		--interactive \
		--tty \
		--volume ${PWD}:${PWD} \
		--env PGDATABASE=${PGDATABASE} \
		--env PGHOST=${PGHOST} \
		--env PGPASSWORD=${PGPASSWORD} \
		--env PGPORT=${PGPORT} \
		--env PGUSER=${PGUSER} \
		--network '${NETWORK}' \
		--workdir ${PWD} \
		postgres:${POSTGRES_VERSION} pg_dump \
			--file ${DUMP_FILENAME} \
			${PGDATABASE} 

# .PHONY: psql
# psql: wait-postgres ## Start an interactive postgres shell
# 	docker run \
# 		--name ${PGHOST}-psql \
# 		--interactive \
# 		--tty \
# 		--rm \
# 		--network '${NETWORK}' \
# 		postgres:${POSTGRES_VERSION} psql \
# 			-d ${DB_CONN_STRING} \
# 			--pset expanded=auto \
# 			-f -

.PHONY: psql
psql: ## Start an interactive postgres shell
	docker-compose \
		--project-name ${DEV_NAMESPACE} \
		run psql

.PHONY: pgadmin
pgadmin: wait-postgres ## Start pgadmin and open in browser window
	-mkdir -p /tmp/.cache/${DEV_NAMESPACE}/pgadmin/state
	-echo ' { "Servers": { "1": { "Name": "${DEV_NAMESPACE}", "Group": "Servers", "Host": "${PGHOST}", "Port": ${PGPORT}, "Username": "${PGUSER}", "SSLMode": "prefer", "MaintenanceDB": "${PGDATABASE}", "PassFile": "/pgpass" } } } ' > /tmp/.cache/${DEV_NAMESPACE}/pgadmin/servers.json
	-echo '${PGHOST}:${PGPORT}:${PGDATABASE}:${PGUSER}:${PGPASSWORD}' > /tmp/.cache/${DEV_NAMESPACE}/pgadmin/pgpass
	-chmod 600 /tmp/.cache/${DEV_NAMESPACE}/pgadmin/pgpass
	sudo chown -R 5050:5050 /tmp/.cache/${DEV_NAMESPACE}/pgadmin
	@echo "admin email: ${PGADMIN_DEFAULT_EMAIL}"
	@echo "admin password: ${PGADMIN_DEFAULT_PASSWORD}"
	docker container inspect --format='pgadmin is {{.State.Status}} at port {{(index (index .NetworkSettings.Ports "80/tcp") 0).HostPort}}' ${PGHOST}-pgadmin || docker run \
		--name ${PGHOST}-pgadmin \
		--rm \
		--detach \
		--network '${NETWORK}' \
		--env PGADMIN_CONFIG_SERVER_MODE=False \
		--env PGADMIN_CONFIG_MASTER_PASSWORD_REQUIRED=False \
		--env PGADMIN_DEFAULT_EMAIL=${PGADMIN_DEFAULT_EMAIL} \
		--env PGADMIN_DEFAULT_PASSWORD=${PGADMIN_DEFAULT_PASSWORD} \
		--volume /tmp/.cache/${DEV_NAMESPACE}/pgadmin/state:/var/lib/pgadmin \
		--volume /tmp/.cache/${DEV_NAMESPACE}/pgadmin/servers.json:/pgadmin4/servers.json \
		--volume /tmp/.cache/${DEV_NAMESPACE}/pgadmin/pgpass:/pgpass \
		--publish ${PGADMIN_PORT}:80 \
		dpage/pgadmin4:${PGADMIN_VERSION}
	until curl "http://localhost:${PGADMIN_PORT}"; \
	do \
			sleep 3; \
	done
	${OPEN_BROWSER} "http://localhost:${PGADMIN_PORT}"

.PHONY: pgadmin-stop
pgadmin-stop: ## Stop running pgadmin instance
	docker stop ${PGHOST}-pgadmin

.PHONY: wait-postgres
wait-postgres: start-all ## Start postgres if it isn't started and wait for it to be ready.
	@until [ "$$(docker inspect -f '{{.State.Health.Status}}' ${POSTGRES_CONTAINER_ID})" = "healthy" ]; do \
		echo "Waiting for db to become healthy..."; \
		sleep 1; \
	done; \


.PHONY: postgres-migrate-create
postgres-migrate-create: ## Helps the user create a pair of up/down migration script files and prompts them for a descriptive filename.
	MIGRATE_PATH=${MIGRATE_PATH} \
	MIGRATE_VERSION=${MIGRATE_VERSION} \
	docker-compose \
		--project-name ${DEV_NAMESPACE} \
		run migrate create \
				-ext sql \
				-dir /${MIGRATE_PATH} \
				-seq \
				-digits 3 \
				${PROMPT_MIGRATION_NAME}
	sudo chown -R $(shell id -u):$(shell id -g) ${PWD}/${MIGRATE_PATH}
	chmod 664 ${PWD}/${MIGRATE_PATH}/*

.PHONY: postgres-migrate
postgres-migrate: ## Run all migrations up to the latest version.
	MIGRATE_PATH=${MIGRATE_PATH} \
	MIGRATE_VERSION=${MIGRATE_VERSION} \
	docker-compose \
		--project-name ${DEV_NAMESPACE} \
		run migrate

.PHONY: postgres-migrate-version
postgres-migrate-version: wait-postgres ## Print the currently applied migration version.
	docker run \
		--interactive \
		--tty \
		--rm \
		--name ${PGHOST}-postgres-migrate-version \
		--network '${NETWORK}' \
		--volume ${PWD}/${MIGRATE_PATH}:/${MIGRATE_PATH} \
		migrate/migrate:${GOMIGRATE_VERSION} \
			-database ${DB_CONN_STRING} \
			-path /${MIGRATE_PATH} \
			version

.PHONY: postgres-migrate-force
postgres-migrate-force: wait-postgres ## Force the migration to a specific version. This is useful in case of a failed migration.
	docker run \
		--interactive \
		--tty \
		--rm \
		--name ${PGHOST}-postgres-migrate-force \
		--network '${NETWORK}' \
		--volume ${PWD}/${MIGRATE_PATH}:/${MIGRATE_PATH} \
		migrate/migrate:${GOMIGRATE_VERSION} \
			-database ${DB_CONN_STRING} \
			-path /${MIGRATE_PATH} \
			force ${MIGRATE_VERSION}

.PHONY: postgres-schema-dump
postgres-schema-dump: postgres-migrate ## create postgres schema dump file under db/sqlc, which is necessary for sqlc
	docker-compose \
		--project-name ${DEV_NAMESPACE} \
		run schema-dump
	sudo chown $(shell id -u):$(shell id -g) db/sqlc/schema.sql
	chmod 440 db/sqlc/schema.sql
