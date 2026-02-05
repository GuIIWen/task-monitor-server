package model

import "time"

// Node 节点信息
type Node struct {
	NodeID        string     `gorm:"column:node_id;primaryKey" json:"nodeId"`
	HostID        *string    `gorm:"column:host_id" json:"hostId"`
	Hostname      *string    `gorm:"column:hostname" json:"hostname"`
	IPAddress     *string    `gorm:"column:ip_address" json:"ipAddress"`
	NPUCount      *int       `gorm:"column:npu_count" json:"npuCount"`
	Status        *string    `gorm:"column:status" json:"status"`
	LastHeartbeat *time.Time `gorm:"column:last_heartbeat" json:"lastHeartbeat"`
	CreatedAt     time.Time  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at" json:"updatedAt"`
}

func (Node) TableName() string {
	return "nodes"
}
