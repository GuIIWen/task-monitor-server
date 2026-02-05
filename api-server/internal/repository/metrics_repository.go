package repository

import (
	"time"

	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
)

// MetricsRepository 指标数据访问层
type MetricsRepository struct {
	db *gorm.DB
}

// NewMetricsRepository 创建指标Repository
func NewMetricsRepository(db *gorm.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

// CreateNPUMetric 创建NPU指标
func (r *MetricsRepository) CreateNPUMetric(metric *model.NPUMetric) error {
	return r.db.Create(metric).Error
}

// BatchCreateNPUMetrics 批量创建NPU指标
func (r *MetricsRepository) BatchCreateNPUMetrics(metrics []model.NPUMetric) error {
	if len(metrics) == 0 {
		return nil
	}
	return r.db.Create(&metrics).Error
}

// FindNPUMetricsByNodeID 根据节点ID查找NPU指标
func (r *MetricsRepository) FindNPUMetricsByNodeID(nodeID string, startTime, endTime time.Time) ([]model.NPUMetric, error) {
	var metrics []model.NPUMetric
	query := r.db.Where("node_id = ?", nodeID)

	if !startTime.IsZero() {
		query = query.Where("timestamp >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("timestamp <= ?", endTime)
	}

	err := query.Order("timestamp DESC").Find(&metrics).Error
	return metrics, err
}

// CreateProcessMetric 创建进程指标
func (r *MetricsRepository) CreateProcessMetric(metric *model.ProcessMetric) error {
	return r.db.Create(metric).Error
}

// BatchCreateProcessMetrics 批量创建进程指标
func (r *MetricsRepository) BatchCreateProcessMetrics(metrics []model.ProcessMetric) error {
	if len(metrics) == 0 {
		return nil
	}
	return r.db.Create(&metrics).Error
}

// FindProcessMetricsByJobID 根据作业ID查找进程指标
func (r *MetricsRepository) FindProcessMetricsByJobID(jobID string, startTime, endTime time.Time) ([]model.ProcessMetric, error) {
	var metrics []model.ProcessMetric
	query := r.db.Where("job_id = ?", jobID)

	if !startTime.IsZero() {
		query = query.Where("timestamp >= ?", startTime)
	}
	if !endTime.IsZero() {
		query = query.Where("timestamp <= ?", endTime)
	}

	err := query.Order("timestamp DESC").Find(&metrics).Error
	return metrics, err
}

// FindLatestProcessMetricByJobID 查找作业的最新进程指标
func (r *MetricsRepository) FindLatestProcessMetricByJobID(jobID string) (*model.ProcessMetric, error) {
	var metric model.ProcessMetric
	err := r.db.Where("job_id = ?", jobID).
		Order("timestamp DESC").
		First(&metric).Error
	if err != nil {
		return nil, err
	}
	return &metric, nil
}
