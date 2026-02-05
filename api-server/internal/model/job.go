package model

import "time"

// Job 作业信息
type Job struct {
	JobID       string     `gorm:"column:job_id;primaryKey" json:"jobId"`
	NodeID      *string    `gorm:"column:node_id;index" json:"nodeId"`
	HostID      *string    `gorm:"column:host_id" json:"hostId"`
	JobName     *string    `gorm:"column:job_name" json:"jobName"`
	JobType     *string    `gorm:"column:job_type" json:"jobType"`
	PID         *int64     `gorm:"column:pid" json:"pid"`
	PPID        *int64     `gorm:"column:ppid" json:"ppid"`
	PGID        *int64     `gorm:"column:pgid" json:"pgid"`
	ProcessName *string    `gorm:"column:process_name" json:"processName"`
	CommandLine *string    `gorm:"column:command_line" json:"commandLine"`
	Framework   *string    `gorm:"column:framework" json:"framework"`
	ModelFormat *string    `gorm:"column:model_format" json:"modelFormat"`
	Status      *string    `gorm:"column:status" json:"status"`
	StartTime   *int64     `gorm:"column:start_time" json:"startTime"`
	EndTime     *int64     `gorm:"column:end_time" json:"endTime"`
	CWD         *string    `gorm:"column:cwd" json:"cwd"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updatedAt"`
}

func (Job) TableName() string {
	return "jobs"
}
