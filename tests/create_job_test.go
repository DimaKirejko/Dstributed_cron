package tests

import (
	"context"
	"testing"

	"github.com/DimaKirejko/Dstributed_cron/internal/core/domain"
	core_job_types "github.com/DimaKirejko/Dstributed_cron/internal/core/types"
	jobs_service "github.com/DimaKirejko/Dstributed_cron/internal/features/jobs/service"
)

type fakeJobRepository struct {
	created domain.Job
}

func (r *fakeJobRepository) Create_job(ctx context.Context, job domain.Job) (core_job_types.ID, error) {
	r.created = job
	return 123, nil
}

func (r *fakeJobRepository) GetJob(ctx context.Context, id core_job_types.ID) (domain.Job, error) {
	return domain.Job{Id: id}, nil
}

func (r *fakeJobRepository) ChangeJob(
	ctx context.Context,
	id core_job_types.ID,
	newStatus core_job_types.Status,
) (domain.Job, error) {
	return domain.Job{Id: id, Status: newStatus}, nil
}

func TestCreateHTTPJobAllowedURLDefaultsToRepetable(t *testing.T) {
	t.Setenv("EXECUTOR_ALLOWLIST", "example.com,httpbin.org")

	repo := &fakeJobRepository{}
	service := jobs_service.NewJobsService(repo)

	method := core_job_types.HTTPGet
	targetURL := "https://example.com/status/200"

	id, err := service.CreateJob(context.Background(), &jobs_service.CreateJobInput{
		Type:         core_job_types.TypeHTTP,
		DailyRunTime: "12:30",
		HTTPMethod:   &method,
		HTTPURL:      &targetURL,
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	if id != 123 {
		t.Fatalf("unexpected job id: got %d want %d", id, 123)
	}

	if !repo.created.IsRepetable {
		t.Fatal("expected omitted is_repetable to default to true")
	}

	if repo.created.HttpURL == nil || *repo.created.HttpURL != targetURL {
		t.Fatalf("unexpected http_url: got %v want %q", repo.created.HttpURL, targetURL)
	}
}

func TestCreateHTTPJobRejectsURLOutsideAllowlist(t *testing.T) {
	t.Setenv("EXECUTOR_ALLOWLIST", "example.com")

	repo := &fakeJobRepository{}
	service := jobs_service.NewJobsService(repo)

	method := core_job_types.HTTPGet
	targetURL := "https://httpbin.org/status/200"

	_, err := service.CreateJob(context.Background(), &jobs_service.CreateJobInput{
		Type:         core_job_types.TypeHTTP,
		DailyRunTime: "12:30",
		HTTPMethod:   &method,
		HTTPURL:      &targetURL,
	})
	if err == nil {
		t.Fatal("expected CreateJob to reject URL outside allowlist")
	}
}

func TestCreateDBJobRejectsTargetWithoutSchema(t *testing.T) {
	repo := &fakeJobRepository{}
	service := jobs_service.NewJobsService(repo)

	action := core_job_types.DBActiYearCleanup
	targetDB := "events"

	_, err := service.CreateJob(context.Background(), &jobs_service.CreateJobInput{
		Type:         core_job_types.TypeDB,
		DailyRunTime: "01:00",
		DBAction:     &action,
		TargetDB:     &targetDB,
	})
	if err == nil {
		t.Fatal("expected CreateJob to reject target_db without schema")
	}
}

func TestCreateDBJobAcceptsSchemaTableTargetAndOneShotFlag(t *testing.T) {
	repo := &fakeJobRepository{}
	service := jobs_service.NewJobsService(repo)

	action := core_job_types.DBActiYearCleanup
	targetDB := "cronapp.events"
	isRepetable := false

	_, err := service.CreateJob(context.Background(), &jobs_service.CreateJobInput{
		Type:         core_job_types.TypeDB,
		DailyRunTime: "01:00",
		IsRepetable:  &isRepetable,
		DBAction:     &action,
		TargetDB:     &targetDB,
	})
	if err != nil {
		t.Fatalf("CreateJob returned error: %v", err)
	}

	if repo.created.IsRepetable {
		t.Fatal("expected explicit is_repetable=false to be preserved")
	}

	if repo.created.TargetDB == nil || *repo.created.TargetDB != targetDB {
		t.Fatalf("unexpected target_db: got %v want %q", repo.created.TargetDB, targetDB)
	}
}
