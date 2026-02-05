package repository

import "github.com/task-monitor/api-server/internal/model"

// NodeRepositoryInterface defines the interface for node repository operations
type NodeRepositoryInterface interface {
	Create(node *model.Node) error
	Update(node *model.Node) error
	FindByID(nodeID string) (*model.Node, error)
	FindAll() ([]model.Node, error)
	FindByStatus(status string) ([]model.Node, error)
	UpdateHeartbeat(nodeID string) error
	Upsert(node *model.Node) error
}

// JobRepositoryInterface defines the interface for job repository operations
type JobRepositoryInterface interface {
	Create(job *model.Job) error
	Update(job *model.Job) error
	FindByID(jobID string) (*model.Job, error)
	FindByNodeID(nodeID string) ([]model.Job, error)
	FindByStatus(status string) ([]model.Job, error)
	UpdateStatus(jobID, status, reason string) error
	Upsert(job *model.Job) error
}

// ParameterRepositoryInterface defines the interface for parameter repository operations
type ParameterRepositoryInterface interface {
	Create(param *model.Parameter) error
	FindByJobID(jobID string) ([]model.Parameter, error)
	BatchCreate(params []model.Parameter) error
}

// CodeRepositoryInterface defines the interface for code repository operations
type CodeRepositoryInterface interface {
	Create(code *model.Code) error
	FindByJobID(jobID string) ([]model.Code, error)
	BatchCreate(codes []model.Code) error
}

// MetricsRepositoryInterface defines the interface for metrics repository operations
type MetricsRepositoryInterface interface {
	CreateNPUMetric(metric *model.NPUMetric) error
	CreateProcessMetric(metric *model.ProcessMetric) error
	BatchCreateNPUMetrics(metrics []model.NPUMetric) error
}
