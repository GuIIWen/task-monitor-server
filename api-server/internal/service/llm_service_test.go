package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/task-monitor/api-server/internal/config"
	"github.com/task-monitor/api-server/internal/model"
)

// MockJobServiceForLLM implements JobServiceInterface for LLM tests
type MockJobServiceForLLM struct {
	mock.Mock
}

func (m *MockJobServiceForLLM) GetJobByID(jobID string) (*model.Job, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Job), args.Error(1)
}

func (m *MockJobServiceForLLM) GetJobDetail(jobID string) (*JobDetailResponse, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*JobDetailResponse), args.Error(1)
}

func (m *MockJobServiceForLLM) GetJobsByNodeID(nodeID string) ([]model.Job, error) {
	args := m.Called(nodeID)
	return args.Get(0).([]model.Job), args.Error(1)
}

func (m *MockJobServiceForLLM) GetJobsByStatus(status string) ([]model.Job, error) {
	args := m.Called(status)
	return args.Get(0).([]model.Job), args.Error(1)
}

func (m *MockJobServiceForLLM) GetAllJobs() ([]model.Job, error) {
	args := m.Called()
	return args.Get(0).([]model.Job), args.Error(1)
}

func (m *MockJobServiceForLLM) GetJobs(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string, page, pageSize int) ([]model.Job, int64, error) {
	args := m.Called(nodeID, statuses, jobTypes, frameworks, sortBy, sortOrder, page, pageSize)
	return args.Get(0).([]model.Job), args.Get(1).(int64), args.Error(2)
}

func (m *MockJobServiceForLLM) GetGroupedJobs(nodeID string, statuses []string, jobTypes []string, frameworks []string, cardCounts []int, sortBy, sortOrder string, page, pageSize int) ([]JobGroup, int64, error) {
	args := m.Called(nodeID, statuses, jobTypes, frameworks, cardCounts, sortBy, sortOrder, page, pageSize)
	return args.Get(0).([]JobGroup), args.Get(1).(int64), args.Error(2)
}

func (m *MockJobServiceForLLM) GetDistinctCardCounts() ([]int, error) {
	args := m.Called()
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockJobServiceForLLM) GetJobParameters(jobID string) ([]model.Parameter, error) {
	args := m.Called(jobID)
	return args.Get(0).([]model.Parameter), args.Error(1)
}

func (m *MockJobServiceForLLM) GetJobCode(jobID string) ([]model.Code, error) {
	args := m.Called(jobID)
	return args.Get(0).([]model.Code), args.Error(1)
}

func (m *MockJobServiceForLLM) GetJobStats() (map[string]int64, error) {
	args := m.Called()
	return args.Get(0).(map[string]int64), args.Error(1)
}

func TestLLMService_AnalyzeJob_Disabled(t *testing.T) {
	mockJobSvc := new(MockJobServiceForLLM)
	cfg := config.LLMConfig{Enabled: false}
	svc := NewLLMService(mockJobSvc, cfg)

	result, err := svc.AnalyzeJob("job-001")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not enabled")
}

func TestLLMService_AnalyzeJob_Success(t *testing.T) {
	// 创建模拟LLM服务器
	analysisResult := JobAnalysisResponse{
		Summary: "vLLM推理服务，使用Qwen2.5-7B模型",
		TaskType: JobAnalysisTaskType{
			Category: "inference",
		},
		ResourceAssessment: JobAnalysisResourceAssessment{
			NpuUtilization: "high",
			HbmUtilization: "medium",
			Description:    "NPU利用率良好",
		},
		Issues:      []JobAnalysisIssue{},
	}
	resultJSON, _ := json.Marshal(analysisResult)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": string(resultJSON)}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// 准备mock数据
	mockJobSvc := new(MockJobServiceForLLM)
	setupMockJobData(mockJobSvc, "job-001")

	cfg := config.LLMConfig{
		Enabled:  true,
		Endpoint: server.URL + "/v1",
		APIKey:   "test-key",
		Model:    "test-model",
		Timeout:  10,
	}
	svc := NewLLMService(mockJobSvc, cfg)

	result, err := svc.AnalyzeJob("job-001")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "vLLM推理服务，使用Qwen2.5-7B模型", result.Summary)
	assert.Equal(t, "inference", result.TaskType.Category)
	mockJobSvc.AssertExpectations(t)
}

func TestLLMService_AnalyzeJob_LLMError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer server.Close()

	mockJobSvc := new(MockJobServiceForLLM)
	setupMockJobData(mockJobSvc, "job-001")

	cfg := config.LLMConfig{
		Enabled:  true,
		Endpoint: server.URL + "/v1",
		Model:    "test-model",
		Timeout:  10,
	}
	svc := NewLLMService(mockJobSvc, cfg)

	result, err := svc.AnalyzeJob("job-001")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "status 500")
}

func TestLLMService_AnalyzeJob_MarkdownWrappedJSON(t *testing.T) {
	analysisResult := JobAnalysisResponse{
		Summary: "训练作业",
		TaskType: JobAnalysisTaskType{
			Category: "training",
		},
		ResourceAssessment: JobAnalysisResourceAssessment{
			NpuUtilization: "medium",
			HbmUtilization: "high",
			Description:    "HBM使用率较高",
		},
		Issues:      []JobAnalysisIssue{},
	}
	resultJSON, _ := json.Marshal(analysisResult)
	// LLM返回markdown包裹的JSON
	wrappedContent := "```json\n" + string(resultJSON) + "\n```"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": wrappedContent}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	mockJobSvc := new(MockJobServiceForLLM)
	setupMockJobData(mockJobSvc, "job-001")

	cfg := config.LLMConfig{
		Enabled:  true,
		Endpoint: server.URL + "/v1",
		Model:    "test-model",
		Timeout:  10,
	}
	svc := NewLLMService(mockJobSvc, cfg)

	result, err := svc.AnalyzeJob("job-001")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "训练作业", result.Summary)
	assert.Equal(t, "training", result.TaskType.Category)
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain JSON",
			input: `{"summary":"test"}`,
			want:  `{"summary":"test"}`,
		},
		{
			name:  "markdown json block",
			input: "```json\n{\"summary\":\"test\"}\n```",
			want:  `{"summary":"test"}`,
		},
		{
			name:  "markdown plain block",
			input: "```\n{\"summary\":\"test\"}\n```",
			want:  `{"summary":"test"}`,
		},
		{
			name:  "text with JSON",
			input: "Here is the result: {\"summary\":\"test\"} done",
			want:  `{"summary":"test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSON(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTruncateStr(t *testing.T) {
	assert.Equal(t, "hello", truncateStr("hello", 10))
	assert.Contains(t, truncateStr("hello world this is long", 5), "hello")
	assert.Contains(t, truncateStr("hello world this is long", 5), "截断")
}

func TestFilterRelevantEnvVars(t *testing.T) {
	envJSON := `{"PATH":"/usr/bin","DB_PASSWORD":"secret123","CUDA_VISIBLE_DEVICES":"0,1","MASTER_ADDR":"localhost","NORMAL_VAR":"value"}`
	result := filterRelevantEnvVars(envJSON)
	assert.NotContains(t, result, "PATH=/usr/bin")
	assert.NotContains(t, result, "NORMAL_VAR")
	assert.Contains(t, result, "CUDA_VISIBLE_DEVICES=0,1")
	assert.Contains(t, result, "MASTER_ADDR=localhost")
	assert.NotContains(t, result, "secret123")
}

func TestLLMService_GetConfig_MasksAPIKey(t *testing.T) {
	mockJobSvc := new(MockJobServiceForLLM)
	cfg := config.LLMConfig{
		Enabled:  true,
		Endpoint: "http://localhost:8000/v1",
		APIKey:   "sk-abcdef123456",
		Model:    "qwen2.5",
		Timeout:  60,
	}
	svc := NewLLMService(mockJobSvc, cfg)

	result := svc.GetConfig()
	assert.Equal(t, "****3456", result.APIKey)
	assert.Equal(t, true, result.Enabled)
	assert.Equal(t, "qwen2.5", result.Model)
}

func TestLLMService_GetConfig_ShortAPIKey(t *testing.T) {
	mockJobSvc := new(MockJobServiceForLLM)
	cfg := config.LLMConfig{APIKey: "ab"}
	svc := NewLLMService(mockJobSvc, cfg)

	result := svc.GetConfig()
	assert.Equal(t, "****", result.APIKey)
}

func TestLLMService_GetConfig_EmptyAPIKey(t *testing.T) {
	mockJobSvc := new(MockJobServiceForLLM)
	cfg := config.LLMConfig{APIKey: ""}
	svc := NewLLMService(mockJobSvc, cfg)

	result := svc.GetConfig()
	assert.Equal(t, "", result.APIKey)
}

func TestLLMService_UpdateConfig(t *testing.T) {
	mockJobSvc := new(MockJobServiceForLLM)
	cfg := config.LLMConfig{
		Enabled: false,
		Model:   "old-model",
		Timeout: 30,
	}
	svc := NewLLMService(mockJobSvc, cfg)

	newCfg := config.LLMConfig{
		Enabled:  true,
		Endpoint: "http://new:8000/v1",
		APIKey:   "new-key",
		Model:    "new-model",
		Timeout:  120,
	}
	svc.UpdateConfig(newCfg)

	result := svc.GetConfig()
	assert.Equal(t, true, result.Enabled)
	assert.Equal(t, "http://new:8000/v1", result.Endpoint)
	assert.Equal(t, "new-model", result.Model)
	assert.Equal(t, 120, result.Timeout)
}

// setupMockJobData 设置mock作业数据
func setupMockJobData(mockJobSvc *MockJobServiceForLLM, jobID string) {
	nodeID := "node-001"
	jobName := "vllm-server"
	framework := "vLLM"
	status := "running"
	cmdLine := "python -m vllm.entrypoints.openai.api_server --model Qwen2.5-7B"
	pid := int64(100)

	aicore := float64(85.0)
	hbmUsed := float64(12000.0)
	hbmTotal := float64(16000.0)

	detail := &JobDetailResponse{
		Job: model.Job{
			JobID:       jobID,
			NodeID:      &nodeID,
			JobName:     &jobName,
			Framework:   &framework,
			Status:      &status,
			CommandLine: &cmdLine,
			PID:         &pid,
		},
		NPUCards: []NPUCardInfo{
			{
				NpuID:         0,
				MemoryUsageMB: 11500.0,
				Metrics: []model.NPUMetric{
					{
						AICoreUsagePercent: &aicore,
						HBMUsageMB:        &hbmUsed,
						HBMTotalMB:        &hbmTotal,
					},
				},
			},
		},
		RelatedJobs: []model.Job{},
	}
	mockJobSvc.On("GetJobDetail", jobID).Return(detail, nil)

	paramData := `{"model":"Qwen2.5-7B","tensor_parallel_size":"1"}`
	envVars := `{"PATH":"/usr/bin","CUDA_VISIBLE_DEVICES":"0"}`
	params := []model.Parameter{
		{ParameterData: &paramData, EnvVars: &envVars},
	}
	mockJobSvc.On("GetJobParameters", jobID).Return(params, nil)

	scriptContent := "from vllm import LLM\nllm = LLM(model='Qwen2.5-7B')"
	scriptPath := "/workspace/serve.py"
	codes := []model.Code{
		{ScriptContent: &scriptContent, ScriptPath: &scriptPath},
	}
	mockJobSvc.On("GetJobCode", jobID).Return(codes, nil)
}
