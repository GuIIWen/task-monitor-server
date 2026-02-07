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

func (m *MockJobRepository) Find(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string, limit, offset int) ([]model.Job, error) {
	args := m.Called(nodeID, statuses, jobTypes, frameworks, sortBy, sortOrder, limit, offset)
	return args.Get(0).([]model.Job), args.Error(1)
}

func (m *MockJobRepository) Count(nodeID string, statuses []string, jobTypes []string, frameworks []string) (int64, error) {
	args := m.Called(nodeID, statuses, jobTypes, frameworks)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockJobRepository) FindFiltered(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string) ([]model.Job, error) {
	args := m.Called(nodeID, statuses, jobTypes, frameworks, sortBy, sortOrder)
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

func (m *MockMetricsRepository) FindNPUCardsByPIDs(nodeID string, pids []int64) (map[int64][]int, error) {
	args := m.Called(nodeID, pids)
	return args.Get(0).(map[int64][]int), args.Error(1)
}

func (m *MockMetricsRepository) DistinctNPUCardCounts() ([]int, error) {
	args := m.Called()
	return args.Get(0).([]int), args.Error(1)
}

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

func TestJobService_GetJobs(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	status := "running"
	jobName := "test-job"
	expectedJobs := []model.Job{
		{JobID: "job-001", NodeID: &nodeID, JobName: &jobName, Status: &status},
	}

	statuses := []string{"running"}
	var jobTypes, frameworks []string

	mockJobRepo.On("Count", nodeID, statuses, jobTypes, frameworks).Return(int64(25), nil)
	mockJobRepo.On("Find", nodeID, statuses, jobTypes, frameworks, "", "", 10, 10).Return(expectedJobs, nil)

	jobs, total, err := svc.GetJobs(nodeID, statuses, jobTypes, frameworks, "", "", 2, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(25), total)
	assert.Len(t, jobs, 1)
	assert.Equal(t, "job-001", jobs[0].JobID)
	mockJobRepo.AssertExpectations(t)
}

func TestJobService_GetGroupedJobs(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	// server.py (pid=100) 是父进程，EngineCore (pid=101, ppid=100) 是子进程
	pid1 := int64(100)
	pid2 := int64(101)
	ppid1 := int64(1) // server.py 的父进程是 init
	ppid2 := int64(100) // EngineCore 的父进程是 server.py
	startTime1 := int64(1770373780000)
	startTime2 := int64(1770373803000) // 晚 23 秒
	jobName1 := "server.py"
	jobName2 := "VLLM::EngineCore"
	status := "running"

	returnedJobs := []model.Job{
		{JobID: "job-001", NodeID: &nodeID, PID: &pid1, PPID: &ppid1, StartTime: &startTime1, JobName: &jobName1, Status: &status},
		{JobID: "job-002", NodeID: &nodeID, PID: &pid2, PPID: &ppid2, StartTime: &startTime2, JobName: &jobName2, Status: &status},
	}

	var statuses, jobTypes, frameworks []string
	var cardCounts []int

	mockJobRepo.On("FindFiltered", "", statuses, jobTypes, frameworks, "", "").Return(returnedJobs, nil)

	npuMap := map[int64][]int{
		100: {7},
		101: {7},
	}
	mockMetricsRepo.On("FindNPUCardsByPIDs", "node-001", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 2
	})).Return(npuMap, nil)

	groups, total, err := svc.GetGroupedJobs("", statuses, jobTypes, frameworks, cardCounts, "", "", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, groups, 1)
	// server.py 应该是 MainJob（它是根进程）
	assert.Equal(t, "job-001", groups[0].MainJob.JobID)
	assert.Len(t, groups[0].ChildJobs, 1)
	assert.Equal(t, "job-002", groups[0].ChildJobs[0].JobID)
	assert.NotNil(t, groups[0].CardCount)
	assert.Equal(t, 1, *groups[0].CardCount)
	mockJobRepo.AssertExpectations(t)
	mockMetricsRepo.AssertExpectations(t)
}

func TestJobService_GetGroupedJobs_MultipleGroups(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	// 两个独立进程，ppid 都不在集合中，各自成组
	pid1 := int64(100)
	pid2 := int64(200)
	ppid1 := int64(1)
	ppid2 := int64(1)
	startTime1 := int64(1770373780000)
	startTime2 := int64(1770373790000)
	jobName1 := "train.py"
	jobName2 := "infer.py"
	status := "running"

	returnedJobs := []model.Job{
		{JobID: "job-001", NodeID: &nodeID, PID: &pid1, PPID: &ppid1, StartTime: &startTime1, JobName: &jobName1, Status: &status},
		{JobID: "job-002", NodeID: &nodeID, PID: &pid2, PPID: &ppid2, StartTime: &startTime2, JobName: &jobName2, Status: &status},
	}

	var statuses, jobTypes, frameworks []string
	var cardCounts []int

	mockJobRepo.On("FindFiltered", "", statuses, jobTypes, frameworks, "", "").Return(returnedJobs, nil)

	npuMap := map[int64][]int{}
	mockMetricsRepo.On("FindNPUCardsByPIDs", "node-001", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 2
	})).Return(npuMap, nil)

	groups, total, err := svc.GetGroupedJobs("", statuses, jobTypes, frameworks, cardCounts, "", "", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, groups, 2)
	assert.Nil(t, groups[0].CardCount)
	assert.Nil(t, groups[1].CardCount)
	assert.Empty(t, groups[0].ChildJobs)
	assert.Empty(t, groups[1].ChildJobs)
	mockJobRepo.AssertExpectations(t)
	mockMetricsRepo.AssertExpectations(t)
}

func TestJobService_GetGroupedJobs_CardCountFilterBeforePagination(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	// 组A: pid=100 (ppid=1) 是父, pid=101 (ppid=100) 是子 → 通过 ppid 合并
	// 组B: pid=200 (ppid=1) 独立
	pid1 := int64(100)
	pid2 := int64(101)
	pid3 := int64(200)
	ppid1 := int64(1)
	ppid2 := int64(100)
	ppid3 := int64(1)
	startTime1 := int64(1770373780000)
	startTime2 := int64(1770373780000)
	startTime3 := int64(1770373790000)
	status := "running"
	jobName1 := "group-a-main"
	jobName2 := "group-a-child"
	jobName3 := "group-b-main"

	returnedJobs := []model.Job{
		{JobID: "job-001", NodeID: &nodeID, PID: &pid1, PPID: &ppid1, StartTime: &startTime1, JobName: &jobName1, Status: &status},
		{JobID: "job-002", NodeID: &nodeID, PID: &pid2, PPID: &ppid2, StartTime: &startTime2, JobName: &jobName2, Status: &status},
		{JobID: "job-003", NodeID: &nodeID, PID: &pid3, PPID: &ppid3, StartTime: &startTime3, JobName: &jobName3, Status: &status},
	}

	var statuses, jobTypes, frameworks []string
	cardCounts := []int{2}

	mockJobRepo.On("FindFiltered", "", statuses, jobTypes, frameworks, "", "").Return(returnedJobs, nil)

	npuMap := map[int64][]int{
		100: {0},
		101: {0},
		200: {0, 1},
	}
	mockMetricsRepo.On("FindNPUCardsByPIDs", "node-001", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 3
	})).Return(npuMap, nil)

	groups, total, err := svc.GetGroupedJobs("", statuses, jobTypes, frameworks, cardCounts, "", "", 1, 1)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, groups, 1)
	assert.Equal(t, "job-003", groups[0].MainJob.JobID)
	assert.NotNil(t, groups[0].CardCount)
	assert.Equal(t, 2, *groups[0].CardCount)
	mockJobRepo.AssertExpectations(t)
	mockMetricsRepo.AssertExpectations(t)
}

func TestJobService_GetGroupedJobs_CardCountFilterPageOutOfRange(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	pid := int64(100)
	ppid := int64(1)
	startTime := int64(1770373780000)
	status := "running"
	jobName := "group-a-main"

	returnedJobs := []model.Job{
		{JobID: "job-001", NodeID: &nodeID, PID: &pid, PPID: &ppid, StartTime: &startTime, JobName: &jobName, Status: &status},
	}

	var statuses, jobTypes, frameworks []string
	cardCounts := []int{1}

	mockJobRepo.On("FindFiltered", "", statuses, jobTypes, frameworks, "", "").Return(returnedJobs, nil)
	mockMetricsRepo.On("FindNPUCardsByPIDs", "node-001", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 1 && pids[0] == 100
	})).Return(map[int64][]int{100: {0}}, nil)

	groups, total, err := svc.GetGroupedJobs("", statuses, jobTypes, frameworks, cardCounts, "", "", 2, 1)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Empty(t, groups)
	mockJobRepo.AssertExpectations(t)
	mockMetricsRepo.AssertExpectations(t)
}

func TestJobService_GetDistinctCardCounts(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	mockMetricsRepo.On("DistinctNPUCardCounts").Return([]int{1, 2, 4, 8}, nil)

	counts, err := svc.GetDistinctCardCounts()

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 4, 8}, counts)
	mockMetricsRepo.AssertExpectations(t)
}

func TestJobService_GetGroupedJobs_PPIDChain(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	// 三层进程树: A(pid=100) → B(pid=200, ppid=100) → C(pid=300, ppid=200)
	pidA := int64(100)
	pidB := int64(200)
	pidC := int64(300)
	ppidA := int64(1)
	ppidB := int64(100)
	ppidC := int64(200)
	startA := int64(1770373780000)
	startB := int64(1770373781000)
	startC := int64(1770373782000)
	nameA := "server.py"
	nameB := "manager"
	nameC := "VLLM::Worker_TP0"
	status := "running"

	returnedJobs := []model.Job{
		{JobID: "job-A", NodeID: &nodeID, PID: &pidA, PPID: &ppidA, StartTime: &startA, JobName: &nameA, Status: &status},
		{JobID: "job-B", NodeID: &nodeID, PID: &pidB, PPID: &ppidB, StartTime: &startB, JobName: &nameB, Status: &status},
		{JobID: "job-C", NodeID: &nodeID, PID: &pidC, PPID: &ppidC, StartTime: &startC, JobName: &nameC, Status: &status},
	}

	var statuses, jobTypes, frameworks []string
	var cardCounts []int

	mockJobRepo.On("FindFiltered", "", statuses, jobTypes, frameworks, "", "").Return(returnedJobs, nil)

	npuMap := map[int64][]int{
		100: {0},
		200: {0},
		300: {1},
	}
	mockMetricsRepo.On("FindNPUCardsByPIDs", "node-001", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 3
	})).Return(npuMap, nil)

	groups, total, err := svc.GetGroupedJobs("", statuses, jobTypes, frameworks, cardCounts, "", "", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, groups, 1)
	// A 应该是 MainJob（根进程）
	assert.Equal(t, "job-A", groups[0].MainJob.JobID)
	assert.Len(t, groups[0].ChildJobs, 2)
	// 卡数=2（NPU 0 和 NPU 1）
	assert.NotNil(t, groups[0].CardCount)
	assert.Equal(t, 2, *groups[0].CardCount)
	mockJobRepo.AssertExpectations(t)
	mockMetricsRepo.AssertExpectations(t)
}
