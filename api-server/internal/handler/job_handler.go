package handler

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/task-monitor/api-server/internal/service"
	"github.com/task-monitor/api-server/internal/utils"
	"gorm.io/gorm"
)

// JobHandler 作业处理器
type JobHandler struct {
	jobService service.JobServiceInterface
}

// NewJobHandler 创建作业处理器
func NewJobHandler(jobService service.JobServiceInterface) *JobHandler {
	return &JobHandler{
		jobService: jobService,
	}
}

// GetJobs 获取作业列表
// 支持多条件筛选：nodeId和status可以单独使用或组合使用
// 无参数时返回全量查询
func (h *JobHandler) GetJobs(c *gin.Context) {
	status := c.Query("status")
	nodeID := c.Query("nodeId")

	// 使用灵活查询方法，支持多条件筛选和全量查询
	jobs, err := h.jobService.GetJobs(nodeID, status)
	if err != nil {
		utils.ErrorResponse(c, 500, "Database error: "+err.Error())
		return
	}

	utils.SuccessResponse(c, jobs)
}

// GetJobByID 获取作业详情
func (h *JobHandler) GetJobByID(c *gin.Context) {
	jobID := c.Param("jobId")

	job, err := h.jobService.GetJobByID(jobID)
	if err != nil {
		// 区分记录不存在和数据库错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, 404, "Job not found")
		} else {
			utils.ErrorResponse(c, 500, "Database error: "+err.Error())
		}
		return
	}

	utils.SuccessResponse(c, job)
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
