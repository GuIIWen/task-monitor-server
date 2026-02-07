package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/task-monitor/api-server/internal/service"
	"github.com/task-monitor/api-server/internal/utils"
	"gorm.io/gorm"
)

// JobHandler 作业处理器
type JobHandler struct {
	jobService service.JobServiceInterface
	llmService service.LLMServiceInterface
}

// NewJobHandler 创建作业处理器
func NewJobHandler(jobService service.JobServiceInterface, llmService service.LLMServiceInterface) *JobHandler {
	return &JobHandler{
		jobService: jobService,
		llmService: llmService,
	}
}

// GetJobs 获取作业列表
// 支持多条件筛选：nodeId、status、type、framework可以单独使用或组合使用
// 支持排序：sortBy指定排序字段，sortOrder指定排序方向(asc/desc)
func (h *JobHandler) GetJobs(c *gin.Context) {
	nodeID := c.Query("nodeId")
	statuses := c.QueryArray("status")
	jobTypes := c.QueryArray("type")
	frameworks := c.QueryArray("framework")
	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if err != nil || pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	jobs, total, err := h.jobService.GetJobs(nodeID, statuses, jobTypes, frameworks, sortBy, sortOrder, page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, 500, "Database error: "+err.Error())
		return
	}

	totalPages := int64(0)
	if pageSize > 0 {
		totalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	}

	utils.SuccessResponse(c, utils.PaginationResponse{
		Items: jobs,
		Pagination: utils.Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GetJobByID 获取作业详情（含 NPU 卡信息和关联进程）
func (h *JobHandler) GetJobByID(c *gin.Context) {
	jobID := c.Param("jobId")

	detail, err := h.jobService.GetJobDetail(jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, 404, "Job not found")
		} else {
			utils.ErrorResponse(c, 500, "Database error: "+err.Error())
		}
		return
	}

	utils.SuccessResponse(c, detail)
}

// GetJobParameters 获取作业参数
func (h *JobHandler) GetJobParameters(c *gin.Context) {
	jobID := c.Param("jobId")

	params, err := h.jobService.GetJobParameters(jobID)
	if err != nil {
		utils.ErrorResponse(c, 500, err.Error())
		return
	}

	utils.SuccessResponse(c, params)
}

// GetJobCode 获取作业代码
func (h *JobHandler) GetJobCode(c *gin.Context) {
	jobID := c.Param("jobId")

	code, err := h.jobService.GetJobCode(jobID)
	if err != nil {
		utils.ErrorResponse(c, 500, err.Error())
		return
	}

	utils.SuccessResponse(c, code)
}

// GetGroupedJobs 获取分组作业列表（按 node_id+pgid 分组）
func (h *JobHandler) GetGroupedJobs(c *gin.Context) {
	nodeID := c.Query("nodeId")
	statuses := c.QueryArray("status")
	jobTypes := c.QueryArray("type")
	frameworks := c.QueryArray("framework")
	cardCountStrs := c.QueryArray("cardCount")
	var cardCounts []int
	for _, s := range cardCountStrs {
		if s == "unknown" {
			// unknown 用 0 表示，service 层会匹配 CardCount == nil
			cardCounts = append(cardCounts, 0)
		} else if v, err := strconv.Atoi(s); err == nil {
			cardCounts = append(cardCounts, v)
		}
	}
	sortBy := c.Query("sortBy")
	sortOrder := c.Query("sortOrder")

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if err != nil || pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	groups, total, err := h.jobService.GetGroupedJobs(nodeID, statuses, jobTypes, frameworks, cardCounts, sortBy, sortOrder, page, pageSize)
	if err != nil {
		utils.ErrorResponse(c, 500, "Database error: "+err.Error())
		return
	}

	totalPages := int64(0)
	if pageSize > 0 {
		totalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	}

	utils.SuccessResponse(c, utils.PaginationResponse{
		Items: groups,
		Pagination: utils.Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// GetDistinctCardCounts 获取所有去重的卡数值
func (h *JobHandler) GetDistinctCardCounts(c *gin.Context) {
	counts, err := h.jobService.GetDistinctCardCounts()
	if err != nil {
		utils.ErrorResponse(c, 500, "Database error: "+err.Error())
		return
	}
	utils.SuccessResponse(c, counts)
}

// GetJobStats 获取作业统计信息
func (h *JobHandler) GetJobStats(c *gin.Context) {
	stats, err := h.jobService.GetJobStats()
	if err != nil {
		utils.ErrorResponse(c, 500, "Database error: "+err.Error())
		return
	}

	utils.SuccessResponse(c, stats)
}

// AnalyzeJob AI分析作业
func (h *JobHandler) AnalyzeJob(c *gin.Context) {
	if h.llmService == nil {
		utils.ErrorResponse(c, 501, "LLM service is not configured")
		return
	}

	jobID := c.Param("jobId")

	result, err := h.llmService.AnalyzeJob(jobID)
	if err != nil {
		utils.ErrorResponse(c, 500, "AI analysis failed: "+err.Error())
		return
	}

	utils.SuccessResponse(c, result)
}
