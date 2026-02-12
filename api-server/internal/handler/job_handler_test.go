package handler

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/task-monitor/api-server/internal/config"
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

func (m *MockJobService) GetJobDetail(jobID string, aggregate bool) (*service.JobDetailResponse, error) {
	args := m.Called(jobID, aggregate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.JobDetailResponse), args.Error(1)
}

func (m *MockJobService) GetJobStats() (map[string]int64, error) {
	args := m.Called()
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockJobService) UpdateJobFields(jobID string, fields map[string]interface{}) error {
	args := m.Called(jobID, fields)
	return args.Error(0)
}

func TestJobHandler_GetJobs_ByNodeID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockJobService)
	handler := NewJobHandler(mockService, nil)

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
	handler := NewJobHandler(mockService, nil)

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
	handler := NewJobHandler(mockService, nil)

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
	handler := NewJobHandler(mockService, nil)

	jobName := "test-job"
	expectedDetail := &service.JobDetailResponse{
		Job: model.Job{
			JobID:   "job-001",
			JobName: &jobName,
		},
		NPUCards:    []service.NPUCardInfo{},
		RelatedJobs: []model.Job{},
	}

	mockService.On("GetJobDetail", "job-001", true).Return(expectedDetail, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "jobId", Value: "job-001"}}
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs/job-001", nil)

	handler.GetJobByID(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	data := response["data"].(map[string]interface{})
	job := data["job"].(map[string]interface{})
	assert.Equal(t, "job-001", job["jobId"])
	mockService.AssertExpectations(t)
}

func TestJobHandler_GetJobByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := new(MockJobService)
	handler := NewJobHandler(mockService, nil)

	mockService.On("GetJobDetail", "non-existent", true).Return(nil, gorm.ErrRecordNotFound)

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
	handler := NewJobHandler(mockService, nil)

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
	handler := NewJobHandler(mockService, nil)

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
	handler := NewJobHandler(mockService, nil)

	nodeID := "node-001"
	jobName := "VLLM::Worker_TP0"
	cardCount := 2
	expectedGroups := []service.JobGroup{
		{
			MainJob:   model.Job{JobID: "job-001", NodeID: &nodeID, JobName: &jobName},
			ChildJobs: []model.Job{},
			CardCount: &cardCount,
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
	handler := NewJobHandler(mockService, nil)

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
	handler := NewJobHandler(mockService, nil)

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

// MockLLMService is a mock implementation of LLMServiceInterface
type MockLLMService struct {
	mock.Mock
}

func (m *MockLLMService) AnalyzeJob(jobID string) (*service.JobAnalysisResponse, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.JobAnalysisResponse), args.Error(1)
}

func (m *MockLLMService) AnalyzeJobWithModel(jobID, modelID string) (*service.JobAnalysisResponse, error) {
	args := m.Called(jobID, modelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.JobAnalysisResponse), args.Error(1)
}

func (m *MockLLMService) GetAnalysis(jobID string) (*service.JobAnalysisResponse, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.JobAnalysisResponse), args.Error(1)
}

func (m *MockLLMService) GetBatchAnalyses(jobIDs []string) (map[string]*service.JobAnalysisResponse, error) {
	args := m.Called(jobIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]*service.JobAnalysisResponse), args.Error(1)
}

func (m *MockLLMService) GetConfig() config.LLMConfig {
	args := m.Called()
	return args.Get(0).(config.LLMConfig)
}

func (m *MockLLMService) UpdateConfig(cfg config.LLMConfig) {
	m.Called(cfg)
}

func TestJobHandler_AnalyzeJob_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockJobService := new(MockJobService)
	mockLLMService := new(MockLLMService)
	handler := NewJobHandler(mockJobService, mockLLMService)

	expectedResult := &service.JobAnalysisResponse{
		Summary: "这是一个vLLM推理作业",
		TaskType: service.JobAnalysisTaskType{
			Category: "inference",
		},
		ResourceAssessment: service.JobAnalysisResourceAssessment{
			NpuUtilization: "high",
			HbmUtilization: "high",
			Description:    "资源利用率良好",
		},
		Issues: []service.JobAnalysisIssue{},
	}

	mockLLMService.On("AnalyzeJob", "job-001").Return(expectedResult, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "jobId", Value: "job-001"}}
	c.Request = httptest.NewRequest("POST", "/api/v1/jobs/job-001/analyze", nil)

	handler.AnalyzeJob(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	mockLLMService.AssertExpectations(t)
}

func TestJobHandler_AnalyzeJob_LLMNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockJobService := new(MockJobService)
	handler := NewJobHandler(mockJobService, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "jobId", Value: "job-001"}}
	c.Request = httptest.NewRequest("POST", "/api/v1/jobs/job-001/analyze", nil)

	handler.AnalyzeJob(c)

	assert.Equal(t, 501, w.Code)
}

func TestJobHandler_AnalyzeJob_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockJobService := new(MockJobService)
	mockLLMService := new(MockLLMService)
	handler := NewJobHandler(mockJobService, mockLLMService)

	mockLLMService.On("AnalyzeJob", "job-001").Return(nil, errors.New("LLM service error"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "jobId", Value: "job-001"}}
	c.Request = httptest.NewRequest("POST", "/api/v1/jobs/job-001/analyze", nil)

	handler.AnalyzeJob(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockLLMService.AssertExpectations(t)
}

func TestJobHandler_AnalyzeJob_WithCustomModel(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockJobService := new(MockJobService)
	mockLLMService := new(MockLLMService)
	handler := NewJobHandler(mockJobService, mockLLMService)

	resp := &service.JobAnalysisResponse{Summary: "ok"}
	mockLLMService.On("AnalyzeJobWithModel", "job-001", "qwen-max").Return(resp, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "jobId", Value: "job-001"}}
	c.Request = httptest.NewRequest("POST", "/api/v1/jobs/job-001/analyze", strings.NewReader(`{"modelId":"qwen-max"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.AnalyzeJob(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockLLMService.AssertExpectations(t)
}

func TestJobHandler_AnalyzeJob_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockJobService := new(MockJobService)
	mockLLMService := new(MockLLMService)
	handler := NewJobHandler(mockJobService, mockLLMService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "jobId", Value: "job-001"}}
	c.Request = httptest.NewRequest("POST", "/api/v1/jobs/job-001/analyze", strings.NewReader("{"))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.AnalyzeJob(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestJobHandler_ExportAnalysesCSV_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockJobService := new(MockJobService)
	mockLLMService := new(MockLLMService)
	handler := NewJobHandler(mockJobService, mockLLMService)

	jobName := "worker-1"
	nodeID := "node-001"
	status := "running"
	jobType := "inference"
	framework := "vllm"
	startTime := int64(1736039823000)
	cardCount := 2
	modelName := "qwen2.5"
	scriptPath := "/workspace/train.py"
	hbmUsage := 2048.0
	hbmTotal := 4096.0
	aiCore := 88.0

	groups := []service.JobGroup{
		{
			MainJob: model.Job{
				JobID:     "job-001",
				JobName:   &jobName,
				NodeID:    &nodeID,
				Status:    &status,
				JobType:   &jobType,
				Framework: &framework,
				StartTime: &startTime,
			},
			CardCount: &cardCount,
		},
	}

	statuses := []string{"running"}
	jobTypes := []string{"inference"}
	frameworks := []string{"vllm"}
	cardCounts := []int{2}
	mockJobService.On("GetGroupedJobs", "node-001", statuses, jobTypes, frameworks, cardCounts, "", "", 1, 100000).
		Return(groups, int64(1), nil)
	mockJobService.On("GetJobCode", "job-001").Return([]model.Code{{ScriptPath: &scriptPath}}, nil)
	mockJobService.On("GetJobDetail", "job-001", true).Return(&service.JobDetailResponse{
		NPUCards: []service.NPUCardInfo{
			{
				NpuID:         0,
				MemoryUsageMB: 1024,
				Metrics: []model.NPUMetric{{
					HBMUsageMB:         &hbmUsage,
					HBMTotalMB:         &hbmTotal,
					AICoreUsagePercent: &aiCore,
				}},
			},
		},
	}, nil)

	mockLLMService.On("GetBatchAnalyses", []string{"job-001"}).Return(map[string]*service.JobAnalysisResponse{
		"job-001": {
			Summary: "=cmd()",
			TaskType: service.JobAnalysisTaskType{
				Category: "inference",
			},
			ModelInfo:       &service.JobAnalysisModelInfo{ModelName: &modelName},
			RuntimeAnalysis: &service.JobAnalysisRuntimeAnalysis{Status: "normal"},
			ResourceAssessment: service.JobAnalysisResourceAssessment{
				NpuUtilization: "high",
				HbmUtilization: "medium",
			},
			Issues: []service.JobAnalysisIssue{{Category: "perf"}},
		},
	}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs/analyses/export?scope=filtered&nodeId=node-001&status=running&type=inference&framework=vllm&cardCount=2", nil)

	handler.ExportAnalysesCSV(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/csv")
	assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment; filename=\"ai-analysis-overview_")

	body := strings.TrimPrefix(w.Body.String(), "\uFEFF")
	rows, err := csv.NewReader(strings.NewReader(body)).ReadAll()
	assert.NoError(t, err)
	if assert.Len(t, rows, 2) {
		headerIdx := make(map[string]int, len(rows[0]))
		for i, name := range rows[0] {
			headerIdx[name] = i
		}
		assert.Equal(t, "job-001", rows[1][headerIdx["jobId"]])
		assert.Equal(t, "worker-1", rows[1][headerIdx["jobName"]])
		assert.Equal(t, scriptPath, rows[1][headerIdx["startupScript"]])
		assert.Equal(t, "1024.00", rows[1][headerIdx["processMemoryMb"]])
		assert.Equal(t, "2048.00", rows[1][headerIdx["hbmUsageMb"]])
		assert.Equal(t, "4096.00", rows[1][headerIdx["hbmTotalMb"]])
		assert.Equal(t, "50.00", rows[1][headerIdx["hbmUsagePercent"]])
		assert.Equal(t, "88.00", rows[1][headerIdx["aicoreUsagePercent"]])
		assert.Equal(t, byte(39), rows[1][headerIdx["summary"]][0])
		assert.Equal(t, "1", rows[1][headerIdx["issuesCount"]])
	}

	mockJobService.AssertExpectations(t)
	mockLLMService.AssertExpectations(t)
}

func TestJobHandler_ExportAnalysesCSV_SelectedWithoutIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockJobService := new(MockJobService)
	mockLLMService := new(MockLLMService)
	handler := NewJobHandler(mockJobService, mockLLMService)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs/analyses/export?scope=selected", nil)

	handler.ExportAnalysesCSV(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(400), response["code"])
	assert.Equal(t, "jobIds is required when scope=selected", response["message"])
}

func TestJobHandler_ExportAnalysesCSV_LLMNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockJobService := new(MockJobService)
	handler := NewJobHandler(mockJobService, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/jobs/analyses/export", nil)

	handler.ExportAnalysesCSV(c)

	assert.Equal(t, http.StatusNotImplemented, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(501), response["code"])
	assert.Equal(t, "LLM service is not configured", response["message"])
}
