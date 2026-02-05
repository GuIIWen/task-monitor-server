package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/task-monitor/api-server/internal/model"
)

func TestCodeRepository_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewCodeRepository(db)
	jobID := "job-001"
	scriptPath := "/path/to/script.py"
	scriptContent := "print('hello')"
	code := &model.Code{
		JobID:         &jobID,
		ScriptPath:    &scriptPath,
		ScriptContent: &scriptContent,
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `code`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(code)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCodeRepository_FindByJobID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewCodeRepository(db)
	jobID := "job-001"

	rows := sqlmock.NewRows([]string{"id", "job_id", "script_path", "script_content"}).
		AddRow(1, "job-001", "/path/to/script.py", "print('hello')").
		AddRow(2, "job-001", "/path/to/config.yaml", "key: value")

	mock.ExpectQuery("SELECT \\* FROM `code` WHERE job_id = \\?").
		WithArgs(jobID).
		WillReturnRows(rows)

	codes, err := repo.FindByJobID(jobID)
	assert.NoError(t, err)
	assert.Len(t, codes, 2)
	assert.Equal(t, "/path/to/script.py", *codes[0].ScriptPath)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCodeRepository_BatchCreate(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewCodeRepository(db)
	jobID := "job-001"
	scriptPath1 := "/path/to/script.py"
	scriptContent1 := "print('hello')"
	scriptPath2 := "/path/to/config.yaml"
	scriptContent2 := "key: value"
	codes := []model.Code{
		{JobID: &jobID, ScriptPath: &scriptPath1, ScriptContent: &scriptContent1},
		{JobID: &jobID, ScriptPath: &scriptPath2, ScriptContent: &scriptContent2},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `code`").
		WillReturnResult(sqlmock.NewResult(1, 2))
	mock.ExpectCommit()

	err := repo.BatchCreate(codes)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
