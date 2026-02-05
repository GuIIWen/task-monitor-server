package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
)

// MockNodeService is a mock implementation of NodeService
type MockNodeService struct {
	mock.Mock
}

func (m *MockNodeService) GetNodes() ([]model.Node, error) {
	args := m.Called()
	return args.Get(0).([]model.Node), args.Error(1)
}

func (m *MockNodeService) GetNodeByID(nodeID string) (*model.Node, error) {
	args := m.Called(nodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Node), args.Error(1)
}

func (m *MockNodeService) GetNodesByStatus(status string) ([]model.Node, error) {
	args := m.Called(status)
	return args.Get(0).([]model.Node), args.Error(1)
}

func TestNodeHandler_GetNodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockNodeService)
	handler := NewNodeHandler(mockService)

	hostname1 := "host1"
	hostname2 := "host2"
	expectedNodes := []model.Node{
		{NodeID: "node-001", Hostname: &hostname1},
		{NodeID: "node-002", Hostname: &hostname2},
	}

	mockService.On("GetNodes").Return(expectedNodes, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/nodes", nil)

	handler.GetNodes(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	assert.Equal(t, "success", response["message"])

	mockService.AssertExpectations(t)
}

func TestNodeHandler_GetNodes_WithStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockNodeService)
	handler := NewNodeHandler(mockService)

	hostname := "host1"
	status := "online"
	expectedNodes := []model.Node{
		{NodeID: "node-001", Hostname: &hostname, Status: &status},
	}

	mockService.On("GetNodesByStatus", "online").Return(expectedNodes, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/nodes?status=online", nil)

	handler.GetNodes(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestNodeHandler_GetNodes_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockNodeService)
	handler := NewNodeHandler(mockService)

	mockService.On("GetNodes").Return([]model.Node{}, errors.New("database error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/nodes", nil)

	handler.GetNodes(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockService.AssertExpectations(t)
}

func TestNodeHandler_GetNodeByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockNodeService)
	handler := NewNodeHandler(mockService)

	hostname := "test-host"
	expectedNode := &model.Node{
		NodeID:   "node-001",
		Hostname: &hostname,
	}

	mockService.On("GetNodeByID", "node-001").Return(expectedNode, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "nodeId", Value: "node-001"}}
	c.Request = httptest.NewRequest("GET", "/api/v1/nodes/node-001", nil)

	handler.GetNodeByID(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])

	mockService.AssertExpectations(t)
}

func TestNodeHandler_GetNodeByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockNodeService)
	handler := NewNodeHandler(mockService)

	mockService.On("GetNodeByID", "non-existent").Return(nil, gorm.ErrRecordNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "nodeId", Value: "non-existent"}}
	c.Request = httptest.NewRequest("GET", "/api/v1/nodes/non-existent", nil)

	handler.GetNodeByID(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
