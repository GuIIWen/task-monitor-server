package model

import "time"

// JobStatusHistory 作业状态变更历史
type JobStatusHistory struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	JobID     *string   `gorm:"column:job_id;index" json:"jobId"`
	OldStatus *string   `gorm:"column:old_status" json:"oldStatus"`
	NewStatus *string   `gorm:"column:new_status" json:"newStatus"`
	Reason    *string   `gorm:"column:reason" json:"reason"`
	ChangedAt time.Time `gorm:"column:changed_at" json:"changedAt"`
}

func (JobStatusHistory) TableName() string {
	return "job_status_histories"
}
