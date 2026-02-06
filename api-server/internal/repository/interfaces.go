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
	FindByStatus(status string) ([]model.Job, error)
	FindAll() ([]model.Job, error)
	Find(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string, limit, offset int) ([]model.Job, error)
	Count(nodeID string, statuses []string, jobTypes []string, frameworks []string) (int64, error)
	FindGrouped(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string, limit, offset int) ([]model.Job, error)
	CountGroups(nodeID string, statuses []string, jobTypes []string, frameworks []string) (int64, error)
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
	// IsMetricsRepository 是一个标记方法，确保类型安全
	// 如果需要查询metrics可以添加查询方法，例如：
	// FindNPUMetricsByJobID(jobID string) ([]model.NPUMetric, error)
	IsMetricsRepository()
}
