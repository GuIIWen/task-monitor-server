package repository

import (
	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
)

// NodeRepository 节点数据访问层
// API Server只负责查询，不负责写入
type NodeRepository struct {
	db *gorm.DB
}

// NewNodeRepository 创建节点Repository
func NewNodeRepository(db *gorm.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

// FindByID 根据ID查找节点
func (r *NodeRepository) FindByID(nodeID string) (*model.Node, error) {
	var node model.Node
	err := r.db.Where("node_id = ?", nodeID).First(&node).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// FindAll 查找所有节点
func (r *NodeRepository) FindAll() ([]model.Node, error) {
	var nodes []model.Node
	err := r.db.Find(&nodes).Error
	return nodes, err
}

// FindByStatus 根据状态查找节点
func (r *NodeRepository) FindByStatus(status string) ([]model.Node, error) {
	var nodes []model.Node
	err := r.db.Where("status = ?", status).Find(&nodes).Error
	return nodes, err
}
