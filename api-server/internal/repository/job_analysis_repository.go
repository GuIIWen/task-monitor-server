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

func (r *JobAnalysisRepository) Upsert(analysis *model.JobAnalysis) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "job_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"result", "updated_at"}),
	}).Create(analysis).Error
}
