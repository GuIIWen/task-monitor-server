package service

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/task-monitor/api-server/internal/model"
	"github.com/task-monitor/api-server/internal/repository"
)

// chipSnapshot 与 agent 端 chipSnapshot 结构一致，用于解析 card_metrics_snapshot JSON
type chipSnapshot struct {
	BusID              string  `json:"busId"`
	Health             string  `json:"health"`
	PowerW             float64 `json:"powerW"`
	TempC              float64 `json:"tempC"`
	AICoreUsagePercent float64 `json:"aicorePercent"`
	HBMUsageMB         float64 `json:"hbmUsageMb"`
	HBMTotalMB         float64 `json:"hbmTotalMb"`
}

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

// GetJobDetail 获取作业详情（含 NPU 卡信息和关联进程）
// aggregate=true 时会聚合同组子进程的 NPU 数据（用于主作业详情页）；
// aggregate=false 时只查询该进程自身的 NPU 数据（用于子进程展开详情）。
func (s *JobService) GetJobDetail(jobID string, aggregate bool) (*JobDetailResponse, error) {
	job, err := s.jobRepo.FindByID(jobID)
	if err != nil {
		return nil, err
	}

	resp := &JobDetailResponse{
		Job:         *job,
		NPUCards:    []NPUCardInfo{},
		RelatedJobs: []model.Job{},
	}

	if job.NodeID == nil || job.PID == nil {
		return resp, nil
	}
	nodeID := *job.NodeID
	pid := *job.PID

	// 1. 查询该进程的 NPU 占用
	npuProcs, err := s.metricsRepo.FindNPUProcessesByPID(nodeID, pid)
	if err != nil {
		return resp, nil // NPU 查询失败不影响基本信息
	}

	// 1.1 主进程无 NPU 记录时，尝试聚合同组子进程的 NPU 数据（仅 aggregate 模式）
	if aggregate && len(npuProcs) == 0 && job.PGID != nil {
		relatedJobs, err := s.jobRepo.FindByNodeIDAndPGID(nodeID, *job.PGID)
		if err == nil && len(relatedJobs) > 0 {
			var childPIDs []int64
			for _, rj := range relatedJobs {
				if rj.PID != nil && *rj.PID != pid {
					childPIDs = append(childPIDs, *rj.PID)
				}
			}
			if len(childPIDs) > 0 {
				childProcs, err := s.metricsRepo.FindNPUProcessesByPIDs(nodeID, childPIDs)
				if err == nil {
					npuProcs = childProcs
				}
			}
		}
	}

	// 1.1b 仍无 NPU 记录时，按 PPID 查找子进程（训练启动器的 worker 可能有独立 PGID）
	if aggregate && len(npuProcs) == 0 {
		childJobs, err := s.jobRepo.FindByNodeIDAndPPID(nodeID, pid)
		if err == nil && len(childJobs) > 0 {
			var childPIDs []int64
			for _, cj := range childJobs {
				if cj.PID != nil {
					childPIDs = append(childPIDs, *cj.PID)
				}
			}
			if len(childPIDs) > 0 {
				childProcs, err := s.metricsRepo.FindNPUProcessesByPIDs(nodeID, childPIDs)
				if err == nil {
					npuProcs = childProcs
				}
			}
		}
	}

	// 1.2 终态作业兜底：running 状态查不到 NPU 记录时，按 running+stopped 重查
	if len(npuProcs) == 0 && isTerminalJobStatus(job.Status) {
		allPIDs := []int64{pid}
		// aggregate 模式下聚合同组兄弟进程 + PPID 子进程
		if aggregate {
			if job.PGID != nil {
				relatedJobs, err := s.jobRepo.FindByNodeIDAndPGID(nodeID, *job.PGID)
				if err == nil {
					for _, rj := range relatedJobs {
						if rj.PID != nil && *rj.PID != pid {
							allPIDs = append(allPIDs, *rj.PID)
						}
					}
				}
			}
			childJobs, err := s.jobRepo.FindByNodeIDAndPPID(nodeID, pid)
			if err == nil {
				for _, cj := range childJobs {
					if cj.PID != nil {
						allPIDs = append(allPIDs, *cj.PID)
					}
				}
			}
		}
		fallbackProcs, err := s.metricsRepo.FindNPUProcessesByPIDsWithStatuses(nodeID, allPIDs, []string{"running", "stopped"})
		if err == nil {
			npuProcs = fallbackProcs
		}
	}

	// 2. 按 npu_id 去重（取最大显存占用），同时收集 chip_id 集合用于精确匹配
	cardMemory := make(map[int]float64)
	procChipIDs := make(map[int]map[int]struct{}) // npu_id -> {chip_id: {}}
	for _, np := range npuProcs {
		if np.NPUID == nil {
			continue
		}
		npuID := *np.NPUID
		if np.ChipID != nil {
			if _, ok := procChipIDs[npuID]; !ok {
				procChipIDs[npuID] = make(map[int]struct{})
			}
			procChipIDs[npuID][*np.ChipID] = struct{}{}
		}
		if np.MemoryUsageMB == nil {
			if _, ok := cardMemory[npuID]; !ok {
				cardMemory[npuID] = 0
			}
			continue
		}
		if prev, ok := cardMemory[npuID]; !ok || *np.MemoryUsageMB > prev {
			cardMemory[npuID] = *np.MemoryUsageMB
		}
	}

	if len(cardMemory) > 0 {
		npuIDs := make([]int, 0, len(cardMemory))
		for id := range cardMemory {
			npuIDs = append(npuIDs, id)
		}
		sort.Ints(npuIDs)

		// 3. 查询卡详情并按 npu_id 建立映射
		metricsByNPU := make(map[int][]model.NPUMetric)
		var metrics []model.NPUMetric
		// 已停止的作业：查询运行期间 HBM 峰值快照，反映真实使用量
		if isTerminalJobStatus(job.Status) && job.StartTime != nil && job.EndTime != nil && *job.EndTime > 0 {
			metrics, err = s.metricsRepo.FindNPUMetricsPeakInPeriod(nodeID, npuIDs, *job.StartTime, *job.EndTime)
		} else {
			metrics, err = s.metricsRepo.FindLatestNPUMetrics(nodeID, npuIDs)
		}
		if err == nil {
			for _, m := range metrics {
				if m.NPUID != nil {
					metricsByNPU[*m.NPUID] = append(metricsByNPU[*m.NPUID], m)
				}
			}
		}

		// 3.1 npu_metrics 为空时，从 npu_processes.card_metrics_snapshot 回退
		if len(metricsByNPU) == 0 {
			snapshotMetrics := parseSnapshotMetrics(npuProcs, nodeID)
			for _, m := range snapshotMetrics {
				if m.NPUID != nil {
					metricsByNPU[*m.NPUID] = append(metricsByNPU[*m.NPUID], m)
				}
			}
		}

		metricsByNPU = filterChipsByID(metricsByNPU, procChipIDs, aggregate)

		// 4. 组装 NPUCardInfo（按 npu_id 有序输出）
		for _, npuID := range npuIDs {
			resp.NPUCards = append(resp.NPUCards, NPUCardInfo{
				NpuID:         npuID,
				MemoryUsageMB: cardMemory[npuID],
				Metrics:       metricsByNPU[npuID],
			})
		}
	}

	// 5. 查关联 NPU 进程（同 pgid 范围内 Union-Find），仅 aggregate 模式
	if aggregate && job.PGID != nil {
		resp.RelatedJobs = s.findRelatedNPUJobs(job)
	}

	return resp, nil
}

// GetJobsByNodeID 根据节点ID获取作业列表
func (s *JobService) GetJobsByNodeID(nodeID string) ([]model.Job, error) {
	return s.jobRepo.FindByNodeID(nodeID)
}

// UpdateJobFields 更新作业的指定字段
func (s *JobService) UpdateJobFields(jobID string, fields map[string]interface{}) error {
	return s.jobRepo.UpdateFields(jobID, fields)
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

// findRelatedNPUJobs 查找同组的 NPU 关联进程（排除自身）
func (s *JobService) findRelatedNPUJobs(job *model.Job) []model.Job {
	if job.NodeID == nil || job.PID == nil {
		return nil
	}
	nodeID := *job.NodeID
	pid := *job.PID

	// 查同 pgid 的所有 jobs
	var allJobs []model.Job
	if job.PGID != nil {
		pgidJobs, err := s.jobRepo.FindByNodeIDAndPGID(nodeID, *job.PGID)
		if err == nil {
			allJobs = append(allJobs, pgidJobs...)
		}
	}

	// 按 PPID 查找子进程（训练启动器的 worker 可能有独立 PGID）
	ppidChildren, err := s.jobRepo.FindByNodeIDAndPPID(nodeID, pid)
	if err == nil {
		seen := make(map[string]bool, len(allJobs))
		for _, j := range allJobs {
			seen[j.JobID] = true
		}
		for _, cj := range ppidChildren {
			if !seen[cj.JobID] {
				allJobs = append(allJobs, cj)
			}
		}
	}

	if len(allJobs) <= 1 {
		return nil
	}

	parent := make([]int, len(allJobs))
	for i := range allJobs {
		parent[i] = i
	}

	var find func(int) int
	find = func(i int) int {
		if parent[i] != i {
			parent[i] = find(parent[i])
		}
		return parent[i]
	}

	union := func(a, b int) {
		rootA := find(a)
		rootB := find(b)
		if rootA != rootB {
			parent[rootA] = rootB
		}
	}

	pidIndexes := make(map[int64][]int)
	for i, j := range allJobs {
		if j.PID == nil {
			continue
		}
		pidIndexes[*j.PID] = append(pidIndexes[*j.PID], i)
	}

	for i, j := range allJobs {
		if j.PID == nil || j.PPID == nil {
			continue
		}
		candidates := pidIndexes[*j.PPID]
		parentIdx := chooseParentProcessIndex(allJobs, candidates, i)
		if parentIdx == -1 {
			continue
		}
		union(i, parentIdx)
	}

	// 兜底：同 pgid 的进程合并（ppid 链断裂时靠 pgid 兜底，与 buildGroupedJobs 一致）
	firstIdx := -1
	for i := range allJobs {
		if allJobs[i].PID == nil {
			continue
		}
		if firstIdx == -1 {
			firstIdx = i
		} else {
			union(i, firstIdx)
		}
	}

	selfIdx := -1
	for i := range allJobs {
		if allJobs[i].JobID == job.JobID {
			selfIdx = i
			break
		}
	}
	if selfIdx == -1 {
		return nil
	}

	selfRoot := find(selfIdx)
	groupPIDs := make([]int64, 0)
	for i, j := range allJobs {
		if i == selfIdx || j.PID == nil {
			continue
		}
		if find(i) == selfRoot {
			groupPIDs = append(groupPIDs, *j.PID)
		}
	}

	if len(groupPIDs) == 0 {
		return nil
	}

	// 查 NPU 占用，只返回有 NPU 记录的进程
	npuMap, err := s.metricsRepo.FindNPUCardsByPIDs(nodeID, dedupeInt64(groupPIDs))
	if err != nil {
		return nil
	}
	// 终态作业兜底：running 查不到时按 running+stopped 重查
	if len(npuMap) == 0 && isTerminalJobStatus(job.Status) {
		npuMap, err = s.metricsRepo.FindNPUCardsByPIDsWithStatuses(nodeID, dedupeInt64(groupPIDs), []string{"running", "stopped"})
		if err != nil || len(npuMap) == 0 {
			return nil
		}
	}
	if len(npuMap) == 0 {
		return nil
	}

	var related []model.Job
	for i, j := range allJobs {
		if i == selfIdx || j.PID == nil {
			continue
		}
		if find(i) == selfRoot && len(npuMap[*j.PID]) > 0 {
			related = append(related, j)
		}
	}
	return related
}

// GetJobStats 获取作业统计信息（按分组统计，与作业管理页一致）
func (s *JobService) GetJobStats() (map[string]int64, error) {
	jobs, err := s.jobRepo.FindAll()
	if err != nil {
		return nil, err
	}

	groups, err := s.buildGroupedJobs(jobs)
	if err != nil {
		return nil, err
	}
	groups = filterStopNameGroups(groups)

	stats := map[string]int64{
		"total":     int64(len(groups)),
		"running":   0,
		"completed": 0,
		"failed":    0,
		"stopped":   0,
		"lost":      0,
	}

	for _, group := range groups {
		if group.MainJob.Status != nil {
			switch *group.MainJob.Status {
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
	jobs, err := s.jobRepo.FindFiltered("", nil, nil, nil, "", "")
	if err != nil {
		return nil, fmt.Errorf("find filtered for card counts: %w", err)
	}

	groups, err := s.buildGroupedJobs(jobs)
	if err != nil {
		return nil, err
	}

	groups = filterStopNameGroups(groups)

	seen := make(map[int]struct{})
	counts := make([]int, 0)
	for _, group := range groups {
		if group.CardCount == nil {
			continue
		}
		if _, ok := seen[*group.CardCount]; ok {
			continue
		}
		seen[*group.CardCount] = struct{}{}
		counts = append(counts, *group.CardCount)
	}

	sort.Ints(counts)
	return counts, nil
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

	// 3. 过滤掉纯停止词进程的独立组（无业务子进程的 shell/容器运行时进程）
	groups = filterStopNameGroups(groups)

	// 4. 应用 cardCount 筛选
	groups = filterJobGroupsByCardCounts(groups, cardCounts)

	// 5. 内存分页
	total := int64(len(groups))
	return paginateJobGroups(groups, offset, pageSize), total, nil
}

// ppidStopNames 非业务进程停止词，Union-Find 合并时不穿越这些进程
var ppidStopNames = map[string]bool{
	"bash": true, "sh": true, "zsh": true, "fish": true, "csh": true, "tcsh": true, "dash": true,
	"sshd": true, "login": true, "su": true, "sudo": true, "screen": true, "tmux": true,
	"containerd-shim": true, "containerd-shim-runc-v2": true, "containerd": true,
	"dockerd": true, "docker": true, "runc": true,
	"systemd": true, "init": true, "supervisord": true,
}

type nodePIDKey struct {
	nodeID string
	pid    int64
}

func normalizeNodeID(nodeID *string) string {
	if nodeID == nil {
		return ""
	}
	return *nodeID
}

// chooseParentProcessIndex 从候选父进程中选择最合理的一个。
// 优先选择：非停止词进程、且 start_time 不晚于子进程且最接近的进程。
func chooseParentProcessIndex(jobs []model.Job, candidates []int, childIdx int) int {
	if len(candidates) == 0 {
		return -1
	}

	child := jobs[childIdx]
	bestIdx := -1
	bestRank := int64(1<<62 - 1)
	bestDiff := int64(1<<62 - 1)

	for _, idx := range candidates {
		if idx == childIdx {
			continue
		}
		parentJob := jobs[idx]
		if parentJob.ProcessName != nil && ppidStopNames[*parentJob.ProcessName] {
			continue
		}

		rank := int64(2)
		diff := int64(0)
		if child.StartTime != nil && parentJob.StartTime != nil {
			if *parentJob.StartTime <= *child.StartTime {
				rank = 0
				diff = *child.StartTime - *parentJob.StartTime
			} else {
				rank = 1
				diff = *parentJob.StartTime - *child.StartTime
			}
		}

		if bestIdx == -1 || rank < bestRank || (rank == bestRank && diff < bestDiff) ||
			(rank == bestRank && diff == bestDiff && idx < bestIdx) {
			bestIdx = idx
			bestRank = rank
			bestDiff = diff
		}
	}

	return bestIdx
}

func isTerminalJobStatus(status *string) bool {
	if status == nil {
		return false
	}
	switch *status {
	case "stopped", "completed", "failed", "lost":
		return true
	default:
		return false
	}
}

// parseSnapshotMetrics 从 npu_processes 的 card_metrics_snapshot 字段解析出 NPUMetric 列表。
// 当 npu_metrics 表数据已被清理时，用此快照作为回退数据源。
func parseSnapshotMetrics(npuProcs []model.NPUProcess, nodeID string) []model.NPUMetric {
	var result []model.NPUMetric
	seen := make(map[int]bool) // 按 npu_id 去重，每张卡只取一次快照

	for _, np := range npuProcs {
		if np.NPUID == nil || np.CardMetricsSnapshot == nil || *np.CardMetricsSnapshot == "" {
			continue
		}
		npuID := *np.NPUID
		if seen[npuID] {
			continue
		}
		seen[npuID] = true

		var chips []chipSnapshot
		if err := json.Unmarshal([]byte(*np.CardMetricsSnapshot), &chips); err != nil {
			continue
		}

		for _, chip := range chips {
			busID := chip.BusID
			health := chip.Health
			powerW := chip.PowerW
			tempC := chip.TempC
			aicore := chip.AICoreUsagePercent
			hbmUsage := chip.HBMUsageMB
			hbmTotal := chip.HBMTotalMB

			result = append(result, model.NPUMetric{
				NodeID:             &nodeID,
				NPUID:              &npuID,
				BusID:              &busID,
				Health:             &health,
				PowerW:             &powerW,
				TempC:              &tempC,
				AICoreUsagePercent: &aicore,
				HBMUsageMB:         &hbmUsage,
				HBMTotalMB:         &hbmTotal,
			})
		}
	}
	return result
}

// filterChipsByID 根据 npu_processes 中记录的 chip_id 集合过滤 chip。
// 规则：
// 1. 该卡有 chip_id 集合：按 bus_id 排序后保留集合内全部索引；若过滤结果为空则回退原始 chips。
// 2. 该卡无 chip_id 集合：aggregate=true 保持原样（优先保召回）；aggregate=false 回退启发式过滤。
func filterChipsByID(metricsByNPU map[int][]model.NPUMetric, procChipIDs map[int]map[int]struct{}, aggregate bool) map[int][]model.NPUMetric {
	for npuID, chips := range metricsByNPU {
		if len(chips) <= 1 {
			continue
		}

		chipSet := procChipIDs[npuID]
		if len(chipSet) > 0 {
			// 按 bus_id 排序，chip_id 对应排序后的索引
			sortedChips := append([]model.NPUMetric(nil), chips...)
			sort.Slice(sortedChips, func(i, j int) bool {
				bi, bj := "", ""
				if sortedChips[i].BusID != nil {
					bi = *sortedChips[i].BusID
				}
				if sortedChips[j].BusID != nil {
					bj = *sortedChips[j].BusID
				}
				return bi < bj
			})

			filtered := make([]model.NPUMetric, 0, len(sortedChips))
			for idx, chip := range sortedChips {
				if _, ok := chipSet[idx]; ok {
					filtered = append(filtered, chip)
				}
			}

			// chip_id 与 bus_id 映射异常时保留原始 chips，避免误删
			if len(filtered) > 0 {
				metricsByNPU[npuID] = filtered
			}
			continue
		}

		if aggregate {
			continue
		}

		// 子进程视图在缺少 chip_id 时对单卡做启发式过滤
		filteredCard := filterActiveChips(map[int][]model.NPUMetric{npuID: chips})
		if fc, ok := filteredCard[npuID]; ok {
			metricsByNPU[npuID] = fc
		}
	}
	return metricsByNPU
}

// filterActiveChips 过滤空闲 chip：对每张卡（npu_id），如果最大 HBM 的 chip
// 比最小 HBM 的 chip 高出 2 倍以上，则只保留 HBM 超过最大值 30% 的 chip。
func filterActiveChips(metricsByNPU map[int][]model.NPUMetric) map[int][]model.NPUMetric {
	for npuID, chips := range metricsByNPU {
		if len(chips) <= 1 {
			continue
		}
		var maxHBM, minHBM float64
		first := true
		for _, c := range chips {
			hbm := float64(0)
			if c.HBMUsageMB != nil {
				hbm = *c.HBMUsageMB
			}
			if first {
				maxHBM, minHBM = hbm, hbm
				first = false
			} else {
				if hbm > maxHBM {
					maxHBM = hbm
				}
				if hbm < minHBM {
					minHBM = hbm
				}
			}
		}
		if minHBM <= 0 || maxHBM < minHBM*2 {
			continue
		}
		threshold := maxHBM * 0.3
		filtered := make([]model.NPUMetric, 0, len(chips))
		for _, c := range chips {
			hbm := float64(0)
			if c.HBMUsageMB != nil {
				hbm = *c.HBMUsageMB
			}
			if hbm >= threshold {
				filtered = append(filtered, c)
			}
		}
		if len(filtered) > 0 {
			metricsByNPU[npuID] = filtered
		}
	}
	return metricsByNPU
}

func hasNPUForAnyPID(npuMap map[int64][]int, pids []int64) bool {
	if npuMap == nil || len(pids) == 0 {
		return false
	}
	for _, pid := range pids {
		if len(npuMap[pid]) > 0 {
			return true
		}
	}
	return false
}

// buildGroupedJobs 使用 Union-Find 按 ppid 链路构建进程树分组，并补充卡数信息
func (s *JobService) buildGroupedJobs(jobs []model.Job) ([]JobGroup, error) {
	if len(jobs) == 0 {
		return []JobGroup{}, nil
	}

	parent := make([]int, len(jobs))
	for i := range jobs {
		parent[i] = i
	}

	var find func(int) int
	find = func(i int) int {
		if parent[i] != i {
			parent[i] = find(parent[i])
		}
		return parent[i]
	}

	union := func(a, b int) {
		rootA := find(a)
		rootB := find(b)
		if rootA != rootB {
			parent[rootA] = rootB
		}
	}

	pidIndexes := make(map[nodePIDKey][]int)
	for i, job := range jobs {
		if job.PID == nil {
			continue
		}
		key := nodePIDKey{nodeID: normalizeNodeID(job.NodeID), pid: *job.PID}
		pidIndexes[key] = append(pidIndexes[key], i)
	}

	parentLink := make(map[int]int)
	// 合并：根据 node_id + ppid 查找候选父进程，再按时间接近度选择最优父进程
	for i, job := range jobs {
		if job.PID == nil || job.PPID == nil {
			continue
		}
		key := nodePIDKey{nodeID: normalizeNodeID(job.NodeID), pid: *job.PPID}
		candidates := pidIndexes[key]
		parentIdx := chooseParentProcessIndex(jobs, candidates, i)
		if parentIdx == -1 {
			continue
		}
		union(i, parentIdx)
		parentLink[i] = parentIdx
	}

	// 兜底：同 node_id + pgid 的进程合并（ppid 链断裂时靠 pgid 兜底）
	type nodePGIDKey struct {
		nodeID string
		pgid   int64
	}
	pgidFirst := make(map[nodePGIDKey]int)
	for i, job := range jobs {
		if job.PID == nil || job.PGID == nil || *job.PGID == 0 {
			continue
		}
		key := nodePGIDKey{nodeID: normalizeNodeID(job.NodeID), pgid: *job.PGID}
		if first, ok := pgidFirst[key]; ok {
			union(i, first)
		} else {
			pgidFirst[key] = i
		}
	}

	// 按根 pid 聚合分组
	type groupInfo struct {
		jobs []int // jobs 数组下标
		nid  string
		pids []int64
	}
	groupMap := make(map[int]*groupInfo)
	var rootOrder []int

	for i, job := range jobs {
		if job.PID == nil {
			continue
		}
		root := find(i)
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

	// 终态作业在 npu_processes 仅保留 stopped 记录时，running 过滤会拿不到卡信息。
	// 仅在终态组且 running 数据为空时，按 running+stopped 做一次兜底查询。
	nodeNPUFallbackMap := make(map[string]map[int64][]int)

	// 组装 JobGroup 结果
	groups := make([]JobGroup, 0, len(rootOrder))
	for _, root := range rootOrder {
		info := groupMap[root]

		// 选择 MainJob：没有父链接的进程作为根；如有多个，选 start_time 最早的
		mainIdx := info.jobs[0]
		for _, idx := range info.jobs {
			_, jobHasParent := parentLink[idx]
			_, curHasParent := parentLink[mainIdx]
			if !jobHasParent && curHasParent {
				mainIdx = idx
			} else if !jobHasParent && !curHasParent {
				job := jobs[idx]
				curMain := jobs[mainIdx]
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
		npuMap := nodeNPUMap[info.nid]
		groupHasNPU := hasNPUForAnyPID(npuMap, info.pids)
		if !groupHasNPU && isTerminalJobStatus(group.MainJob.Status) && info.nid != "" {
			fallbackMap, ok := nodeNPUFallbackMap[info.nid]
			if !ok {
				var err error
				fallbackMap, err = s.metricsRepo.FindNPUCardsByPIDsWithStatuses(info.nid, dedupeInt64(nodeAllPIDs[info.nid]), []string{"running", "stopped"})
				if err != nil {
					return nil, fmt.Errorf("find fallback npu cards: %w", err)
				}
				nodeNPUFallbackMap[info.nid] = fallbackMap
			}
			if hasNPUForAnyPID(fallbackMap, info.pids) {
				npuMap = fallbackMap
				groupHasNPU = true
			}
		}

		includeAllChildren := isTerminalJobStatus(group.MainJob.Status) && !groupHasNPU
		for _, idx := range info.jobs {
			if idx != mainIdx {
				child := jobs[idx]
				// 运行态默认只展示 NPU 子进程；终态且无 NPU 数据时兜底展示链路子进程。
				if includeAllChildren || (child.PID != nil && npuMap != nil && len(npuMap[*child.PID]) > 0) {
					group.ChildJobs = append(group.ChildJobs, child)
				}
			}
		}

		// 计算卡数（复用上面已获取的 npuMap）
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

// filterStopNameGroups 过滤掉纯停止词进程的独立组（组内所有进程都是非业务进程）
func filterStopNameGroups(groups []JobGroup) []JobGroup {
	filtered := make([]JobGroup, 0, len(groups))
	for _, group := range groups {
		hasBusinessJob := false
		// 检查 MainJob
		if group.MainJob.ProcessName == nil || !ppidStopNames[*group.MainJob.ProcessName] {
			hasBusinessJob = true
		}
		// 检查 ChildJobs
		if !hasBusinessJob {
			for _, child := range group.ChildJobs {
				if child.ProcessName == nil || !ppidStopNames[*child.ProcessName] {
					hasBusinessJob = true
					break
				}
			}
		}
		if hasBusinessJob {
			filtered = append(filtered, group)
		}
	}
	return filtered
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
