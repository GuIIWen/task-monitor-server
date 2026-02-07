package service

import (
	"fmt"

	"github.com/task-monitor/api-server/internal/model"
	"github.com/task-monitor/api-server/internal/repository"
)

// JobService 作业服务
type JobService struct {
	jobRepo     repository.JobRepositoryInterface
	paramRepo   repository.ParameterRepositoryInterface
	codeRepo    repository.CodeRepositoryInterface
	metricsRepo repository.MetricsRepositoryInterface
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

// GetDistinctCardCounts 获取所有去重的卡数值（基于 npu_processes）
func (s *JobService) GetDistinctCardCounts() ([]int, error) {
	return s.metricsRepo.DistinctNPUCardCounts()
}

// GetGroupedJobs 按 node_id+pgid 分组查询作业
func (s *JobService) GetGroupedJobs(nodeID string, statuses []string, jobTypes []string, frameworks []string, cardCounts []int, sortBy, sortOrder string, page, pageSize int) ([]JobGroup, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 当存在 cardCount 筛选时，需要先拿到完整分组后再过滤，否则会出现分页总数错误/空页问题。
	if len(cardCounts) > 0 {
		jobs, err := s.jobRepo.FindGrouped(nodeID, statuses, jobTypes, frameworks, sortBy, sortOrder, 0, 0)
		if err != nil {
			return nil, 0, fmt.Errorf("find grouped for cardCount filter: %w", err)
		}

		groups, err := s.buildGroupedJobs(jobs)
		if err != nil {
			return nil, 0, err
		}

		groups = filterJobGroupsByCardCounts(groups, cardCounts)
		total := int64(len(groups))
		return paginateJobGroups(groups, offset, pageSize), total, nil
	}

	total, err := s.jobRepo.CountGroups(nodeID, statuses, jobTypes, frameworks)
	if err != nil {
		return nil, 0, fmt.Errorf("count groups: %w", err)
	}

	jobs, err := s.jobRepo.FindGrouped(nodeID, statuses, jobTypes, frameworks, sortBy, sortOrder, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("find grouped: %w", err)
	}

	groups, err := s.buildGroupedJobs(jobs)
	if err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// buildGroupedJobs 将仓储层返回的 jobs 组装为分组结果，并补充卡数信息
func (s *JobService) buildGroupedJobs(jobs []model.Job) ([]JobGroup, error) {
	if len(jobs) == 0 {
		return []JobGroup{}, nil
	}

	// 按 node_id+pgid+start_time 组装分组
	type groupInfo struct {
		group *JobGroup
		nid   string
		pids  []int64
	}
	groupMap := make(map[string]*groupInfo)
	var groupOrder []string

	for i := range jobs {
		job := jobs[i]
		var nid string
		if job.NodeID != nil {
			nid = *job.NodeID
		}
		key := buildGroupIdentity(job)

		if g, ok := groupMap[key]; ok {
			g.group.ChildJobs = append(g.group.ChildJobs, job)
			if job.PID != nil {
				g.pids = append(g.pids, *job.PID)
			}
		} else {
			var pids []int64
			if job.PID != nil {
				pids = append(pids, *job.PID)
			}
			groupMap[key] = &groupInfo{
				group: &JobGroup{
					MainJob:   job,
					ChildJobs: []model.Job{},
					CardCount: nil,
				},
				nid:  nid,
				pids: pids,
			}
			groupOrder = append(groupOrder, key)
		}
	}

	// 按 node_id 聚合所有 pid，批量查询 NPU 卡信息
	nodeAllPIDs := make(map[string][]int64)
	for _, key := range groupOrder {
		info := groupMap[key]
		nodeAllPIDs[info.nid] = append(nodeAllPIDs[info.nid], info.pids...)
	}

	// 批量查询每个 node 的 NPU 信息
	nodeNPUMap := make(map[string]map[int64][]int) // node_id -> pid -> []npu_id
	for nid, pids := range nodeAllPIDs {
		if nid == "" || len(pids) == 0 {
			continue
		}
		npuMap, err := s.metricsRepo.FindNPUCardsByPIDs(nid, dedupeInt64(pids))
		if err != nil {
			return nil, fmt.Errorf("find npu cards: %w", err)
		}
		nodeNPUMap[nid] = npuMap
	}

	// 计算每组的卡数
	for _, key := range groupOrder {
		info := groupMap[key]
		npuMap := nodeNPUMap[info.nid]
		if npuMap == nil {
			continue
		}
		// 收集该组所有 pid 对应的 npu_id，去重
		npuSet := make(map[int]bool)
		for _, pid := range info.pids {
			for _, npuID := range npuMap[pid] {
				npuSet[npuID] = true
			}
		}
		if len(npuSet) > 0 {
			count := len(npuSet)
			info.group.CardCount = &count
		}
	}

	// 按查询顺序组装结果
	groups := make([]JobGroup, 0, len(groupOrder))
	for _, key := range groupOrder {
		groups = append(groups, *groupMap[key].group)
	}

	return groups, nil
}

// matchCardCount 检查卡数是否在筛选列表中
func matchCardCount(cardCount *int, cardCounts []int) bool {
	for _, c := range cardCounts {
		if c == 0 && cardCount == nil {
			return true
		}
		if cardCount != nil && *cardCount == c {
			return true
		}
	}
	return false
}

// filterJobGroupsByCardCounts 根据 cardCount 条件过滤分组
func filterJobGroupsByCardCounts(groups []JobGroup, cardCounts []int) []JobGroup {
	if len(cardCounts) == 0 {
		return groups
	}

	filtered := make([]JobGroup, 0, len(groups))
	for _, group := range groups {
		if matchCardCount(group.CardCount, cardCounts) {
			filtered = append(filtered, group)
		}
	}
	return filtered
}

// paginateJobGroups 对分组结果进行内存分页
func paginateJobGroups(groups []JobGroup, offset, limit int) []JobGroup {
	if offset >= len(groups) {
		return []JobGroup{}
	}
	end := offset + limit
	if end > len(groups) {
		end = len(groups)
	}
	return groups[offset:end]
}

// buildGroupIdentity 生成可区分 NULL 与零值的分组键
func buildGroupIdentity(job model.Job) string {
	return fmt.Sprintf(
		"node:%s|pgid:%s|start:%s",
		nullableString(job.NodeID),
		nullableInt64(job.PGID),
		nullableInt64(job.StartTime),
	)
}

func nullableString(v *string) string {
	if v == nil {
		return "<nil>"
	}
	return *v
}

func nullableInt64(v *int64) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%d", *v)
}

// dedupeInt64 去重并保留原始顺序，避免重复 pid 放大 IN 条件
func dedupeInt64(items []int64) []int64 {
	seen := make(map[int64]struct{}, len(items))
	result := make([]int64, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}
