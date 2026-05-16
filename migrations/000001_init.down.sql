DROP TYPE IF EXISTS job_type;
DROP TYPE IF EXISTS job_status;
DROP TYPE IF EXISTS job_http_methods;
DROP TYPE IF EXISTS job_db_action;
DROP TYPE IF EXISTS attempt_results;

DROP INDEX IF EXISTS jobs_queued_daily_run_time_idx ON cronapp.jobs;
DROP TABLE IF EXISTS cronapp.jobs;

DROP SCHEMA IF EXISTS cronapp;
