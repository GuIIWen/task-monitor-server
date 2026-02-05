package repository

import (
	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
)

// ParameterRepository 参数数据访问层
type ParameterRepository struct {
	db *gorm.DB
}

// NewParameterRepository 创建参数Repository
func NewParameterRepository(db *gorm.DB) *ParameterRepository {
	return &ParameterRepository{db: db}
}

// Create 创建参数记录
func (r *ParameterRepository) Create(param *model.Parameter) error {
	return r.db.Create(param).Error
}

// FindByJobID 根据作业ID查找参数
func (r *ParameterRepository) FindByJobID(jobID string) ([]model.Parameter, error) {
	var params []model.Parameter
	err := r.db.Where("job_id = ?", jobID).
		Order("timestamp DESC").
		Find(&params).Error
	return params, err
}

// FindLatestByJobID 查找作业的最新参数
func (r *ParameterRepository) FindLatestByJobID(jobID string) (*model.Parameter, error) {
	var param model.Parameter
	err := r.db.Where("job_id = ?", jobID).
		Order("timestamp DESC").
		First(&param).Error
	if err != nil {
		return nil, err
	}
	return &param, nil
}

// BatchCreate 批量创建参数记录
func (r *ParameterRepository) BatchCreate(params []model.Parameter) error {
	return r.db.Create(&params).Error
}
