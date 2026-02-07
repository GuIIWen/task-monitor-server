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

// FindByNodeIDAndPGID 根据节点ID和进程组ID查找作业
func (r *JobRepository) FindByNodeIDAndPGID(nodeID string, pgid int64) ([]model.Job, error) {
	var jobs []model.Job
	err := r.db.Where("node_id = ? AND pgid = ?", nodeID, pgid).Find(&jobs).Error
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

// allowedSortColumns 允许排序的列名映射（前端字段名 -> 数据库列名）
var allowedSortColumns = map[string]string{
	"jobName":   "job_name",
	"jobType":   "job_type",
	"framework": "framework",
	"nodeId":    "node_id",
	"status":    "status",
	"startTime": "start_time",
}

// Find 灵活查询作业，支持多条件筛选、排序和分页
func (r *JobRepository) Find(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string, limit, offset int) ([]model.Job, error) {
	var jobs []model.Job
	query := r.db

	if nodeID != "" {
		query = query.Where("node_id = ?", nodeID)
	}
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	if len(jobTypes) > 0 {
		query = query.Where("job_type IN ?", jobTypes)
	}
	if len(frameworks) > 0 {
		query = query.Where("framework IN ?", frameworks)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// 排序：验证字段名防止SQL注入
	orderClause := "start_time DESC, job_id DESC"
	if col, ok := allowedSortColumns[sortBy]; ok {
		dir := "ASC"
		if sortOrder == "desc" {
			dir = "DESC"
		}
		orderClause = col + " " + dir + ", job_id DESC"
	}

	err := query.Order(orderClause).Find(&jobs).Error
	return jobs, err
}

// Count 统计符合条件的作业数量
func (r *JobRepository) Count(nodeID string, statuses []string, jobTypes []string, frameworks []string) (int64, error) {
	var total int64
	query := r.db.Model(&model.Job{})

	if nodeID != "" {
		query = query.Where("node_id = ?", nodeID)
	}
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	if len(jobTypes) > 0 {
		query = query.Where("job_type IN ?", jobTypes)
	}
	if len(frameworks) > 0 {
		query = query.Where("framework IN ?", frameworks)
	}

	err := query.Count(&total).Error
	return total, err
}

// FindFiltered 查出所有符合筛选条件的 jobs（不分页，不分组），用于 service 层在内存中构建进程树
func (r *JobRepository) FindFiltered(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string) ([]model.Job, error) {
	var jobs []model.Job
	query := r.db

	if nodeID != "" {
		query = query.Where("node_id = ?", nodeID)
	}
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	if len(jobTypes) > 0 {
		query = query.Where("job_type IN ?", jobTypes)
	}
	if len(frameworks) > 0 {
		query = query.Where("framework IN ?", frameworks)
	}

	orderClause := "start_time DESC, job_id DESC"
	if col, ok := allowedSortColumns[sortBy]; ok {
		dir := "ASC"
		if sortOrder == "desc" {
			dir = "DESC"
		}
		orderClause = col + " " + dir + ", job_id DESC"
	}

	err := query.Order(orderClause).Find(&jobs).Error
	return jobs, err
}