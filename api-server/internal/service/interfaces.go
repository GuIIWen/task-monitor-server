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
	GetJobParameters(jobID string) ([]model.Parameter, error)
	GetJobCode(jobID string) ([]model.Code, error)
}
