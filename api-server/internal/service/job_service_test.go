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

func (m *MockJobRepository) FindByNodeIDAndPGID(nodeID string, pgid int64) ([]model.Job, error) {
	args := m.Called(nodeID, pgid)
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

func (m *MockMetricsRepository) FindNPUCardsByPIDsWithStatuses(nodeID string, pids []int64, statuses []string) (map[int64][]int, error) {
	args := m.Called(nodeID, pids, statuses)
	return args.Get(0).(map[int64][]int), args.Error(1)
}

func (m *MockMetricsRepository) DistinctNPUCardCounts() ([]int, error) {
	args := m.Called()
	return args.Get(0).([]int), args.Error(1)
}

func (m *MockMetricsRepository) FindNPUProcessesByPID(nodeID string, pid int64) ([]model.NPUProcess, error) {
	args := m.Called(nodeID, pid)
	return args.Get(0).([]model.NPUProcess), args.Error(1)
}

func (m *MockMetricsRepository) FindNPUProcessesByPIDs(nodeID string, pids []int64) ([]model.NPUProcess, error) {
	args := m.Called(nodeID, pids)
	return args.Get(0).([]model.NPUProcess), args.Error(1)
}

func (m *MockMetricsRepository) FindLatestNPUMetrics(nodeID string, npuIDs []int) ([]model.NPUMetric, error) {
	args := m.Called(nodeID, npuIDs)
	return args.Get(0).([]model.NPUMetric), args.Error(1)
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
	ppid1 := int64(1)   // server.py 的父进程是 init
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

	node1 := "node-001"
	node2 := "node-002"
	pid1 := int64(100)
	pid2 := int64(101)
	pid3 := int64(200)
	ppidRoot := int64(1)
	ppidChild := int64(100)
	start1 := int64(1770373780000)
	start2 := int64(1770373781000)
	start3 := int64(1770373790000)
	name1 := "train.py"
	name2 := "worker"
	name3 := "infer.py"
	status := "running"

	jobs := []model.Job{
		{JobID: "job-001", NodeID: &node1, PID: &pid1, PPID: &ppidRoot, StartTime: &start1, ProcessName: &name1, Status: &status},
		{JobID: "job-002", NodeID: &node1, PID: &pid2, PPID: &ppidChild, StartTime: &start2, ProcessName: &name2, Status: &status},
		{JobID: "job-003", NodeID: &node2, PID: &pid3, PPID: &ppidRoot, StartTime: &start3, ProcessName: &name3, Status: &status},
	}

	mockJobRepo.On("FindFiltered", "", []string(nil), []string(nil), []string(nil), "", "").Return(jobs, nil)
	mockMetricsRepo.On("FindNPUCardsByPIDs", "node-001", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 2
	})).Return(map[int64][]int{100: {0}, 101: {0}}, nil)
	mockMetricsRepo.On("FindNPUCardsByPIDs", "node-002", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 1 && pids[0] == 200
	})).Return(map[int64][]int{200: {0, 1}}, nil)

	counts, err := svc.GetDistinctCardCounts()

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2}, counts)
	mockJobRepo.AssertExpectations(t)
	mockMetricsRepo.AssertExpectations(t)
}

func TestJobService_GetGroupedJobs_NoCrossNodePIDMerge(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	node1 := "node-001"
	node2 := "node-002"
	pidRoot := int64(100)
	pidChild1 := int64(101)
	pidChild2 := int64(201)
	ppidRoot := int64(1)
	ppidChild := int64(100)
	start1 := int64(1770373780000)
	start2 := int64(1770373781000)
	start3 := int64(1770373790000)
	start4 := int64(1770373791000)
	nameRoot1 := "node1-main"
	nameChild1 := "node1-worker"
	nameRoot2 := "node2-main"
	nameChild2 := "node2-worker"
	status := "running"

	jobs := []model.Job{
		{JobID: "job-001", NodeID: &node1, PID: &pidRoot, PPID: &ppidRoot, StartTime: &start1, ProcessName: &nameRoot1, Status: &status},
		{JobID: "job-002", NodeID: &node1, PID: &pidChild1, PPID: &ppidChild, StartTime: &start2, ProcessName: &nameChild1, Status: &status},
		{JobID: "job-003", NodeID: &node2, PID: &pidRoot, PPID: &ppidRoot, StartTime: &start3, ProcessName: &nameRoot2, Status: &status},
		{JobID: "job-004", NodeID: &node2, PID: &pidChild2, PPID: &ppidChild, StartTime: &start4, ProcessName: &nameChild2, Status: &status},
	}

	var statuses, jobTypes, frameworks []string
	var cardCounts []int

	mockJobRepo.On("FindFiltered", "", statuses, jobTypes, frameworks, "", "").Return(jobs, nil)
	mockMetricsRepo.On("FindNPUCardsByPIDs", "node-001", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 2
	})).Return(map[int64][]int{100: {0}, 101: {0}}, nil)
	mockMetricsRepo.On("FindNPUCardsByPIDs", "node-002", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 2
	})).Return(map[int64][]int{100: {1}, 201: {1}}, nil)

	groups, total, err := svc.GetGroupedJobs("", statuses, jobTypes, frameworks, cardCounts, "", "", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, groups, 2)

	groupByNode := make(map[string]JobGroup)
	for _, g := range groups {
		if g.MainJob.NodeID != nil {
			groupByNode[*g.MainJob.NodeID] = g
		}
	}
	assert.Len(t, groupByNode, 2)
	assert.Equal(t, "job-001", groupByNode["node-001"].MainJob.JobID)
	assert.Equal(t, "job-003", groupByNode["node-002"].MainJob.JobID)
	assert.Len(t, groupByNode["node-001"].ChildJobs, 1)
	assert.Len(t, groupByNode["node-002"].ChildJobs, 1)

	mockJobRepo.AssertExpectations(t)
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

func TestJobService_GetGroupedJobs_ChildJobsFilteredByNPU(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	// train.py (pid=100) 是主进程
	// python3.1 (pid=101, ppid=100) 在 NPU 上运行
	// python3.1 (pid=102, ppid=100) 在 NPU 上运行
	// pt_data_worker (pid=103, ppid=100) 不在 NPU 上
	// pt_data_worker (pid=104, ppid=100) 不在 NPU 上
	pid1 := int64(100)
	pid2 := int64(101)
	pid3 := int64(102)
	pid4 := int64(103)
	pid5 := int64(104)
	ppid1 := int64(1)
	ppid2 := int64(100)
	ppid3 := int64(100)
	ppid4 := int64(100)
	ppid5 := int64(100)
	startTime1 := int64(1770373780000)
	startTime2 := int64(1770373781000)
	startTime3 := int64(1770373781000)
	startTime4 := int64(1770373782000)
	startTime5 := int64(1770373782000)
	name1 := "train.py"
	name2 := "python3.1"
	name3 := "python3.1"
	name4 := "pt_data_worker"
	name5 := "pt_data_worker"
	status := "running"

	returnedJobs := []model.Job{
		{JobID: "job-001", NodeID: &nodeID, PID: &pid1, PPID: &ppid1, StartTime: &startTime1, JobName: &name1, ProcessName: &name1, Status: &status},
		{JobID: "job-002", NodeID: &nodeID, PID: &pid2, PPID: &ppid2, StartTime: &startTime2, JobName: &name2, ProcessName: &name2, Status: &status},
		{JobID: "job-003", NodeID: &nodeID, PID: &pid3, PPID: &ppid3, StartTime: &startTime3, JobName: &name3, ProcessName: &name3, Status: &status},
		{JobID: "job-004", NodeID: &nodeID, PID: &pid4, PPID: &ppid4, StartTime: &startTime4, JobName: &name4, ProcessName: &name4, Status: &status},
		{JobID: "job-005", NodeID: &nodeID, PID: &pid5, PPID: &ppid5, StartTime: &startTime5, JobName: &name5, ProcessName: &name5, Status: &status},
	}

	var statuses, jobTypes, frameworks []string
	var cardCounts []int

	mockJobRepo.On("FindFiltered", "", statuses, jobTypes, frameworks, "", "").Return(returnedJobs, nil)

	// 只有 pid=100,101,102 在 NPU 上，103,104 不在
	npuMap := map[int64][]int{
		100: {0, 1},
		101: {0},
		102: {1},
	}
	mockMetricsRepo.On("FindNPUCardsByPIDs", "node-001", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 5
	})).Return(npuMap, nil)

	groups, total, err := svc.GetGroupedJobs("", statuses, jobTypes, frameworks, cardCounts, "", "", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, groups, 1)
	// train.py 是 MainJob
	assert.Equal(t, "job-001", groups[0].MainJob.JobID)
	// childJobs 只包含 NPU 进程（python3.1），不包含 pt_data_worker
	assert.Len(t, groups[0].ChildJobs, 2)
	childIDs := make(map[string]bool)
	for _, child := range groups[0].ChildJobs {
		childIDs[child.JobID] = true
	}
	assert.True(t, childIDs["job-002"], "NPU child job-002 should be included")
	assert.True(t, childIDs["job-003"], "NPU child job-003 should be included")
	assert.False(t, childIDs["job-004"], "non-NPU child job-004 should be excluded")
	assert.False(t, childIDs["job-005"], "non-NPU child job-005 should be excluded")
	// 卡数=2（NPU 0 和 NPU 1）
	assert.NotNil(t, groups[0].CardCount)
	assert.Equal(t, 2, *groups[0].CardCount)
	mockJobRepo.AssertExpectations(t)
	mockMetricsRepo.AssertExpectations(t)
}

func TestJobService_GetGroupedJobs_StoppedFallbackToHistoricalNPU(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	pidMain := int64(100)
	pidChild := int64(101)
	ppidRoot := int64(1)
	ppidChild := int64(100)
	startMain := int64(1770373780000)
	startChild := int64(1770373781000)
	nameMain := "python"
	nameChild := "VLLM::Worker_TP0"
	statusStopped := "stopped"

	returnedJobs := []model.Job{
		{JobID: "job-main", NodeID: &nodeID, PID: &pidMain, PPID: &ppidRoot, StartTime: &startMain, JobName: &nameMain, ProcessName: &nameMain, Status: &statusStopped},
		{JobID: "job-child", NodeID: &nodeID, PID: &pidChild, PPID: &ppidChild, StartTime: &startChild, JobName: &nameChild, ProcessName: &nameChild, Status: &statusStopped},
	}

	var statuses, jobTypes, frameworks []string
	var cardCounts []int

	mockJobRepo.On("FindFiltered", "", statuses, jobTypes, frameworks, "", "").Return(returnedJobs, nil)
	mockMetricsRepo.On("FindNPUCardsByPIDs", "node-001", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 2
	})).Return(map[int64][]int{}, nil)
	mockMetricsRepo.On("FindNPUCardsByPIDsWithStatuses", "node-001", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 2
	}), []string{"running", "stopped"}).Return(map[int64][]int{
		100: {0, 1},
		101: {0, 1},
	}, nil)

	groups, total, err := svc.GetGroupedJobs("", statuses, jobTypes, frameworks, cardCounts, "", "", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, groups, 1)
	assert.Equal(t, "job-main", groups[0].MainJob.JobID)
	assert.Len(t, groups[0].ChildJobs, 1)
	assert.Equal(t, "job-child", groups[0].ChildJobs[0].JobID)
	assert.NotNil(t, groups[0].CardCount)
	assert.Equal(t, 2, *groups[0].CardCount)
	mockJobRepo.AssertExpectations(t)
	mockMetricsRepo.AssertExpectations(t)
}

func TestJobService_GetGroupedJobs_StoppedIncludeChildrenWhenNoNPUData(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	pidMain := int64(100)
	pidChild := int64(101)
	ppidRoot := int64(1)
	ppidChild := int64(100)
	startMain := int64(1770373780000)
	startChild := int64(1770373781000)
	nameMain := "python"
	nameChild := "python"
	statusStopped := "stopped"

	returnedJobs := []model.Job{
		{JobID: "job-main", NodeID: &nodeID, PID: &pidMain, PPID: &ppidRoot, StartTime: &startMain, JobName: &nameMain, ProcessName: &nameMain, Status: &statusStopped},
		{JobID: "job-child", NodeID: &nodeID, PID: &pidChild, PPID: &ppidChild, StartTime: &startChild, JobName: &nameChild, ProcessName: &nameChild, Status: &statusStopped},
	}

	var statuses, jobTypes, frameworks []string
	var cardCounts []int

	mockJobRepo.On("FindFiltered", "", statuses, jobTypes, frameworks, "", "").Return(returnedJobs, nil)
	mockMetricsRepo.On("FindNPUCardsByPIDs", "node-001", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 2
	})).Return(map[int64][]int{}, nil)
	mockMetricsRepo.On("FindNPUCardsByPIDsWithStatuses", "node-001", mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 2
	}), []string{"running", "stopped"}).Return(map[int64][]int{}, nil)

	groups, total, err := svc.GetGroupedJobs("", statuses, jobTypes, frameworks, cardCounts, "", "", 1, 20)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, groups, 1)
	assert.Equal(t, "job-main", groups[0].MainJob.JobID)
	assert.Len(t, groups[0].ChildJobs, 1)
	assert.Equal(t, "job-child", groups[0].ChildJobs[0].JobID)
	assert.Nil(t, groups[0].CardCount)
	mockJobRepo.AssertExpectations(t)
	mockMetricsRepo.AssertExpectations(t)
}

func TestJobService_GetJobDetail_WithNPU(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	pid := int64(100)
	ppid := int64(1)
	pgid := int64(100)
	jobName := "train.py"
	status := "running"
	startTime := int64(1770373780000)

	job := &model.Job{
		JobID:     "job-001",
		NodeID:    &nodeID,
		PID:       &pid,
		PPID:      &ppid,
		PGID:      &pgid,
		JobName:   &jobName,
		Status:    &status,
		StartTime: &startTime,
	}

	mockJobRepo.On("FindByID", "job-001").Return(job, nil)

	// NPU processes for this pid
	npuID0 := 0
	npuID1 := 1
	mem0 := float64(1024.5)
	mem1 := float64(2048.0)
	npuProcs := []model.NPUProcess{
		{NodeID: &nodeID, PID: &pid, NPUID: &npuID0, MemoryUsageMB: &mem0},
		{NodeID: &nodeID, PID: &pid, NPUID: &npuID1, MemoryUsageMB: &mem1},
	}
	mockMetricsRepo.On("FindNPUProcessesByPID", nodeID, pid).Return(npuProcs, nil)

	// Latest NPU metrics
	powerW := float64(150.0)
	tempC := float64(65.0)
	metrics := []model.NPUMetric{
		{NPUID: &npuID0, PowerW: &powerW, TempC: &tempC},
		{NPUID: &npuID1, PowerW: &powerW, TempC: &tempC},
	}
	mockMetricsRepo.On("FindLatestNPUMetrics", nodeID, mock.MatchedBy(func(ids []int) bool {
		return len(ids) == 2
	})).Return(metrics, nil)

	// Same pgid jobs for related processes
	pid2 := int64(101)
	ppid2 := int64(100)
	childName := "worker"
	samePGIDJobs := []model.Job{
		*job,
		{JobID: "job-002", NodeID: &nodeID, PID: &pid2, PPID: &ppid2, PGID: &pgid, JobName: &childName, Status: &status, StartTime: &startTime},
	}
	mockJobRepo.On("FindByNodeIDAndPGID", nodeID, pgid).Return(samePGIDJobs, nil)

	// NPU cards for related pids
	npuMap := map[int64][]int{101: {0}}
	mockMetricsRepo.On("FindNPUCardsByPIDs", nodeID, mock.MatchedBy(func(pids []int64) bool {
		return len(pids) == 1 && pids[0] == 101
	})).Return(npuMap, nil)

	detail, err := svc.GetJobDetail("job-001")

	assert.NoError(t, err)
	assert.NotNil(t, detail)
	assert.Equal(t, "job-001", detail.Job.JobID)
	assert.Len(t, detail.NPUCards, 2)
	assert.Len(t, detail.RelatedJobs, 1)
	assert.Equal(t, "job-002", detail.RelatedJobs[0].JobID)
	mockJobRepo.AssertExpectations(t)
	mockMetricsRepo.AssertExpectations(t)
}

func TestJobService_GetJobDetail_DeduplicateNPUCards(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	pid := int64(100)
	jobName := "train.py"

	job := &model.Job{
		JobID:   "job-001",
		NodeID:  &nodeID,
		PID:     &pid,
		JobName: &jobName,
	}

	mockJobRepo.On("FindByID", "job-001").Return(job, nil)

	npuID0 := 0
	npuID1 := 1
	mem0 := float64(1024)
	mem0dup := float64(2048)
	mem1 := float64(512)
	npuProcs := []model.NPUProcess{
		{NodeID: &nodeID, PID: &pid, NPUID: &npuID0, MemoryUsageMB: &mem0},
		{NodeID: &nodeID, PID: &pid, NPUID: &npuID0, MemoryUsageMB: &mem0dup},
		{NodeID: &nodeID, PID: &pid, NPUID: &npuID1, MemoryUsageMB: &mem1},
	}
	mockMetricsRepo.On("FindNPUProcessesByPID", nodeID, pid).Return(npuProcs, nil)

	metrics := []model.NPUMetric{
		{NPUID: &npuID0},
		{NPUID: &npuID1},
	}
	mockMetricsRepo.On("FindLatestNPUMetrics", nodeID, mock.MatchedBy(func(ids []int) bool {
		return len(ids) == 2
	})).Return(metrics, nil)

	detail, err := svc.GetJobDetail("job-001")

	assert.NoError(t, err)
	assert.NotNil(t, detail)
	assert.Len(t, detail.NPUCards, 2)
	assert.Equal(t, 0, detail.NPUCards[0].NpuID)
	assert.Equal(t, float64(2048), detail.NPUCards[0].MemoryUsageMB)
	assert.Equal(t, 1, detail.NPUCards[1].NpuID)
	assert.Equal(t, float64(512), detail.NPUCards[1].MemoryUsageMB)
	mockJobRepo.AssertExpectations(t)
	mockMetricsRepo.AssertExpectations(t)
}

func TestJobService_GetJobDetail_NoNPU(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	nodeID := "node-001"
	pid := int64(100)
	jobName := "simple-job"

	job := &model.Job{
		JobID:   "job-001",
		NodeID:  &nodeID,
		PID:     &pid,
		JobName: &jobName,
	}

	mockJobRepo.On("FindByID", "job-001").Return(job, nil)
	mockMetricsRepo.On("FindNPUProcessesByPID", nodeID, pid).Return([]model.NPUProcess{}, nil)

	detail, err := svc.GetJobDetail("job-001")

	assert.NoError(t, err)
	assert.NotNil(t, detail)
	assert.Equal(t, "job-001", detail.Job.JobID)
	assert.Empty(t, detail.NPUCards)
	assert.Empty(t, detail.RelatedJobs)
	mockJobRepo.AssertExpectations(t)
	mockMetricsRepo.AssertExpectations(t)
}

func TestJobService_GetJobDetail_NotFound(t *testing.T) {
	mockJobRepo := new(MockJobRepository)
	mockParamRepo := new(MockParameterRepository)
	mockCodeRepo := new(MockCodeRepository)
	mockMetricsRepo := new(MockMetricsRepository)

	svc := NewJobService(mockJobRepo, mockParamRepo, mockCodeRepo, mockMetricsRepo)

	mockJobRepo.On("FindByID", "non-existent").Return(nil, errors.New("record not found"))

	detail, err := svc.GetJobDetail("non-existent")

	assert.Error(t, err)
	assert.Nil(t, detail)
	mockJobRepo.AssertExpectations(t)
}
