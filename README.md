# Distributed Cron

Small delayed-jobs service for scheduled HTTP calls and internal PostgreSQL maintenance tasks.

## What It Does

- Creates jobs through HTTP API.
- Schedules jobs by daily `run_at` time.
- Executes jobs with a DB-backed worker pool.
- Prevents double execution with PostgreSQL locks: `locked_by` and `lock_until`.
- Supports retry attempts and attempt history.
- Supports one-shot and repeatable jobs via `is_repetable`.

## Main Components

- **API**: create, inspect, cancel, and rerun jobs.
- **Scheduler**: marks due jobs as ready for execution.
- **Worker pool**: claims ready jobs, executes them, and stores attempts.
- **PostgreSQL**: source of truth for jobs, locks, and attempts.

## Local Run

Create/update `.env`, then:

```bash
make pg-up
make migrate-up
make run-DC
```

Useful commands:

```bash
make test
make pg-down
make pg-cleanup
```

## Important Env

```env
PG_USER=main_user
PG_PASS=main_password
PG_DB=cron_db
PG_TIMEZONE=Europe/Kyiv
PG_TIMEOUT=10s

LOGGER_LEVEL=INFO
LOGGER_FOLDER=./out/logs

EXECUTOR_ALLOWLIST=httpbin.org,example.com
SCHEDULER_IS_TEST_MODE=false
```

`SCHEDULER_IS_TEST_MODE=true` keeps the more permissive scheduler behavior useful for local testing.

## Tests

```bash
make test
```

GitHub Actions runs the same target on push and pull request.

## Job Types

HTTP job:

```json
{
  "type": "http",
  "run_at": "12:30",
  "http_method": "GET",
  "http_url": "https://example.com",
  "is_repetable": true
}
```

DB job:

```json
{
  "type": "db",
  "run_at": "01:00",
  "db_action": "create_partition",
  "target_db": "cronapp.events",
  "is_repetable": true
}
```

Supported DB actions:

- `create_partition`
- `year_cleanup`

## Notes

HTTP jobs are restricted by `EXECUTOR_ALLOWLIST`.

DB jobs do not execute arbitrary SQL. They use internal allowlisted actions only.
