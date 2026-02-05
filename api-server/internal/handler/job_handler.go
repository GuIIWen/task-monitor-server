package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/task-monitor/api-server/internal/service"
	"github.com/task-monitor/api-server/internal/utils"
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
func (h *JobHandler) GetJobs(c *gin.Context) {
	status := c.Query("status")
	nodeID := c.Query("nodeId")

	var jobs interface{}
	var err error

	if nodeID != "" {
		jobs, err = h.jobService.GetJobsByNodeID(nodeID)
	} else if status != "" {
		jobs, err = h.jobService.GetJobsByStatus(status)
	} else {
		// TODO: 实现完整的作业列表查询（带分页）
		utils.ErrorResponse(c, 400, "Please provide nodeId or status parameter")
		return
	}

	if err != nil {
		utils.ErrorResponse(c, 500, err.Error())
		return
	}

	utils.SuccessResponse(c, jobs)
}

// GetJobByID 获取作业详情
func (h *JobHandler) GetJobByID(c *gin.Context) {
	jobID := c.Param("jobId")

	job, err := h.jobService.GetJobByID(jobID)
	if err != nil {
		utils.ErrorResponse(c, 404, "Job not found")
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
