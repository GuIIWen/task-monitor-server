package service

import (
	"fmt"

	"github.com/task-monitor/api-server/internal/model"
	"github.com/task-monitor/api-server/internal/repository"
)

// JobService 作业服务
type JobService struct {
	jobRepo       repository.JobRepositoryInterface
	paramRepo     repository.ParameterRepositoryInterface
	codeRepo      repository.CodeRepositoryInterface
	metricsRepo   repository.MetricsRepositoryInterface
}

// NewJobService 创建作业服务
func NewJobService(
	jobRepo repository.JobRepositoryInterface,
	paramRepo repository.ParameterRepositoryInterface,
	codeRepo repository.CodeRepositoryInterface,
	metricsRepo repository.MetricsRepositoryInterface,
) *JobService {
	return &JobService{
		jobRepo:     jobRepo,
		paramRepo:   paramRepo,
		codeRepo:    codeRepo,
		metricsRepo: metricsRepo,
	}
}

// GetJobByID 根据ID获取作业
func (s *JobService) GetJobByID(jobID string) (*model.Job, error) {
	return s.jobRepo.FindByID(jobID)
}

// GetJobsByNodeID 根据节点ID获取作业列表
func (s *JobService) GetJobsByNodeID(nodeID string) ([]model.Job, error) {
	return s.jobRepo.FindByNodeID(nodeID)
}

// GetJobsByStatus 根据状态获取作业列表
func (s *JobService) GetJobsByStatus(status string) ([]model.Job, error) {
	return s.jobRepo.FindByStatus(status)
}

// GetJobParameters 获取作业参数
func (s *JobService) GetJobParameters(jobID string) ([]model.Parameter, error) {
	return s.paramRepo.FindByJobID(jobID)
}

// GetJobCode 获取作业代码
func (s *JobService) GetJobCode(jobID string) ([]model.Code, error) {
	return s.codeRepo.FindByJobID(jobID)
}

// GetAllJobs 获取所有作业
func (s *JobService) GetAllJobs() ([]model.Job, error) {
	return s.jobRepo.FindAll()
}

// GetJobs 灵活查询作业，支持多条件筛选和排序
func (s *JobService) GetJobs(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string, page, pageSize int) ([]model.Job, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	total, err := s.jobRepo.Count(nodeID, statuses, jobTypes, frameworks)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	jobs, err := s.jobRepo.Find(nodeID, statuses, jobTypes, frameworks, sortBy, sortOrder, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}

// GetJobStats 获取作业统计信息
func (s *JobService) GetJobStats() (map[string]int64, error) {
	jobs, err := s.jobRepo.FindAll()
	if err != nil {
		return nil, err
	}

	stats := map[string]int64{
		"total":     int64(len(jobs)),
		"running":   0,
		"completed": 0,
		"failed":    0,
		"stopped":   0,
		"lost":      0,
	}

	for _, job := range jobs {
		if job.Status != nil {
			switch *job.Status {
			case "running":
				stats["running"]++
			case "completed":
				stats["completed"]++
			case "failed":
				stats["failed"]++
			case "stopped":
				stats["stopped"]++
			case "lost":
				stats["lost"]++
			}
		}
	}

	return stats, nil
}

// GetGroupedJobs 按 node_id+pgid 分组查询作业
func (s *JobService) GetGroupedJobs(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string, page, pageSize int) ([]JobGroup, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	total, err := s.jobRepo.CountGroups(nodeID, statuses, jobTypes, frameworks)
	if err != nil {
		return nil, 0, fmt.Errorf("count groups: %w", err)
	}

	offset := (page - 1) * pageSize
	jobs, err := s.jobRepo.FindGrouped(nodeID, statuses, jobTypes, frameworks, sortBy, sortOrder, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("find grouped: %w", err)
	}

	// 按 node_id+pgid 组装分组
	groupMap := make(map[string]*JobGroup)
	var groupOrder []string

	for i := range jobs {
		job := jobs[i]
		var nid string
		var pgid int64
		if job.NodeID != nil {
			nid = *job.NodeID
		}
		if job.PGID != nil {
			pgid = *job.PGID
		}
		key := fmt.Sprintf("%s_%d", nid, pgid)

		if g, ok := groupMap[key]; ok {
			g.ChildJobs = append(g.ChildJobs, job)
			g.CardCount++
		} else {
			groupMap[key] = &JobGroup{
				MainJob:   job,
				ChildJobs: []model.Job{},
				CardCount: 1,
			}
			groupOrder = append(groupOrder, key)
		}
	}

	// 按查询顺序组装结果
	groups := make([]JobGroup, 0, len(groupOrder))
	for _, key := range groupOrder {
		groups = append(groups, *groupMap[key])
	}

	return groups, total, nil
}
