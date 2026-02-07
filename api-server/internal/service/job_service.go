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

// GetGroupedJobs 按 ppid 链路构建进程树分组查询作业
func (s *JobService) GetGroupedJobs(nodeID string, statuses []string, jobTypes []string, frameworks []string, cardCounts []int, sortBy, sortOrder string, page, pageSize int) ([]JobGroup, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 1. 查出所有符合条件的 jobs
	jobs, err := s.jobRepo.FindFiltered(nodeID, statuses, jobTypes, frameworks, sortBy, sortOrder)
	if err != nil {
		return nil, 0, fmt.Errorf("find filtered: %w", err)
	}

	// 2. 在内存中按 ppid 链路构建进程树分组
	groups, err := s.buildGroupedJobs(jobs)
	if err != nil {
		return nil, 0, err
	}

	// 3. 应用 cardCount 筛选
	groups = filterJobGroupsByCardCounts(groups, cardCounts)

	// 4. 内存分页
	total := int64(len(groups))
	return paginateJobGroups(groups, offset, pageSize), total, nil
}

// buildGroupedJobs 使用 Union-Find 按 ppid 链路构建进程树分组，并补充卡数信息
func (s *JobService) buildGroupedJobs(jobs []model.Job) ([]JobGroup, error) {
	if len(jobs) == 0 {
		return []JobGroup{}, nil
	}

	// Union-Find: 按 node_id 分别构建进程树
	parent := make(map[int64]int64)
	pidIndex := make(map[int64]int) // pid -> jobs 数组下标

	for i, job := range jobs {
		if job.PID == nil {
			continue
		}
		pid := *job.PID
		parent[pid] = pid
		pidIndex[pid] = i
	}

	var find func(int64) int64
	find = func(pid int64) int64 {
		if parent[pid] != pid {
			parent[pid] = find(parent[pid])
		}
		return parent[pid]
	}

	// 合并：如果 ppid 在同一 node 的集合中，合并到 ppid 的根
	for _, job := range jobs {
		if job.PID == nil || job.PPID == nil {
			continue
		}
		ppid := *job.PPID
		if _, exists := parent[ppid]; exists {
			// 确保同一 node_id 才合并
			ppidIdx := pidIndex[ppid]
			parentJob := jobs[ppidIdx]
			sameNode := (job.NodeID == nil && parentJob.NodeID == nil) ||
				(job.NodeID != nil && parentJob.NodeID != nil && *job.NodeID == *parentJob.NodeID)
			if sameNode {
				rootA := find(*job.PID)
				rootB := find(ppid)
				if rootA != rootB {
					parent[rootA] = rootB // 子合并到父
				}
			}
		}
	}

	// 按根 pid 聚合分组
	type groupInfo struct {
		jobs []int // jobs 数组下标
		nid  string
		pids []int64
	}
	groupMap := make(map[int64]*groupInfo)
	var rootOrder []int64

	for i, job := range jobs {
		if job.PID == nil {
			continue
		}
		root := find(*job.PID)
		if g, ok := groupMap[root]; ok {
			g.jobs = append(g.jobs, i)
			g.pids = append(g.pids, *job.PID)
		} else {
			var nid string
			if job.NodeID != nil {
				nid = *job.NodeID
			}
			groupMap[root] = &groupInfo{
				jobs: []int{i},
				nid:  nid,
				pids: []int64{*job.PID},
			}
			rootOrder = append(rootOrder, root)
		}
	}

	// 批量查询 NPU 卡信息
	nodeAllPIDs := make(map[string][]int64)
	for _, root := range rootOrder {
		info := groupMap[root]
		nodeAllPIDs[info.nid] = append(nodeAllPIDs[info.nid], info.pids...)
	}

	nodeNPUMap := make(map[string]map[int64][]int)
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

	// 组装 JobGroup 结果
	groups := make([]JobGroup, 0, len(rootOrder))
	for _, root := range rootOrder {
		info := groupMap[root]

		// 选择 MainJob：PPID 不在组内的进程作为根；如有多个，选 start_time 最早的
		pidSet := make(map[int64]bool, len(info.pids))
		for _, pid := range info.pids {
			pidSet[pid] = true
		}

		mainIdx := info.jobs[0]
		for _, idx := range info.jobs {
			job := jobs[idx]
			curMain := jobs[mainIdx]
			jobIsRoot := job.PPID == nil || !pidSet[*job.PPID]
			curIsRoot := curMain.PPID == nil || !pidSet[*curMain.PPID]
			if jobIsRoot && !curIsRoot {
				mainIdx = idx
			} else if jobIsRoot && curIsRoot {
				// 都是根，选 start_time 最早的
				if job.StartTime != nil && curMain.StartTime != nil && *job.StartTime < *curMain.StartTime {
					mainIdx = idx
				}
			}
		}

		group := JobGroup{
			MainJob:   jobs[mainIdx],
			ChildJobs: make([]model.Job, 0, len(info.jobs)-1),
		}
		for _, idx := range info.jobs {
			if idx != mainIdx {
				group.ChildJobs = append(group.ChildJobs, jobs[idx])
			}
		}

		// 计算卡数
		npuMap := nodeNPUMap[info.nid]
		if npuMap != nil {
			npuSet := make(map[int]bool)
			for _, pid := range info.pids {
				for _, npuID := range npuMap[pid] {
					npuSet[npuID] = true
				}
			}
			if len(npuSet) > 0 {
				count := len(npuSet)
				group.CardCount = &count
			}
		}

		groups = append(groups, group)
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
