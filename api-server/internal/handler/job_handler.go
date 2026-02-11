package handler

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/task-monitor/api-server/internal/service"
	"github.com/task-monitor/api-server/internal/utils"
	"gorm.io/gorm"
)

// failedItem 记录单个失败作业的信息
type failedItem struct {
	JobID string `json:"jobId"`
	Error string `json:"error"`
}

// batchAnalyzeState 批量分析任务状态
type batchAnalyzeState struct {
	Status      string `json:"status"` // running / done / cancelled
	Total       int    `json:"total"`
	Current     int64  `json:"current"`
	Success     int64  `json:"success"`
	Failed      int64  `json:"failed"`
	FailedItems []failedItem
	mu          sync.Mutex
	cancelCh    chan struct{}
}

var (
	batchStates   = sync.Map{} // map[string]*batchAnalyzeState
	batchIDSeq    int64
)

// JobHandler 作业处理器
type JobHandler struct {
	jobService       service.JobServiceInterface
	llmService       service.LLMServiceInterface
	batchConcurrency int
}

// NewJobHandler 创建作业处理器
func NewJobHandler(jobService service.JobServiceInterface, llmService service.LLMServiceInterface, batchConcurrency int) *JobHandler {
	if batchConcurrency <= 0 {
		batchConcurrency = 5
	}
	return &JobHandler{
		jobService:       jobService,
		llmService:       llmService,
		batchConcurrency: batchConcurrency,
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
	aggregate := c.DefaultQuery("aggregate", "true") != "false"

	detail, err := h.jobService.GetJobDetail(jobID, aggregate)
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

// AnalyzeJob AI分析作业（异步）
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

// GetJobAnalysis 获取已保存的AI分析结果
func (h *JobHandler) GetJobAnalysis(c *gin.Context) {
	if h.llmService == nil {
		utils.ErrorResponse(c, 501, "LLM service is not configured")
		return
	}

	jobID := c.Param("jobId")

	result, err := h.llmService.GetAnalysis(jobID)
	if err != nil {
		utils.ErrorResponse(c, 500, "Failed to get analysis: "+err.Error())
		return
	}

	utils.SuccessResponse(c, result)
}

// BatchAnalyze 批量AI分析作业
func (h *JobHandler) BatchAnalyze(c *gin.Context) {
	if h.llmService == nil {
		utils.ErrorResponse(c, 501, "LLM service is not configured")
		return
	}

	var req struct {
		JobIDs []string `json:"jobIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.JobIDs) == 0 {
		utils.ErrorResponse(c, 400, "jobIds is required")
		return
	}

	batchID := fmt.Sprintf("batch-%d-%d", time.Now().UnixMilli(), atomic.AddInt64(&batchIDSeq, 1))
	state := &batchAnalyzeState{
		Status:   "running",
		Total:    len(req.JobIDs),
		cancelCh: make(chan struct{}),
	}
	batchStates.Store(batchID, state)

	go func() {
		sem := make(chan struct{}, h.batchConcurrency)
		var wg sync.WaitGroup
		for _, jobID := range req.JobIDs {
			// 检查是否已取消
			select {
			case <-state.cancelCh:
				// 已取消，不再提交新任务
				break
			default:
			}
			// 再次检查（select default 不会 break 外层 for）
			cancelled := false
			select {
			case <-state.cancelCh:
				cancelled = true
			default:
			}
			if cancelled {
				break
			}

			wg.Add(1)
			sem <- struct{}{}
			go func(id string) {
				defer wg.Done()
				defer func() { <-sem }()
				// worker 内也检查取消
				select {
				case <-state.cancelCh:
					return
				default:
				}
				if _, err := h.llmService.AnalyzeJob(id); err != nil {
					atomic.AddInt64(&state.Failed, 1)
					state.mu.Lock()
					state.FailedItems = append(state.FailedItems, failedItem{JobID: id, Error: err.Error()})
					state.mu.Unlock()
				} else {
					atomic.AddInt64(&state.Success, 1)
				}
				atomic.AddInt64(&state.Current, 1)
			}(jobID)
		}
		wg.Wait()
		// 判断最终状态
		select {
		case <-state.cancelCh:
			state.Status = "cancelled"
		default:
			state.Status = "done"
		}
	}()

	utils.SuccessResponse(c, gin.H{"batchId": batchID})
}

// GetBatchAnalyses 批量获取分析摘要
func (h *JobHandler) GetBatchAnalyses(c *gin.Context) {
	jobIDs := c.QueryArray("jobIds")
	if len(jobIDs) == 0 {
		utils.SuccessResponse(c, gin.H{})
		return
	}
	result, err := h.llmService.GetBatchAnalyses(jobIDs)
	if err != nil {
		utils.ErrorResponse(c, 500, "failed to fetch analyses: "+err.Error())
		return
	}
	utils.SuccessResponse(c, result)
}

// GetBatchAnalyzeProgress 查询批量分析进度
func (h *JobHandler) GetBatchAnalyzeProgress(c *gin.Context) {
	batchID := c.Param("batchId")
	val, ok := batchStates.Load(batchID)
	if !ok {
		utils.ErrorResponse(c, 404, "batch not found")
		return
	}
	state := val.(*batchAnalyzeState)
	state.mu.Lock()
	failedItems := make([]failedItem, len(state.FailedItems))
	copy(failedItems, state.FailedItems)
	state.mu.Unlock()

	utils.SuccessResponse(c, gin.H{
		"status":      state.Status,
		"total":       state.Total,
		"current":     atomic.LoadInt64(&state.Current),
		"success":     atomic.LoadInt64(&state.Success),
		"failed":      atomic.LoadInt64(&state.Failed),
		"failedItems": failedItems,
	})
}

// CancelBatchAnalyze 取消批量分析任务
func (h *JobHandler) CancelBatchAnalyze(c *gin.Context) {
	batchID := c.Param("batchId")
	val, ok := batchStates.Load(batchID)
	if !ok {
		utils.ErrorResponse(c, 404, "batch not found")
		return
	}
	state := val.(*batchAnalyzeState)
	if state.Status != "running" {
		utils.ErrorResponse(c, 400, "batch is not running")
		return
	}
	close(state.cancelCh)
	utils.SuccessResponse(c, gin.H{"message": "cancel signal sent"})
}
