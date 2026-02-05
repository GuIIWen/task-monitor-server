package repository

import (
	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
)

// CodeRepository 代码数据访问层
type CodeRepository struct {
	db *gorm.DB
}

// NewCodeRepository 创建代码Repository
func NewCodeRepository(db *gorm.DB) *CodeRepository {
	return &CodeRepository{db: db}
}

// Create 创建代码记录
func (r *CodeRepository) Create(code *model.Code) error {
	return r.db.Create(code).Error
}

// FindByJobID 根据作业ID查找代码
func (r *CodeRepository) FindByJobID(jobID string) ([]model.Code, error) {
	var codes []model.Code
	err := r.db.Where("job_id = ?", jobID).
		Order("timestamp DESC").
		Find(&codes).Error
	return codes, err
}

// FindLatestByJobID 查找作业的最新代码
func (r *CodeRepository) FindLatestByJobID(jobID string) (*model.Code, error) {
	var code model.Code
	err := r.db.Where("job_id = ?", jobID).
		Order("timestamp DESC").
		First(&code).Error
	if err != nil {
		return nil, err
	}
	return &code, nil
}

// BatchCreate 批量创建代码记录
func (r *CodeRepository) BatchCreate(codes []model.Code) error {
	return r.db.Create(&codes).Error
}
