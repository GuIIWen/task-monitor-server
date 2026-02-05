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

// FindAll 查找所有作业
func (r *JobRepository) FindAll() ([]model.Job, error) {
	var jobs []model.Job
	err := r.db.Find(&jobs).Error
	return jobs, err
}

// Find 灵活查询作业，支持多条件筛选和分页
// nodeID和status为空时忽略该条件
func (r *JobRepository) Find(nodeID, status string, limit, offset int) ([]model.Job, error) {
	var jobs []model.Job
	query := r.db

	if nodeID != "" {
		query = query.Where("node_id = ?", nodeID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("start_time DESC, job_id DESC").Find(&jobs).Error
	return jobs, err
}

// Count 统计符合条件的作业数量
// nodeID和status为空时忽略该条件
func (r *JobRepository) Count(nodeID, status string) (int64, error) {
	var total int64
	query := r.db.Model(&model.Job{})

	if nodeID != "" {
		query = query.Where("node_id = ?", nodeID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	return total, err
}
