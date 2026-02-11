package repository

import (
	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JobAnalysisRepository struct {
	db *gorm.DB
}

func NewJobAnalysisRepository(db *gorm.DB) *JobAnalysisRepository {
	return &JobAnalysisRepository{db: db}
}

func (r *JobAnalysisRepository) FindByJobID(jobID string) (*model.JobAnalysis, error) {
	var analysis model.JobAnalysis
	if err := r.db.Where("job_id = ?", jobID).First(&analysis).Error; err != nil {
		return nil, err
	}
	return &analysis, nil
}

func (r *JobAnalysisRepository) FindByJobIDs(jobIDs []string) ([]model.JobAnalysis, error) {
	if len(jobIDs) == 0 {
		return []model.JobAnalysis{}, nil
	}
	var analyses []model.JobAnalysis
	if err := r.db.Where("job_id IN ?", jobIDs).Find(&analyses).Error; err != nil {
		return nil, err
	}
	return analyses, nil
}

func (r *JobAnalysisRepository) Upsert(analysis *model.JobAnalysis) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "job_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "result", "updated_at"}),
	}).Create(analysis).Error
}

func (r *JobAnalysisRepository) UpdateStatus(jobID, status, result string) error {
	updates := map[string]interface{}{"status": status, "updated_at": gorm.Expr("NOW()")}
	if result != "" {
		updates["result"] = result
	}
	return r.db.Model(&model.JobAnalysis{}).Where("job_id = ?", jobID).Updates(updates).Error
}
