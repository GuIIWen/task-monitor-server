package repository

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	cleanup := func() {
		sqlDB.Close()
	}

	return gormDB, mock, cleanup
}

func TestNodeRepository_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNodeRepository(db)
	hostID := "host-001"
	hostname := "test-host"
	status := "online"
	node := &model.Node{
		NodeID:   "node-001",
		HostID:   &hostID,
		Hostname: &hostname,
		Status:   &status,
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `nodes`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(node)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNodeRepository_FindByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNodeRepository(db)
	nodeID := "node-001"

	rows := sqlmock.NewRows([]string{"node_id", "host_id", "hostname", "status"}).
		AddRow("node-001", "host-001", "test-host", "online")

	mock.ExpectQuery("SELECT \\* FROM `nodes` WHERE node_id = \\? ORDER BY `nodes`.`node_id` LIMIT 1").
		WithArgs(nodeID).
		WillReturnRows(rows)

	node, err := repo.FindByID(nodeID)
	assert.NoError(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, "node-001", node.NodeID)
	assert.Equal(t, "test-host", *node.Hostname)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNodeRepository_FindByID_NotFound(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNodeRepository(db)
	nodeID := "non-existent"

	mock.ExpectQuery("SELECT \\* FROM `nodes` WHERE node_id = \\? ORDER BY `nodes`.`node_id` LIMIT 1").
		WithArgs(nodeID).
		WillReturnError(sql.ErrNoRows)

	node, err := repo.FindByID(nodeID)
	assert.Error(t, err)
	assert.Nil(t, node)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNodeRepository_FindAll(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNodeRepository(db)

	rows := sqlmock.NewRows([]string{"node_id", "host_id", "hostname", "status"}).
		AddRow("node-001", "host-001", "test-host-1", "online").
		AddRow("node-002", "host-002", "test-host-2", "offline")

	mock.ExpectQuery("SELECT \\* FROM `nodes`").
		WillReturnRows(rows)

	nodes, err := repo.FindAll()
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)
	assert.Equal(t, "node-001", nodes[0].NodeID)
	assert.Equal(t, "node-002", nodes[1].NodeID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNodeRepository_FindByStatus(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNodeRepository(db)
	status := "online"

	rows := sqlmock.NewRows([]string{"node_id", "host_id", "hostname", "status"}).
		AddRow("node-001", "host-001", "test-host-1", "online")

	mock.ExpectQuery("SELECT \\* FROM `nodes` WHERE status = \\?").
		WithArgs(status).
		WillReturnRows(rows)

	nodes, err := repo.FindByStatus(status)
	assert.NoError(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "online", *nodes[0].Status)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNodeRepository_Update(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNodeRepository(db)
	hostID := "host-001"
	hostname := "updated-host"
	status := "online"
	node := &model.Node{
		NodeID:   "node-001",
		HostID:   &hostID,
		Hostname: &hostname,
		Status:   &status,
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `nodes`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Update(node)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNodeRepository_Upsert(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNodeRepository(db)
	hostID := "host-001"
	hostname := "test-host"
	status := "online"
	node := &model.Node{
		NodeID:   "node-001",
		HostID:   &hostID,
		Hostname: &hostname,
		Status:   &status,
	}

	mock.ExpectBegin()
	// GORM's Save() does UPDATE when primary key exists
	mock.ExpectExec("UPDATE `nodes`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Upsert(node)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestNodeRepository_UpdateHeartbeat(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewNodeRepository(db)
	nodeID := "node-001"

	mock.ExpectBegin()
	// GORM automatically adds updated_at field
	mock.ExpectExec("UPDATE `nodes` SET `last_heartbeat`=NOW\\(\\),`updated_at`=\\? WHERE node_id = \\?").
		WithArgs(sqlmock.AnyArg(), nodeID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateHeartbeat(nodeID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
