package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/task-monitor/api-server/internal/model"
)

// MockJobRepository is a mock implementation of JobRepository
type MockJobRepository struct {
	mock.Mock
}

func (m *MockJobRepository) Create(job *model.Job) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockJobRepository) Update(job *model.Job) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockJobRepository) FindByID(jobID string) (*model.Job, error) {
	args := m.Called(jobID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Job), args.Error(1)
}

func (m *MockJobRepository) FindByNodeID(nodeID string) ([]model.Job, error) {
	args := m.Called(nodeID)
	return args.Get(0).([]model.Job), args.Error(1)
}

func (m *MockJobRepository) FindByStatus(status string) ([]model.Job, error) {
	args := m.Called(status)
	return args.Get(0).([]model.Job), args.Error(1)
}

func (m *MockJobRepository) FindAll() ([]model.Job, error) {
	args := m.Called()
	return args.Get(0).([]model.Job), args.Error(1)
}

func (m *MockJobRepository) Find(nodeID, status string) ([]model.Job, error) {
	args := m.Called(nodeID, status)
	return args.Get(0).([]model.Job), args.Error(1)
}

func (m *MockJobRepository) UpdateStatus(jobID, status, reason string) error {
	args := m.Called(jobID, status, reason)
	return args.Error(0)
}

func (m *MockJobRepository) Upsert(job *model.Job) error {
	args := m.Called(job)
	return args.Error(0)
}

func (m *MockJobRepository) UpsertJobs(jobs []model.Job) error {
	args := m.Called(jobs)
	return args.Error(0)
}

func (m *MockJobRepository) MarkStaleJobsAsCompleted(nodeID string, activeJobIDs []string) error {
	args := m.Called(nodeID, activeJobIDs)
	return args.Error(0)
}

// MockParameterRepository is a mock implementation of ParameterRepository
type MockParameterRepository struct {
	mock.Mock
}

func (m *MockParameterRepository) Create(param *model.Parameter) error {
	args := m.Called(param)
	return args.Error(0)
}

func (m *MockParameterRepository) FindByJobID(jobID string) ([]model.Parameter, error) {
	args := m.Called(jobID)
	return args.Get(0).([]model.Parameter), args.Error(1)
}

func (m *MockParameterRepository) BatchCreate(params []model.Parameter) error {
	args := m.Called(params)
	return args.Error(0)
}

// MockCodeRepository is a mock implementation of CodeRepository
type MockCodeRepository struct {
	mock.Mock
}

func (m *MockCodeRepository) Create(code *model.Code) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *MockCodeRepository) FindByJobID(jobID string) ([]model.Code, error) {
	args := m.Called(jobID)
	return args.Get(0).([]model.Code), args.Error(1)
}

func (m *MockCodeRepository) BatchCreate(codes []model.Code) error {
	args := m.Called(codes)
	return args.Error(0)
}

// MockMetricsRepository is a mock implementation of MetricsRepository
type MockMetricsRepository struct {
	mock.Mock
}

// IsMetricsRepository 实现MetricsRepositoryInterface的标记方法
func (m *MockMetricsRepository) IsMetricsRepository() {}

func (m *MockMetricsRepository) CreateNPUMetric(metric *model.NPUMetric) error {
	args := m.Called(metric)
	return args.Error(0)
}

func (m *MockMetricsRepository) CreateProcessMetric(metric *model.ProcessMetric) error {
	args := m.Called(metric)
	return args.Error(0)
}

func (m *MockMetricsRepository) BatchCreateNPUMetrics(metrics []model.NPUMetric) error {
	args := m.Called(metrics)
	return args.Error(0)
}

func TestJobService_GetJobByID(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	service := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	jobName := "test-job"
	expectedJob := &model.Job{
		JobID:   "job-001",
		JobName: &jobName,
	}

	mockJobRepo.On("FindByID", "job-001").Return(expectedJob, nil)

	job, err := service.GetJobByID("job-001")

	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "job-001", job.JobID)
	assert.Equal(t, "test-job", *job.JobName)
	mockJobRepo.AssertExpectations(t)
}

func TestJobService_GetJobByID_NotFound(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	service := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	mockJobRepo.On("FindByID", "non-existent").Return(nil, errors.New("not found"))

	job, err := service.GetJobByID("non-existent")

	assert.Error(t, err)
	assert.Nil(t, job)
	mockJobRepo.AssertExpectations(t)
}

func TestJobService_GetJobsByNodeID(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	service := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	jobName := "test-job"
	expectedJobs := []model.Job{
		{JobID: "job-001", NodeID: &nodeID, JobName: &jobName},
	}

	mockJobRepo.On("FindByNodeID", "node-001").Return(expectedJobs, nil)

	jobs, err := service.GetJobsByNodeID("node-001")

	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "job-001", jobs[0].JobID)
	mockJobRepo.AssertExpectations(t)
}

func TestJobService_GetJobsByStatus(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	service := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	status := "running"
	expectedJobs := []model.Job{
		{JobID: "job-001", Status: &status},
	}

	mockJobRepo.On("FindByStatus", "running").Return(expectedJobs, nil)

	jobs, err := service.GetJobsByStatus("running")

	assert.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "running", *jobs[0].Status)
	mockJobRepo.AssertExpectations(t)
}

func TestJobService_GetJobParameters(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	service := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	jobID := "job-001"
	paramRaw := "learning_rate=0.001"
	expectedParams := []model.Parameter{
		{JobID: &jobID, ParameterRaw: &paramRaw},
	}

	mockParamRepo.On("FindByJobID", "job-001").Return(expectedParams, nil)

	params, err := service.GetJobParameters("job-001")

	assert.NoError(t, err)
	assert.Len(t, params, 1)
	assert.Equal(t, "learning_rate=0.001", *params[0].ParameterRaw)
	mockParamRepo.AssertExpectations(t)
}

func TestJobService_GetJobCode(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	service := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	jobID := "job-001"
	scriptPath := "/path/to/script.py"
	expectedCodes := []model.Code{
		{JobID: &jobID, ScriptPath: &scriptPath},
	}

	mockCodeRepo.On("FindByJobID", "job-001").Return(expectedCodes, nil)

	codes, err := service.GetJobCode("job-001")

	assert.NoError(t, err)
	assert.Len(t, codes, 1)
	assert.Equal(t, "/path/to/script.py", *codes[0].ScriptPath)
	mockCodeRepo.AssertExpectations(t)
}
