package service

import (
	"github.com/task-monitor/api-server/internal/model"
	"github.com/task-monitor/api-server/internal/repository"
)

// NodeService 节点服务
type NodeService struct {
	nodeRepo repository.NodeRepositoryInterface
}

// NewNodeService 创建节点服务
func NewNodeService(nodeRepo repository.NodeRepositoryInterface) *NodeService {
	return &NodeService{
		nodeRepo: nodeRepo,
	}
}

// GetNodes 获取节点列表
func (s *NodeService) GetNodes() ([]model.Node, error) {
	return s.nodeRepo.FindAll()
}

// GetNodeByID 根据ID获取节点
func (s *NodeService) GetNodeByID(nodeID string) (*model.Node, error) {
	return s.nodeRepo.FindByID(nodeID)
}

// GetNodesByStatus 根据状态获取节点
func (s *NodeService) GetNodesByStatus(status string) ([]model.Node, error) {
	return s.nodeRepo.FindByStatus(status)
}
