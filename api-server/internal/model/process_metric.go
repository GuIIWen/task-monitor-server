package model

import "time"

// ProcessMetric 进程指标
type ProcessMetric struct {
	ID          uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	JobID       *string   `gorm:"column:job_id;index" json:"jobId"`
	PID         *int      `gorm:"column:pid" json:"pid"`
	CPUPercent  *float64  `gorm:"column:cpu_percent" json:"cpuPercent"`
	MemoryMB    *float64  `gorm:"column:memory_mb" json:"memoryMb"`
	ThreadCount *int      `gorm:"column:thread_count" json:"threadCount"`
	OpenFiles   *int      `gorm:"column:open_files" json:"openFiles"`
	Status      *string   `gorm:"column:status" json:"status"`
	Timestamp   time.Time `gorm:"column:timestamp" json:"timestamp"`
}

func (ProcessMetric) TableName() string {
	return "process_metrics"
}
