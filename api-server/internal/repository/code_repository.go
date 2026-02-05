package repository

import (
	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
)

// CodeRepository 代码数据访问层
// API Server只负责查询，不负责写入
type CodeRepository struct {
	db *gorm.DB
}

// NewCodeRepository 创建代码Repository
func NewCodeRepository(db *gorm.DB) *CodeRepository {
	return &CodeRepository{db: db}
}

// FindByJobID 根据作业ID查找代码
func (r *CodeRepository) FindByJobID(jobID string) ([]model.Code, error) {
	var codes []model.Code
	err := r.db.Where("job_id = ?", jobID).Order("timestamp DESC").Find(&codes).Error
	return codes, err
}
