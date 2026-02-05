package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestParameterRepository_FindByJobID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewParameterRepository(db)
	jobID := "job-001"

	rows := sqlmock.NewRows([]string{"id", "job_id", "parameter_raw", "parameter_data"}).
		AddRow(1, "job-001", "learning_rate=0.001", `{"learning_rate": 0.001}`).
		AddRow(2, "job-001", "batch_size=32", `{"batch_size": 32}`)

	mock.ExpectQuery("SELECT \\* FROM `parameters` WHERE job_id = \\? ORDER BY .*timestamp.* DESC").
		WithArgs(jobID).
		WillReturnRows(rows)

	params, err := repo.FindByJobID(jobID)
	assert.NoError(t, err)
	assert.Len(t, params, 2)
	assert.Equal(t, "learning_rate=0.001", *params[0].ParameterRaw)
	assert.Equal(t, "batch_size=32", *params[1].ParameterRaw)
	assert.NoError(t, mock.ExpectationsWereMet())
}
