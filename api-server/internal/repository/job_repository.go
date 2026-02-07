package repository

import (
	"fmt"

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
	NodeID    string `gorm:"column:node_id"`
	PGID      int64  `gorm:"column:pgid"`
	StartTime int64  `gorm:"column:start_time"`
}

// CountGroups 统计按 node_id+pgid+start_time 分组后的组数
func (r *JobRepository) CountGroups(nodeID string, statuses []string, jobTypes []string, frameworks []string, cardCounts []int) (int64, error) {
	var total int64
	subQuery := r.db.Model(&model.Job{}).Select("node_id, pgid, start_time, COUNT(*) AS card_count")
	subQuery = r.applyFilters(subQuery, nodeID, statuses, jobTypes, frameworks)
	subQuery = subQuery.Group("node_id, pgid, start_time")
	if len(cardCounts) > 0 {
		subQuery = subQuery.Having("COUNT(*) IN ?", cardCounts)
	}

	err := r.db.Table("(?) AS job_groups", subQuery).Count(&total).Error
	return total, err
}

// FindGrouped 按 node_id+pgid 分组查询作业
// 两阶段查询：先查分组列表（带分页），再查各组下的所有作业
func (r *JobRepository) FindGrouped(nodeID string, statuses []string, jobTypes []string, frameworks []string, cardCounts []int, sortBy, sortOrder string, limit, offset int) ([]model.Job, error) {
	// 阶段1：查出分页后的分组 (node_id, pgid) 列表
	orderClause := "MIN(start_time) DESC"
	if sortBy == "cardCount" {
		dir := "ASC"
		if sortOrder == "desc" {
			dir = "DESC"
		}
		orderClause = fmt.Sprintf("COUNT(*) %s", dir)
	} else if col, ok := allowedSortColumns[sortBy]; ok {
		dir := "ASC"
		if sortOrder == "desc" {
			dir = "DESC"
		}
		orderClause = fmt.Sprintf("MIN(%s) %s", col, dir)
	}

	groupQuery := r.db.Model(&model.Job{}).Select("node_id, pgid, start_time")
	groupQuery = r.applyFilters(groupQuery, nodeID, statuses, jobTypes, frameworks)
	groupQuery = groupQuery.Group("node_id, pgid, start_time")
	if len(cardCounts) > 0 {
		groupQuery = groupQuery.Having("COUNT(*) IN ?", cardCounts)
	}
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
		if i == 0 {
			orConditions = orConditions.Where("(node_id = ? AND pgid = ? AND start_time = ?)", g.NodeID, g.PGID, g.StartTime)
		} else {
			orConditions = orConditions.Or("(node_id = ? AND pgid = ? AND start_time = ?)", g.NodeID, g.PGID, g.StartTime)
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

// DistinctCardCounts 获取所有去重的卡数值
func (r *JobRepository) DistinctCardCounts() ([]int, error) {
	var counts []int
	err := r.db.Model(&model.Job{}).
		Select("COUNT(*) AS card_count").
		Group("node_id, pgid, start_time").
		Pluck("card_count", &counts).Error
	if err != nil {
		return nil, err
	}
	// 去重
	seen := make(map[int]bool)
	var result []int
	for _, c := range counts {
		if !seen[c] {
			seen[c] = true
			result = append(result, c)
		}
	}
	return result, nil
}
