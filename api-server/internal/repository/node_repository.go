package repository

import (
	"github.com/task-monitor/api-server/internal/model"
	"gorm.io/gorm"
)

// NodeRepository 节点数据访问层
type NodeRepository struct {
	db *gorm.DB
}

// NewNodeRepository 创建节点Repository
func NewNodeRepository(db *gorm.DB) *NodeRepository {
	return &NodeRepository{db: db}
}

// Create 创建节点
func (r *NodeRepository) Create(node *model.Node) error {
	return r.db.Create(node).Error
}

// Update 更新节点
func (r *NodeRepository) Update(node *model.Node) error {
	return r.db.Save(node).Error
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

// UpdateHeartbeat 更新节点心跳时间
func (r *NodeRepository) UpdateHeartbeat(nodeID string) error {
	return r.db.Model(&model.Node{}).
		Where("node_id = ?", nodeID).
		Update("last_heartbeat", gorm.Expr("NOW()")).Error
}

// Upsert 插入或更新节点
func (r *NodeRepository) Upsert(node *model.Node) error {
	return r.db.Save(node).Error
}
