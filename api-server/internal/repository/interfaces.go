package repository

import "github.com/task-monitor/api-server/internal/model"

// NodeRepositoryInterface defines the interface for node repository operations
// API Server只需要查询功能，不需要写入功能
type NodeRepositoryInterface interface {
	FindByID(nodeID string) (*model.Node, error)
	FindAll() ([]model.Node, error)
	FindByStatus(status string) ([]model.Node, error)
}

// JobRepositoryInterface defines the interface for job repository operations
// API Server只需要查询功能，不需要写入功能
type JobRepositoryInterface interface {
	FindByID(jobID string) (*model.Job, error)
	FindByNodeID(nodeID string) ([]model.Job, error)
	FindByNodeIDAndPGID(nodeID string, pgid int64) ([]model.Job, error)
	FindByStatus(status string) ([]model.Job, error)
	FindAll() ([]model.Job, error)
	Find(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string, limit, offset int) ([]model.Job, error)
	Count(nodeID string, statuses []string, jobTypes []string, frameworks []string) (int64, error)
	FindFiltered(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string) ([]model.Job, error)
	UpdateFields(jobID string, fields map[string]interface{}) error
}

// ParameterRepositoryInterface defines the interface for parameter repository operations
// API Server只需要查询功能，不需要写入功能
type ParameterRepositoryInterface interface {
	FindByJobID(jobID string) ([]model.Parameter, error)
}

// CodeRepositoryInterface defines the interface for code repository operations
// API Server只需要查询功能，不需要写入功能
type CodeRepositoryInterface interface {
	FindByJobID(jobID string) ([]model.Code, error)
}

// MetricsRepositoryInterface defines the interface for metrics repository operations
// API Server只需要查询功能，不需要写入功能
type MetricsRepositoryInterface interface {
	IsMetricsRepository()
	// FindNPUCardsByPIDs 根据 node_id 和 pid 列表查询每个 pid 占用的 NPU 卡号
	FindNPUCardsByPIDs(nodeID string, pids []int64) (map[int64][]int, error)
	// FindNPUCardsByPIDsWithStatuses 根据状态过滤查询每个 pid 占用的 NPU 卡号；statuses 为空表示不过滤状态
	FindNPUCardsByPIDsWithStatuses(nodeID string, pids []int64, statuses []string) (map[int64][]int, error)
	// DistinctNPUCardCounts 查询所有任务组的去重卡数列表
	DistinctNPUCardCounts() ([]int, error)
	// FindNPUProcessesByPID 查询单个进程占用的所有 NPU 记录
	FindNPUProcessesByPID(nodeID string, pid int64) ([]model.NPUProcess, error)
	// FindNPUProcessesByPIDs 批量查询多个进程占用的所有 NPU 记录
	FindNPUProcessesByPIDs(nodeID string, pids []int64) ([]model.NPUProcess, error)
	// FindNPUProcessesByPIDsWithStatuses 按状态过滤批量查询多个进程占用的所有 NPU 记录
	FindNPUProcessesByPIDsWithStatuses(nodeID string, pids []int64, statuses []string) ([]model.NPUProcess, error)
	// FindLatestNPUMetrics 查询指定卡号的最新 NPU 指标
	FindLatestNPUMetrics(nodeID string, npuIDs []int) ([]model.NPUMetric, error)
}

// JobAnalysisRepositoryInterface defines the interface for job analysis repository operations
type JobAnalysisRepositoryInterface interface {
	FindByJobID(jobID string) (*model.JobAnalysis, error)
	FindByJobIDs(jobIDs []string) ([]model.JobAnalysis, error)
	Upsert(analysis *model.JobAnalysis) error
}

// UserRepositoryInterface defines the interface for user repository operations
type UserRepositoryInterface interface {
	FindByID(id uint) (*model.User, error)
	FindByUsername(username string) (*model.User, error)
	FindAll() ([]model.User, error)
	Create(user *model.User) error
	Update(user *model.User) error
	Delete(id uint) error
	Count() (int64, error)
}
