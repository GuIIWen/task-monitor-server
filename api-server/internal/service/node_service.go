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

// GetNodeStats 获取节点统计信息
func (s *NodeService) GetNodeStats() (map[string]int64, error) {
	nodes, err := s.nodeRepo.FindAll()
	if err != nil {
		return nil, err
	}

	stats := map[string]int64{
		"total":    int64(len(nodes)),
		"active":   0,
		"inactive": 0,
		"error":    0,
	}

	for _, node := range nodes {
		if node.Status != nil {
			switch *node.Status {
			case "active":
				stats["active"]++
			case "inactive":
				stats["inactive"]++
			case "error":
				stats["error"]++
			}
		}
	}

	return stats, nil
}
