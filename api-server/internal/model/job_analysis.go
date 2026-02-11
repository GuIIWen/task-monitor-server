package model

import "time"

// JobAnalysis AI分析结果持久化
type JobAnalysis struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement"`
	JobID     string    `gorm:"column:job_id;type:varchar(255);uniqueIndex;not null"`
	Status    string    `gorm:"column:status;type:varchar(32);not null;default:'completed'"`
	Result    string    `gorm:"column:result;type:longtext;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (JobAnalysis) TableName() string {
	return "job_analysis"
}
