package model

import "time"

// NPUMetric NPU指标
type NPUMetric struct {
	ID                  uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	NodeID              *string   `gorm:"column:node_id;index" json:"nodeId"`
	NPUID               *int      `gorm:"column:npu_id" json:"npuId"`
	Name                *string   `gorm:"column:name" json:"name"`
	Health              *string   `gorm:"column:health" json:"health"`
	PowerW              *float64  `gorm:"column:power_w" json:"powerW"`
	TempC               *float64  `gorm:"column:temp_c" json:"tempC"`
	AICoreUsagePercent  *float64  `gorm:"column:aicore_usage_percent" json:"aicoreUsagePercent"`
	MemoryUsageMB       *float64  `gorm:"column:memory_usage_mb" json:"memoryUsageMb"`
	MemoryTotalMB       *float64  `gorm:"column:memory_total_mb" json:"memoryTotalMb"`
	HBMUsageMB          *float64  `gorm:"column:hbm_usage_mb" json:"hbmUsageMb"`
	HBMTotalMB          *float64  `gorm:"column:hbm_total_mb" json:"hbmTotalMb"`
	BusID               *string   `gorm:"column:bus_id" json:"busId"`
	Timestamp           time.Time `gorm:"column:timestamp;index" json:"timestamp"`
}

func (NPUMetric) TableName() string {
	return "npu_metrics"
}
