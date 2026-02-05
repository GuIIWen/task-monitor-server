package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/task-monitor/api-server/internal/model"
)

func TestJobRepository_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	nodeID := "node-001"
	jobName := "test-job"
	jobType := "training"
	status := "running"
	job := &model.Job{
		JobID:   "job-001",
		NodeID:  &nodeID,
		JobName: &jobName,
		JobType: &jobType,
		Status:  &status,
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `jobs`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(job)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestJobRepository_FindByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	jobID := "job-001"

	rows := sqlmock.NewRows([]string{"job_id", "node_id", "job_name", "status"}).
		AddRow("job-001", "node-001", "test-job", "running")

	mock.ExpectQuery("SELECT \\* FROM `jobs` WHERE job_id = \\?").
		WithArgs(jobID).
		WillReturnRows(rows)

	job, err := repo.FindByID(jobID)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-001", job.JobID)
	assert.Equal(t, "test-job", *job.JobName)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestJobRepository_FindByNodeID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	nodeID := "node-001"

	rows := sqlmock.NewRows([]string{"job_id", "node_id", "job_name", "status"}).
		AddRow("job-001", "node-001", "test-job-1", "running").
		AddRow("job-002", "node-001", "test-job-2", "completed")

	mock.ExpectQuery("SELECT \\* FROM `jobs` WHERE node_id = \\?").
		WithArgs(nodeID).
		WillReturnRows(rows)

	jobs, err := repo.FindByNodeID(nodeID)
	assert.NoError(t, err)
	assert.Len(t, jobs, 2)
	assert.Equal(t, "node-001", *jobs[0].NodeID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestJobRepository_FindByStatus(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	status := "running"

	rows := sqlmock.NewRows([]string{"job_id", "node_id", "job_name", "status"}).
		AddRow("job-001", "node-001", "test-job", "running")

	mock.ExpectQuery("SELECT \\* FROM `jobs` WHERE status = \\?").
		WithArgs(status).
		WillReturnRows(rows)

	jobs, err := repo.FindByStatus(status)
	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "running", *jobs[0].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestJobRepository_UpdateStatus(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	jobID := "job-001"
	status := "completed"
	reason := "task finished"

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `jobs` SET").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery("SELECT \\* FROM `jobs`").
		WillReturnRows(sqlmock.NewRows([]string{"job_id", "status"}).AddRow(jobID, "running"))
	mock.ExpectExec("INSERT INTO `job_status_histories`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateStatus(jobID, status, reason)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestJobRepository_Upsert(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewJobRepository(db)
	nodeID := "node-001"
	jobName := "test-job"
	status := "running"
	job := &model.Job{
		JobID:   "job-001",
		NodeID:  &nodeID,
		JobName: &jobName,
		Status:  &status,
	}

	mock.ExpectBegin()
	// GORM's Save() does UPDATE when primary key exists
	mock.ExpectExec("UPDATE `jobs`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Upsert(job)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
