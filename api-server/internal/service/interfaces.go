package service

import "github.com/task-monitor/api-server/internal/model"

// NodeServiceInterface defines the interface for node service operations
type NodeServiceInterface interface {
	GetNodes() ([]model.Node, error)
	GetNodeByID(nodeID string) (*model.Node, error)
	GetNodesByStatus(status string) ([]model.Node, error)
	GetNodeStats() (map[string]int64, error)
}

// JobGroup 作业分组（按 node_id + pgid 分组）
type JobGroup struct {
	MainJob   model.Job   `json:"mainJob"`
	ChildJobs []model.Job `json:"childJobs"`
	CardCount int         `json:"cardCount"`
}

// JobServiceInterface defines the interface for job service operations
type JobServiceInterface interface {
	GetJobByID(jobID string) (*model.Job, error)
	GetJobsByNodeID(nodeID string) ([]model.Job, error)
	GetJobsByStatus(status string) ([]model.Job, error)
	GetAllJobs() ([]model.Job, error)
	GetJobs(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string, page, pageSize int) ([]model.Job, int64, error)
	GetGroupedJobs(nodeID string, statuses []string, jobTypes []string, frameworks []string, cardCounts []int, sortBy, sortOrder string, page, pageSize int) ([]JobGroup, int64, error)
	GetDistinctCardCounts() ([]int, error)
	GetJobParameters(jobID string) ([]model.Parameter, error)
	GetJobCode(jobID string) ([]model.Code, error)
	GetJobStats() (map[string]int64, error)
}
