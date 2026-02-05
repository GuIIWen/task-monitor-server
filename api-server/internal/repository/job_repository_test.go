package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

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
