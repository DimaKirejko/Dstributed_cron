CREATE SCHEMA cronapp;

CREATE TYPE job_type AS ENUM ('http', 'db');
CREATE TYPE job_status AS ENUM ('queued', 'running', 'succeeded', 'failed', 'canceled');
CREATE TYPE job_http_methods AS ENUM ('GET');
CREATE TYPE job_db_action AS ENUM ('crate_partition');

CREATE TABLE cronapp.jobs (
    id SERIAL PRIMARY KEY,
    type        job_type,
    status      job_status,
    run_at      TIMESTAMPTZ NOT NULL,
    attempt     INT,
    max_retries INT NOT NULL,
    http_method job_http_methods,
    http_url    VARCHAR(100),
    db_action   job_db_action,
    last_error  text,
    locked_by   INT,
    lock_until  TIMESTAMPTZ
);