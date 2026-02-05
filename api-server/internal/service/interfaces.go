package service

import "github.com/task-monitor/api-server/internal/model"

// NodeServiceInterface defines the interface for node service operations
type NodeServiceInterface interface {
	GetNodes() ([]model.Node, error)
	GetNodeByID(nodeID string) (*model.Node, error)
	GetNodesByStatus(status string) ([]model.Node, error)
}

// JobServiceInterface defines the interface for job service operations
type JobServiceInterface interface {
	GetJobByID(jobID string) (*model.Job, error)
	GetJobsByNodeID(nodeID string) ([]model.Job, error)
	GetJobsByStatus(status string) ([]model.Job, error)
	GetAllJobs() ([]model.Job, error)
	GetJobs(nodeID, status string, page, pageSize int) ([]model.Job, int64, error) // 灵活查询，支持多条件筛选和分页
	GetJobParameters(jobID string) ([]model.Parameter, error)
	GetJobCode(jobID string) ([]model.Code, error)
}
