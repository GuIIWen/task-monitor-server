package repository

import (
	"time"

	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
)

// MetricsRepository 指标数据访问层
// API Server只负责查询，不负责写入
// 如果需要查询metrics数据，可以在这里添加查询方法
type MetricsRepository struct {
	db *gorm.DB
}

// NewMetricsRepository 创建指标Repository
func NewMetricsRepository(db *gorm.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

// IsMetricsRepository 实现MetricsRepositoryInterface的标记方法
func (r *MetricsRepository) IsMetricsRepository() {}

// npuPidRow 用于扫描 FindNPUCardsByPIDs 查询结果
type npuPidRow struct {
	PID   int64 `gorm:"column:pid"`
	NPUID int   `gorm:"column:npu_id"`
}

// FindNPUCardsByPIDs 根据 node_id 和 pid 列表查询每个 pid 占用的去重 NPU 卡号
func (r *MetricsRepository) FindNPUCardsByPIDs(nodeID string, pids []int64) (map[int64][]int, error) {
	return r.FindNPUCardsByPIDsWithStatuses(nodeID, pids, []string{"running"})
}

// FindNPUCardsByPIDsWithStatuses 根据 node_id、pid 列表和状态过滤查询每个 pid 占用的去重 NPU 卡号
func (r *MetricsRepository) FindNPUCardsByPIDsWithStatuses(nodeID string, pids []int64, statuses []string) (map[int64][]int, error) {
	if len(pids) == 0 {
		return make(map[int64][]int), nil
	}

	var rows []npuPidRow
	query := r.db.Table("npu_processes").
		Select("DISTINCT pid, npu_id").
		Where("node_id = ? AND pid IN ?", nodeID, pids)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	err := query.Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64][]int)
	for _, row := range rows {
		result[row.PID] = append(result[row.PID], row.NPUID)
	}
	return result, nil
}

// npuCardCountRow 用于扫描 DistinctNPUCardCounts 查询结果
type npuCardCountRow struct {
	CardCount int `gorm:"column:card_count"`
}

// DistinctNPUCardCounts 查询所有任务组的去重卡数列表
func (r *MetricsRepository) DistinctNPUCardCounts() ([]int, error) {
	// 子查询：按 jobs 的 node_id+pgid+start_time 分组，关联 npu_processes 查每组去重 npu_id 数
	var rows []npuCardCountRow
	err := r.db.Raw(`
		SELECT COUNT(DISTINCT np.npu_id) AS card_count
		FROM jobs j
		INNER JOIN npu_processes np ON j.node_id = np.node_id AND j.pid = np.pid AND np.status = 'running'
		GROUP BY j.node_id, j.pgid, j.start_time
		HAVING COUNT(DISTINCT np.npu_id) > 0
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	// 去重
	seen := make(map[int]bool)
	var result []int
	for _, row := range rows {
		if !seen[row.CardCount] {
			seen[row.CardCount] = true
			result = append(result, row.CardCount)
		}
	}
	return result, nil
}

// FindNPUProcessesByPID 查询单个进程占用的所有 NPU 记录
func (r *MetricsRepository) FindNPUProcessesByPID(nodeID string, pid int64) ([]model.NPUProcess, error) {
	var processes []model.NPUProcess
	err := r.db.Where("node_id = ? AND pid = ? AND status = ?", nodeID, pid, "running").Find(&processes).Error
	return processes, err
}

// FindNPUProcessesByPIDs 批量查询多个进程占用的所有 NPU 记录
func (r *MetricsRepository) FindNPUProcessesByPIDs(nodeID string, pids []int64) ([]model.NPUProcess, error) {
	if len(pids) == 0 {
		return []model.NPUProcess{}, nil
	}
	var processes []model.NPUProcess
	err := r.db.Where("node_id = ? AND pid IN ? AND status = ?", nodeID, pids, "running").Find(&processes).Error
	return processes, err
}

// FindNPUProcessesByPIDsWithStatuses 按状态过滤批量查询多个进程占用的所有 NPU 记录
func (r *MetricsRepository) FindNPUProcessesByPIDsWithStatuses(nodeID string, pids []int64, statuses []string) ([]model.NPUProcess, error) {
	if len(pids) == 0 {
		return []model.NPUProcess{}, nil
	}
	query := r.db.Where("node_id = ? AND pid IN ?", nodeID, pids)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	var processes []model.NPUProcess
	err := query.Find(&processes).Error
	return processes, err
}

// FindNPUMetricsNearTime 查询指定卡号在某个时间点之前最近的 NPU 指标
func (r *MetricsRepository) FindNPUMetricsNearTime(nodeID string, npuIDs []int, beforeMs int64) ([]model.NPUMetric, error) {
	if len(npuIDs) == 0 {
		return []model.NPUMetric{}, nil
	}

	beforeTime := time.Unix(beforeMs/1000, (beforeMs%1000)*1e6)
	var metrics []model.NPUMetric
	err := r.db.Raw(`
		SELECT m.* FROM npu_metrics m
		INNER JOIN (
			SELECT npu_id, bus_id, MAX(timestamp) AS max_ts
			FROM npu_metrics
			WHERE node_id = ? AND npu_id IN ? AND timestamp <= ?
			GROUP BY npu_id, bus_id
		) latest ON m.npu_id = latest.npu_id AND m.bus_id = latest.bus_id AND m.timestamp = latest.max_ts
		WHERE m.node_id = ?
	`, nodeID, npuIDs, beforeTime, nodeID).Scan(&metrics).Error
	return metrics, err
}

// FindNPUMetricsPeakInPeriod 查询指定卡号在时间段内 HBM 峰值对应的指标记录
func (r *MetricsRepository) FindNPUMetricsPeakInPeriod(nodeID string, npuIDs []int, startMs, endMs int64) ([]model.NPUMetric, error) {
	if len(npuIDs) == 0 {
		return []model.NPUMetric{}, nil
	}

	startTime := time.Unix(startMs/1000, (startMs%1000)*1e6)
	endTime := time.Unix(endMs/1000, (endMs%1000)*1e6)
	var metrics []model.NPUMetric
	err := r.db.Raw(`
		SELECT * FROM (
			SELECT m.*,
				ROW_NUMBER() OVER (PARTITION BY m.npu_id, m.bus_id ORDER BY m.hbm_usage_mb DESC, m.timestamp DESC) AS rn
			FROM npu_metrics m
			WHERE m.node_id = ? AND m.npu_id IN ? AND m.timestamp >= ? AND m.timestamp <= ?
		) ranked WHERE rn = 1
	`, nodeID, npuIDs, startTime, endTime).Scan(&metrics).Error
	return metrics, err
}

// FindLatestNPUMetrics 查询指定卡号的最新 NPU 指标
func (r *MetricsRepository) FindLatestNPUMetrics(nodeID string, npuIDs []int) ([]model.NPUMetric, error) {
	if len(npuIDs) == 0 {
		return []model.NPUMetric{}, nil
	}

	var metrics []model.NPUMetric
	err := r.db.Raw(`
		SELECT m.* FROM npu_metrics m
		INNER JOIN (
			SELECT npu_id, bus_id, MAX(timestamp) AS max_ts
			FROM npu_metrics
			WHERE node_id = ? AND npu_id IN ?
			GROUP BY npu_id, bus_id
		) latest ON m.npu_id = latest.npu_id AND m.bus_id = latest.bus_id AND m.timestamp = latest.max_ts
		WHERE m.node_id = ?
	`, nodeID, npuIDs, nodeID).Scan(&metrics).Error
	return metrics, err
}
