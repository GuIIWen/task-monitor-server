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
	"github.com/task-monitor/api-server/internal/service"
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

func (m *MockJobService) GetJobs(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string, page, pageSize int) ([]model.Job, int64, error) {
	args := m.Called(nodeID, statuses, jobTypes, frameworks, sortBy, sortOrder, page, pageSize)
	return args.Get(0).([]model.Job), args.Get(1).(int64), args.Error(2)
}

func (m *MockJobService) GetGroupedJobs(nodeID string, statuses []string, jobTypes []string, frameworks []string, cardCounts []int, sortBy, sortOrder string, page, pageSize int) ([]service.JobGroup, int64, error) {
	args := m.Called(nodeID, statuses, jobTypes, frameworks, cardCounts, sortBy, sortOrder, page, pageSize)
	return args.Get(0).([]service.JobGroup), args.Get(1).(int64), args.Error(2)
}

func (m *MockJobService) GetDistinctCardCounts() ([]int, error) {
	args := m.Called()
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockJobService) GetJobStats() (map[string]int64, error) {
	args := m.Called()
	return args.Get(0).(map[string]int64), args.Error(1)
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

	var statuses, jobTypes, frameworks []string
	mockService.On("GetJobs", "node-001", statuses, jobTypes, frameworks, "", "", 1, 20).Return(expectedJobs, int64(1), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs?nodeId=node-001", nil)

	handler.GetJobs(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	data := response["data"].(map[string]interface{})
	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(20), pagination["pageSize"])
	assert.Equal(t, float64(1), pagination["total"])

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

	statuses := []string{"running"}
	var jobTypes, frameworks []string
	mockService.On("GetJobs", "", statuses, jobTypes, frameworks, "", "", 1, 20).Return(expectedJobs, int64(1), nil)

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
	var statuses2, jobTypes2, frameworks2 []string
	mockService.On("GetJobs", "", statuses2, jobTypes2, frameworks2, "", "", 1, 20).Return(expectedJobs, int64(0), nil)

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

func TestJobHandler_GetGroupedJobs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockJobService)
	handler := NewJobHandler(mockService)

	nodeID := "node-001"
	jobName := "VLLM::Worker_TP0"
	expectedGroups := []service.JobGroup{
		{
			MainJob:   model.Job{JobID: "job-001", NodeID: &nodeID, JobName: &jobName},
			ChildJobs: []model.Job{},
			CardCount: 2,
		},
	}

	var statuses, jobTypes, frameworks []string
	var cardCounts []int
	mockService.On("GetGroupedJobs", "node-001", statuses, jobTypes, frameworks, cardCounts, "", "", 1, 20).
		Return(expectedGroups, int64(1), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs/grouped?nodeId=node-001", nil)

	handler.GetGroupedJobs(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	data := response["data"].(map[string]interface{})
	pagination := data["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["total"])
	mockService.AssertExpectations(t)
}

func TestJobHandler_GetGroupedJobs_WithCardCountFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockJobService)
	handler := NewJobHandler(mockService)

	expectedGroups := []service.JobGroup{}

	var statuses, jobTypes, frameworks []string
	cardCounts := []int{4}
	mockService.On("GetGroupedJobs", "", statuses, jobTypes, frameworks, cardCounts, "", "", 1, 20).
		Return(expectedGroups, int64(0), nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs/grouped?cardCount=4", nil)

	handler.GetGroupedJobs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestJobHandler_GetDistinctCardCounts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockJobService)
	handler := NewJobHandler(mockService)

	mockService.On("GetDistinctCardCounts").Return([]int{1, 2, 4, 8, 16}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs/grouped/card-counts", nil)

	handler.GetDistinctCardCounts(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	data := response["data"].([]interface{})
	assert.Len(t, data, 5)
	mockService.AssertExpectations(t)
}
