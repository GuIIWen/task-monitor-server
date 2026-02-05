package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/task-monitor/api-server/internal/service"
	"github.com/task-monitor/api-server/internal/utils"
	"gorm.io/gorm"
)

// NodeHandler 节点处理器
type NodeHandler struct {
	nodeService service.NodeServiceInterface
}

// NewNodeHandler 创建节点处理器
func NewNodeHandler(nodeService service.NodeServiceInterface) *NodeHandler {
	return &NodeHandler{
		nodeService: nodeService,
	}
}

// GetNodes 获取节点列表
func (h *NodeHandler) GetNodes(c *gin.Context) {
	status := c.Query("status")

	var nodes interface{}
	var err error

	if status != "" {
		nodes, err = h.nodeService.GetNodesByStatus(status)
	} else {
		nodes, err = h.nodeService.GetNodes()
	}

	if err != nil {
		utils.ErrorResponse(c, 500, err.Error())
		return
	}

	utils.SuccessResponse(c, nodes)
}

// GetNodeByID 获取节点详情
func (h *NodeHandler) GetNodeByID(c *gin.Context) {
	nodeID := c.Param("nodeId")

	node, err := h.nodeService.GetNodeByID(nodeID)
	if err != nil {
		// 区分记录不存在和数据库错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, 404, "Node not found")
		} else {
			utils.ErrorResponse(c, 500, "Database error: "+err.Error())
		}
		return
	}

	utils.SuccessResponse(c, node)
}

// GetNodeStats 获取节点统计信息
func (h *NodeHandler) GetNodeStats(c *gin.Context) {
	stats, err := h.nodeService.GetNodeStats()
	if err != nil {
		utils.ErrorResponse(c, 500, "Database error: "+err.Error())
		return
	}

	utils.SuccessResponse(c, stats)
}
