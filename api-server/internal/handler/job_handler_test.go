package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
)

// MockJobService is a mock implementation of JobService
type MockJobService struct {
	mock.Mock
}

func (m *MockJobService) GetJobByID(jobID string) (*model.Job, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Job), args.Error(1)
}

func (m *MockJobService) GetJobsByNodeID(nodeID string) ([]model.Job, error) {
	args := m.Called(nodeID)
	return args.Get(0).([]model.Job), args.Error(1)
}

func (m *MockJobService) GetJobsByStatus(status string) ([]model.Job, error) {
	args := m.Called(status)
	return args.Get(0).([]model.Job), args.Error(1)
}

func (m *MockJobService) GetJobParameters(jobID string) ([]model.Parameter, error) {
	args := m.Called(jobID)
	return args.Get(0).([]model.Parameter), args.Error(1)
}

func (m *MockJobService) GetJobCode(jobID string) ([]model.Code, error) {
	args := m.Called(jobID)
	return args.Get(0).([]model.Code), args.Error(1)
}

func (m *MockJobService) GetAllJobs() ([]model.Job, error) {
	args := m.Called()
	return args.Get(0).([]model.Job), args.Error(1)
}

func (m *MockJobService) GetJobs(nodeID, status string) ([]model.Job, error) {
	args := m.Called(nodeID, status)
	return args.Get(0).([]model.Job), args.Error(1)
}

func TestJobHandler_GetJobs_ByNodeID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockJobService)
	handler := NewJobHandler(mockService)

	nodeID := "node-001"
	jobName := "test-job"
	expectedJobs := []model.Job{
		{JobID: "job-001", NodeID: &nodeID, JobName: &jobName},
	}

	mockService.On("GetJobs", "node-001", "").Return(expectedJobs, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs?nodeId=node-001", nil)

	handler.GetJobs(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])

	mockService.AssertExpectations(t)
}

func TestJobHandler_GetJobs_ByStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockJobService)
	handler := NewJobHandler(mockService)

	status := "running"
	expectedJobs := []model.Job{
		{JobID: "job-001", Status: &status},
	}

	mockService.On("GetJobs", "", "running").Return(expectedJobs, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs?status=running", nil)

	handler.GetJobs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestJobHandler_GetJobs_NoParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockJobService)
	handler := NewJobHandler(mockService)

	expectedJobs := []model.Job{}
	mockService.On("GetJobs", "", "").Return(expectedJobs, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs", nil)

	handler.GetJobs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestJobHandler_GetJobByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockJobService)
	handler := NewJobHandler(mockService)

	jobName := "test-job"
	expectedJob := &model.Job{
		JobID:   "job-001",
		JobName: &jobName,
	}

	mockService.On("GetJobByID", "job-001").Return(expectedJob, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "jobId", Value: "job-001"}}
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs/job-001", nil)

	handler.GetJobByID(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestJobHandler_GetJobByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockJobService)
	handler := NewJobHandler(mockService)

	mockService.On("GetJobByID", "non-existent").Return(nil, gorm.ErrRecordNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "jobId", Value: "non-existent"}}
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs/non-existent", nil)

	handler.GetJobByID(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestJobHandler_GetJobParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockJobService)
	handler := NewJobHandler(mockService)

	jobID := "job-001"
	paramRaw := "learning_rate=0.001"
	expectedParams := []model.Parameter{
		{JobID: &jobID, ParameterRaw: &paramRaw},
	}

	mockService.On("GetJobParameters", "job-001").Return(expectedParams, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "jobId", Value: "job-001"}}
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs/job-001/parameters", nil)

	handler.GetJobParameters(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestJobHandler_GetJobCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockJobService)
	handler := NewJobHandler(mockService)

	jobID := "job-001"
	scriptPath := "/path/to/script.py"
	expectedCodes := []model.Code{
		{JobID: &jobID, ScriptPath: &scriptPath},
	}

	mockService.On("GetJobCode", "job-001").Return(expectedCodes, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "jobId", Value: "job-001"}}
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs/job-001/code", nil)

	handler.GetJobCode(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
