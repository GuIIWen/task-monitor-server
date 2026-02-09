package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestMetricsRepository_FindNPUCardsByPIDs_OnlyRunning(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewMetricsRepository(db)

	rows := sqlmock.NewRows([]string{"pid", "npu_id"}).
		AddRow(int64(100), 0).
		AddRow(int64(100), 1).
		AddRow(int64(101), 1)

	mock.ExpectQuery("SELECT DISTINCT pid, npu_id FROM `npu_processes` WHERE node_id = \\? AND pid IN \\(\\?,\\?\\) AND status = \\?").
		WithArgs("node-001", int64(100), int64(101), "running").
		WillReturnRows(rows)

	result, err := repo.FindNPUCardsByPIDs("node-001", []int64{100, 101})
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.ElementsMatch(t, []int{0, 1}, result[100])
	assert.ElementsMatch(t, []int{1}, result[101])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepository_FindNPUCardsByPIDs_EmptyPIDs(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewMetricsRepository(db)

	result, err := repo.FindNPUCardsByPIDs("node-001", nil)
	assert.NoError(t, err)
	assert.Empty(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepository_DistinctNPUCardCounts_OnlyRunning(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewMetricsRepository(db)

	rows := sqlmock.NewRows([]string{"card_count"}).
		AddRow(1).
		AddRow(2).
		AddRow(2)

	mock.ExpectQuery("SELECT COUNT\\(DISTINCT np.npu_id\\) AS card_count[\\s\\S]*np.status = 'running'[\\s\\S]*GROUP BY j.node_id, j.pgid, j.start_time").
		WillReturnRows(rows)

	result, err := repo.DistinctNPUCardCounts()
	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2}, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepository_FindNPUProcessesByPID_OnlyRunning(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewMetricsRepository(db)

	rows := sqlmock.NewRows([]string{"id", "node_id", "pid", "npu_id", "status"}).
		AddRow(1, "node-001", int64(100), 0, "running").
		AddRow(2, "node-001", int64(100), 1, "running")

	mock.ExpectQuery("SELECT \\* FROM `npu_processes` WHERE node_id = \\? AND pid = \\? AND status = \\?").
		WithArgs("node-001", int64(100), "running").
		WillReturnRows(rows)

	processes, err := repo.FindNPUProcessesByPID("node-001", 100)
	assert.NoError(t, err)
	assert.Len(t, processes, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}
