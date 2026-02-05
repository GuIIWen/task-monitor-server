package repository

import (
	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
)

// ParameterRepository 参数数据访问层
// API Server只负责查询，不负责写入
type ParameterRepository struct {
	db *gorm.DB
}

// NewParameterRepository 创建参数Repository
func NewParameterRepository(db *gorm.DB) *ParameterRepository {
	return &ParameterRepository{db: db}
}

// FindByJobID 根据作业ID查找参数
func (r *ParameterRepository) FindByJobID(jobID string) ([]model.Parameter, error) {
	var params []model.Parameter
	err := r.db.Where("job_id = ?", jobID).Order("timestamp DESC").Find(&params).Error
	return params, err
}
