package repository

import (
	"fmt"
	"strings"

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

// applyFilters 应用通用筛选条件
func (r *JobRepository) applyFilters(query *gorm.DB, nodeID string, statuses []string, jobTypes []string, frameworks []string) *gorm.DB {
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
	return query
}

// groupKey 分组键
type groupKey struct {
	NodeID    *string `gorm:"column:node_id"`
	PGID      *int64  `gorm:"column:pgid"`
	StartTime *int64  `gorm:"column:start_time"`
}

// buildNullableGroupCondition 构建支持 NULL 值的分组匹配条件
func buildNullableGroupCondition(g groupKey) (string, []interface{}) {
	parts := make([]string, 0, 3)
	args := make([]interface{}, 0, 3)

	if g.NodeID == nil {
		parts = append(parts, "node_id IS NULL")
	} else {
		parts = append(parts, "node_id = ?")
		args = append(args, *g.NodeID)
	}

	if g.PGID == nil {
		parts = append(parts, "pgid IS NULL")
	} else {
		parts = append(parts, "pgid = ?")
		args = append(args, *g.PGID)
	}

	if g.StartTime == nil {
		parts = append(parts, "start_time IS NULL")
	} else {
		parts = append(parts, "start_time = ?")
		args = append(args, *g.StartTime)
	}

	return "(" + strings.Join(parts, " AND ") + ")", args
}

// CountGroups 统计按 node_id+pgid+start_time 分组后的组数
func (r *JobRepository) CountGroups(nodeID string, statuses []string, jobTypes []string, frameworks []string) (int64, error) {
	var total int64
	subQuery := r.db.Model(&model.Job{}).Select("node_id, pgid, start_time")
	subQuery = r.applyFilters(subQuery, nodeID, statuses, jobTypes, frameworks)
	subQuery = subQuery.Group("node_id, pgid, start_time")

	err := r.db.Table("(?) AS job_groups", subQuery).Count(&total).Error
	return total, err
}

// FindGrouped 按 node_id+pgid 分组查询作业
// 两阶段查询：先查分组列表（带分页），再查各组下的所有作业
func (r *JobRepository) FindGrouped(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string, limit, offset int) ([]model.Job, error) {
	// 阶段1：查出分页后的分组 (node_id, pgid, start_time) 列表
	orderClause := "MIN(start_time) DESC, node_id ASC, pgid ASC, start_time ASC"
	if col, ok := allowedSortColumns[sortBy]; ok {
		dir := "ASC"
		if sortOrder == "desc" {
			dir = "DESC"
		}
		orderClause = fmt.Sprintf("MIN(%s) %s, node_id ASC, pgid ASC, start_time ASC", col, dir)
	}

	groupQuery := r.db.Model(&model.Job{}).Select("node_id, pgid, start_time")
	groupQuery = r.applyFilters(groupQuery, nodeID, statuses, jobTypes, frameworks)
	groupQuery = groupQuery.Group("node_id, pgid, start_time")
	groupQuery = groupQuery.Order(orderClause)

	if limit > 0 {
		groupQuery = groupQuery.Limit(limit)
	}
	if offset > 0 {
		groupQuery = groupQuery.Offset(offset)
	}

	var groups []groupKey
	if err := groupQuery.Find(&groups).Error; err != nil {
		return nil, err
	}

	if len(groups) == 0 {
		return []model.Job{}, nil
	}

	// 阶段2：构建 OR 条件查出这些分组下的所有作业
	query := r.db.Session(&gorm.Session{})
	orConditions := r.db.Session(&gorm.Session{})
	for i, g := range groups {
		condition, args := buildNullableGroupCondition(g)
		if i == 0 {
			orConditions = orConditions.Where(condition, args...)
		} else {
			orConditions = orConditions.Or(condition, args...)
		}
	}
	query = query.Where(orConditions)
	query = r.applyFilters(query, nodeID, statuses, jobTypes, frameworks)
	query = query.Order("node_id, pgid, start_time, pid ASC")

	var jobs []model.Job
	if err := query.Find(&jobs).Error; err != nil {
		return nil, err
	}

	return jobs, nil
}
