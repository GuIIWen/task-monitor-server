package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/task-monitor/api-server/internal/model"
)

func TestParameterRepository_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewParameterRepository(db)
	jobID := "job-001"
	paramRaw := "learning_rate=0.001"
	paramData := `{"learning_rate": 0.001}`
	param := &model.Parameter{
		JobID:         &jobID,
		ParameterRaw:  &paramRaw,
		ParameterData: &paramData,
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `parameters`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(param)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestParameterRepository_FindByJobID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewParameterRepository(db)
	jobID := "job-001"

	rows := sqlmock.NewRows([]string{"id", "job_id", "parameter_raw", "parameter_data"}).
		AddRow(1, "job-001", "learning_rate=0.001", `{"learning_rate": 0.001}`).
		AddRow(2, "job-001", "batch_size=32", `{"batch_size": 32}`)

	mock.ExpectQuery("SELECT \\* FROM `parameters` WHERE job_id = \\?").
		WithArgs(jobID).
		WillReturnRows(rows)

	params, err := repo.FindByJobID(jobID)
	assert.NoError(t, err)
	assert.Len(t, params, 2)
	assert.Equal(t, "learning_rate=0.001", *params[0].ParameterRaw)
	assert.Equal(t, "batch_size=32", *params[1].ParameterRaw)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestParameterRepository_BatchCreate(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewParameterRepository(db)
	jobID := "job-001"
	paramRaw1 := "learning_rate=0.001"
	paramData1 := `{"learning_rate": 0.001}`
	paramRaw2 := "batch_size=32"
	paramData2 := `{"batch_size": 32}`
	params := []model.Parameter{
		{JobID: &jobID, ParameterRaw: &paramRaw1, ParameterData: &paramData1},
		{JobID: &jobID, ParameterRaw: &paramRaw2, ParameterData: &paramData2},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `parameters`").
		WillReturnResult(sqlmock.NewResult(1, 2))
	mock.ExpectCommit()

	err := repo.BatchCreate(params)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
