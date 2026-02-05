package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/task-monitor/api-server/internal/model"
)

func TestMetricsRepository_CreateNPUMetric(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewMetricsRepository(db)
	nodeID := "node-001"
	npuID := 0
	utilization := 85.5
	memory := 8192.0
	metric := &model.NPUMetric{
		NodeID:             &nodeID,
		NPUID:              &npuID,
		AICoreUsagePercent: &utilization,
		MemoryUsageMB:      &memory,
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `npu_metrics`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.CreateNPUMetric(metric)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepository_CreateProcessMetric(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewMetricsRepository(db)
	jobID := "job-001"
	cpuPercent := 45.2
	memoryMB := 2048.0
	metric := &model.ProcessMetric{
		JobID:      &jobID,
		CPUPercent: &cpuPercent,
		MemoryMB:   &memoryMB,
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `process_metrics`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.CreateProcessMetric(metric)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMetricsRepository_BatchCreateNPUMetrics(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewMetricsRepository(db)
	nodeID := "node-001"
	npuID0 := 0
	npuID1 := 1
	util0 := 85.5
	util1 := 90.0
	metrics := []model.NPUMetric{
		{NodeID: &nodeID, NPUID: &npuID0, AICoreUsagePercent: &util0},
		{NodeID: &nodeID, NPUID: &npuID1, AICoreUsagePercent: &util1},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `npu_metrics`").
		WillReturnResult(sqlmock.NewResult(1, 2))
	mock.ExpectCommit()

	err := repo.BatchCreateNPUMetrics(metrics)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
