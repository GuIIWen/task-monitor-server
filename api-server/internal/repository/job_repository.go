package repository

import (
	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
)

// JobRepository 作业数据访问层
// API Server只负责查询，不负责写入
type JobRepository struct {
	db *gorm.DB
}

// NewJobRepository 创建作业Repository
func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

// FindByID 根据ID查找作业
func (r *JobRepository) FindByID(jobID string) (*model.Job, error) {
	var job model.Job
	err := r.db.Where("job_id = ?", jobID).First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// FindByNodeID 根据节点ID查找作业
func (r *JobRepository) FindByNodeID(nodeID string) ([]model.Job, error) {
	var jobs []model.Job
	err := r.db.Where("node_id = ?", nodeID).Find(&jobs).Error
	return jobs, err
}

// FindByStatus 根据状态查找作业
func (r *JobRepository) FindByStatus(status string) ([]model.Job, error) {
	var jobs []model.Job
	err := r.db.Where("status = ?", status).Find(&jobs).Error
	return jobs, err
}
