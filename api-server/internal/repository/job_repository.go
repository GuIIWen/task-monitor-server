package repository

import (
	"time"

	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
)

// JobRepository 作业数据访问层
type JobRepository struct {
	db *gorm.DB
}

// NewJobRepository 创建作业Repository
func NewJobRepository(db *gorm.DB) *JobRepository {
	return &JobRepository{db: db}
}

// Create 创建作业
func (r *JobRepository) Create(job *model.Job) error {
	return r.db.Create(job).Error
}

// Update 更新作业
func (r *JobRepository) Update(job *model.Job) error {
	return r.db.Save(job).Error
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

// UpdateStatus 更新作业状态
func (r *JobRepository) UpdateStatus(jobID, status, reason string) error {
	now := time.Now()
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 更新作业状态
		if err := tx.Model(&model.Job{}).
			Where("job_id = ?", jobID).
			Updates(map[string]interface{}{
				"status":     status,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		// 记录状态变更历史
		var oldJob model.Job
		if err := tx.Where("job_id = ?", jobID).First(&oldJob).Error; err != nil {
			return err
		}

		history := &model.JobStatusHistory{
			JobID:     &jobID,
			OldStatus: oldJob.Status,
			NewStatus: &status,
			Reason:    &reason,
			ChangedAt: now,
		}
		return tx.Create(history).Error
	})
}

// FindStaleJobs 查找超时未更新的作业
func (r *JobRepository) FindStaleJobs(timeout time.Duration) ([]model.Job, error) {
	var jobs []model.Job
	cutoff := time.Now().Add(-timeout)
	err := r.db.Where("status = ? AND updated_at < ?", "running", cutoff).Find(&jobs).Error
	return jobs, err
}

// Upsert 插入或更新作业
func (r *JobRepository) Upsert(job *model.Job) error {
	return r.db.Save(job).Error
}
