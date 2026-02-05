package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

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
