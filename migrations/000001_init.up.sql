CREATE SCHEMA cronapp;

CREATE TYPE job_type AS ENUM ('http', 'db');
CREATE TYPE job_status AS ENUM ('queued', 'running', 'succeeded', 'failed', 'canceled');
CREATE TYPE job_http_methods AS ENUM ('GET');
CREATE TYPE job_db_action AS ENUM ('create_partition');
CREATE TYPE attempt_results AS ENUM ('success', 'retry', 'failed');

CREATE TABLE cronapp.jobs (
    id          SERIAL PRIMARY KEY,
    type        job_type NOT NULL,
    status      job_status NOT NULL,
    daily_run_time      TIME NOT NULL,
    attempt     INTEGER NOT NULL,
    max_retries INTEGER NOT NULL,
    http_method job_http_methods,
    http_url    VARCHAR(100),
    db_action   job_db_action,
    last_error  text,
    locked_by   INTEGER,
    lock_until  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE INDEX jobs_queued_daily_run_time_idx
ON cronapp.jobs (daily_run_time, id)
WHERE status = 'queued';

CREATE TABLE cronapp.attempts (
    id SERIAL PRIMARY KEY,
    job_id INTEGER NOT NULL REFERENCES cronapp.jobs(id),
    result attempt_results,
    http_status INTEGER,
    error_message text,
    started_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX attempts_job_id_idx
ON cronapp.attempts (job_id);