package repository

import (
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

// 暂时没有查询方法，如果需要可以添加
// 例如：FindNPUMetricsByJobID, FindProcessMetricsByJobID 等
