package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/task-monitor/api-server/internal/model"
)

// MockNodeRepository is a mock implementation of NodeRepository
type MockNodeRepository struct {
	mock.Mock
}

func (m *MockNodeRepository) Create(node *model.Node) error {
	args := m.Called(node)
	return args.Error(0)
}

func (m *MockNodeRepository) Update(node *model.Node) error {
	args := m.Called(node)
	return args.Error(0)
}

func (m *MockNodeRepository) FindByID(nodeID string) (*model.Node, error) {
	args := m.Called(nodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Node), args.Error(1)
}

func (m *MockNodeRepository) FindAll() ([]model.Node, error) {
	args := m.Called()
	return args.Get(0).([]model.Node), args.Error(1)
}

func (m *MockNodeRepository) FindByStatus(status string) ([]model.Node, error) {
	args := m.Called(status)
	return args.Get(0).([]model.Node), args.Error(1)
}

func (m *MockNodeRepository) UpdateHeartbeat(nodeID string) error {
	args := m.Called(nodeID)
	return args.Error(0)
}

func (m *MockNodeRepository) Upsert(node *model.Node) error {
	args := m.Called(node)
	return args.Error(0)
}

func TestNodeService_GetNodes(t *testing.T) {
	mockRepo := new(MockNodeRepository)
	service := NewNodeService(mockRepo)

	hostname1 := "host1"
	hostname2 := "host2"
	expectedNodes := []model.Node{
		{NodeID: "node-001", Hostname: &hostname1},
		{NodeID: "node-002", Hostname: &hostname2},
	}

	mockRepo.On("FindAll").Return(expectedNodes, nil)

	nodes, err := service.GetNodes()

	assert.NoError(t, err)
	assert.Len(t, nodes, 2)
	assert.Equal(t, "node-001", nodes[0].NodeID)
	mockRepo.AssertExpectations(t)
}

func TestNodeService_GetNodes_Error(t *testing.T) {
	mockRepo := new(MockNodeRepository)
	service := NewNodeService(mockRepo)

	mockRepo.On("FindAll").Return([]model.Node{}, errors.New("database error"))

	nodes, err := service.GetNodes()

	assert.Error(t, err)
	assert.Equal(t, "database error", err.Error())
	assert.Empty(t, nodes)
	mockRepo.AssertExpectations(t)
}

func TestNodeService_GetNodeByID(t *testing.T) {
	mockRepo := new(MockNodeRepository)
	service := NewNodeService(mockRepo)

	hostname := "test-host"
	expectedNode := &model.Node{
		NodeID:   "node-001",
		Hostname: &hostname,
	}

	mockRepo.On("FindByID", "node-001").Return(expectedNode, nil)

	node, err := service.GetNodeByID("node-001")

	assert.NoError(t, err)
	assert.NotNil(t, node)
	assert.Equal(t, "node-001", node.NodeID)
	assert.Equal(t, "test-host", *node.Hostname)
	mockRepo.AssertExpectations(t)
}

func TestNodeService_GetNodeByID_NotFound(t *testing.T) {
	mockRepo := new(MockNodeRepository)
	service := NewNodeService(mockRepo)

	mockRepo.On("FindByID", "non-existent").Return(nil, errors.New("not found"))

	node, err := service.GetNodeByID("non-existent")

	assert.Error(t, err)
	assert.Nil(t, node)
	mockRepo.AssertExpectations(t)
}

func TestNodeService_GetNodesByStatus(t *testing.T) {
	mockRepo := new(MockNodeRepository)
	service := NewNodeService(mockRepo)

	hostname := "host1"
	status := "online"
	expectedNodes := []model.Node{
		{NodeID: "node-001", Hostname: &hostname, Status: &status},
	}

	mockRepo.On("FindByStatus", "online").Return(expectedNodes, nil)

	nodes, err := service.GetNodesByStatus("online")

	assert.NoError(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, "online", *nodes[0].Status)
	mockRepo.AssertExpectations(t)
}

func TestNodeService_GetNodesByStatus_Error(t *testing.T) {
	mockRepo := new(MockNodeRepository)
	service := NewNodeService(mockRepo)

	mockRepo.On("FindByStatus", "online").Return([]model.Node{}, errors.New("database error"))

	nodes, err := service.GetNodesByStatus("online")

	assert.Error(t, err)
	assert.Empty(t, nodes)
	mockRepo.AssertExpectations(t)
}
