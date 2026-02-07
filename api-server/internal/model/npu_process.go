package model

// NPUProcess NPU进程信息
type NPUProcess struct {
	ID            uint     `gorm:"column:id;primaryKey" json:"id"`
	NodeID        *string  `gorm:"column:node_id" json:"nodeId"`
	NPUID         *int     `gorm:"column:npu_id" json:"npuId"`
	PID           *int64   `gorm:"column:pid" json:"pid"`
	ProcessName   *string  `gorm:"column:process_name" json:"processName"`
	MemoryUsageMB *float64 `gorm:"column:memory_usage_mb" json:"memoryUsageMb"`
	Status        *string  `gorm:"column:status" json:"status"`
}

func (NPUProcess) TableName() string {
	return "npu_processes"
}
