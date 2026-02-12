package handler

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/task-monitor/api-server/internal/model"
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
	batchStates = sync.Map{} // map[string]*batchAnalyzeState
	batchIDSeq  int64
)

// JobHandler 作业处理器
type JobHandler struct {
	jobService       service.JobServiceInterface
	llmService       service.LLMServiceInterface
	batchConcurrency int
}

// NewJobHandler 创建作业处理器
// 兼容旧调用方式：未传 batchConcurrency 时默认 5。
func NewJobHandler(jobService service.JobServiceInterface, llmService service.LLMServiceInterface, batchConcurrency ...int) *JobHandler {
	concurrency := 5
	if len(batchConcurrency) > 0 && batchConcurrency[0] > 0 {
		concurrency = batchConcurrency[0]
	}
	return &JobHandler{
		jobService:       jobService,
		llmService:       llmService,
		batchConcurrency: concurrency,
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
	var req struct {
		ModelID string `json:"modelId"`
	}
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.ErrorResponse(c, 400, "invalid request body: "+err.Error())
			return
		}
	}

	modelID := strings.TrimSpace(req.ModelID)
	var (
		result *service.AnalysisWithStatus
		err    error
	)
	if modelID != "" {
		if modelLLM, ok := h.llmService.(service.LLMServiceWithModelInterface); ok {
			result, err = modelLLM.AnalyzeJobWithModel(jobID, modelID)
		} else {
			utils.ErrorResponse(c, 501, "LLM service does not support custom model selection")
			return
		}
	} else {
		result, err = h.llmService.AnalyzeJob(jobID)
	}

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
				if err := h.llmService.AnalyzeJobSync(id); err != nil {
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

// ExportAnalysesCSV 导出 AI 分析概览 CSV。
// scope:
// - filtered: 导出当前筛选条件下的全部主作业
// - page: 导出当前页主作业
// - selected: 导出 jobIds 指定的主作业
func (h *JobHandler) ExportAnalysesCSV(c *gin.Context) {
	if h.llmService == nil {
		utils.ErrorResponse(c, 501, "LLM service is not configured")
		return
	}

	scope := c.DefaultQuery("scope", "filtered")
	nodeID := c.Query("nodeId")
	statuses := c.QueryArray("status")
	jobTypes := c.QueryArray("type")
	frameworks := c.QueryArray("framework")
	cardCountStrs := c.QueryArray("cardCount")
	var cardCounts []int
	for _, s := range cardCountStrs {
		if s == "unknown" {
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

	selectedIDs := dedupeStrings(c.QueryArray("jobIds"))
	if scope == "selected" && len(selectedIDs) == 0 {
		utils.ErrorResponse(c, 400, "jobIds is required when scope=selected")
		return
	}

	queryPage := page
	queryPageSize := pageSize
	if scope == "filtered" || scope == "selected" {
		queryPage = 1
		queryPageSize = 100000
	}

	groups, _, err := h.jobService.GetGroupedJobs(nodeID, statuses, jobTypes, frameworks, cardCounts, sortBy, sortOrder, queryPage, queryPageSize)
	if err != nil {
		utils.ErrorResponse(c, 500, "Database error: "+err.Error())
		return
	}

	if scope == "selected" {
		idSet := make(map[string]struct{}, len(selectedIDs))
		for _, id := range selectedIDs {
			idSet[id] = struct{}{}
		}
		filtered := make([]service.JobGroup, 0, len(selectedIDs))
		for _, g := range groups {
			if _, ok := idSet[g.MainJob.JobID]; ok {
				filtered = append(filtered, g)
			}
		}
		groups = filtered
	}

	jobIDs := make([]string, 0, len(groups))
	for _, g := range groups {
		jobIDs = append(jobIDs, g.MainJob.JobID)
	}
	analyses, err := h.llmService.GetBatchAnalyses(jobIDs)
	if err != nil {
		utils.ErrorResponse(c, 500, "failed to fetch analyses: "+err.Error())
		return
	}

	filename := fmt.Sprintf("ai-analysis-overview_%s.csv", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Cache-Control", "no-store")

	_, _ = c.Writer.Write([]byte("\xEF\xBB\xBF"))
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	_ = writer.Write([]string{
		"jobId", "jobName", "nodeId", "status", "jobType", "framework", "processName", "commandLine", "startupScript", "startTime", "cardCount",
		"processMemoryMb", "hbmUsageMb", "hbmTotalMb", "hbmUsagePercent", "aicoreUsagePercent", "hardwareOccupancy",
		"summary", "taskType", "modelName", "runtimeStatus", "npuUtilization", "hbmUtilization", "issuesCount",
	})

	for _, group := range groups {
		job := group.MainJob
		analysis := analyses[job.JobID]

		startupScript := "-"
		if codes, codeErr := h.jobService.GetJobCode(job.JobID); codeErr == nil {
			startupScript = extractStartupScript(codes)
		}

		hwStats := exportHardwareStats{}
		if detail, detailErr := h.jobService.GetJobDetail(job.JobID, true); detailErr == nil && detail != nil {
			hwStats = buildExportHardwareStats(detail.NPUCards)
		}

		cardCount := "unknown"
		if group.CardCount != nil {
			cardCount = strconv.Itoa(*group.CardCount)
		}

		summary := ""
		taskType := ""
		modelName := ""
		runtimeStatus := ""
		npuUtil := ""
		hbmUtil := ""
		issuesCount := "0"
		if analysis != nil {
			summary = analysis.Summary
			taskType = analysis.TaskType.Category
			if analysis.ModelInfo != nil && analysis.ModelInfo.ModelName != nil {
				modelName = *analysis.ModelInfo.ModelName
			}
			if analysis.RuntimeAnalysis != nil {
				runtimeStatus = analysis.RuntimeAnalysis.Status
			}
			npuUtil = analysis.ResourceAssessment.NpuUtilization
			hbmUtil = analysis.ResourceAssessment.HbmUtilization
			issuesCount = strconv.Itoa(len(analysis.Issues))
		}

		_ = writer.Write([]string{
			sanitizeCSVCell(job.JobID),
			sanitizeCSVCell(valueOrDash(job.JobName)),
			sanitizeCSVCell(valueOrDash(job.NodeID)),
			sanitizeCSVCell(valueOrDash(job.Status)),
			sanitizeCSVCell(valueOrDash(job.JobType)),
			sanitizeCSVCell(valueOrDash(job.Framework)),
			sanitizeCSVCell(valueOrDash(job.ProcessName)),
			sanitizeCSVCell(valueOrDash(job.CommandLine)),
			sanitizeCSVCell(startupScript),
			sanitizeCSVCell(formatTimeMs(job.StartTime)),
			sanitizeCSVCell(cardCount),
			sanitizeCSVCell(hwStats.ProcessMemoryMB),
			sanitizeCSVCell(hwStats.HBMUsageMB),
			sanitizeCSVCell(hwStats.HBMTotalMB),
			sanitizeCSVCell(hwStats.HBMUsagePercent),
			sanitizeCSVCell(hwStats.AICoreUsagePercent),
			sanitizeCSVCell(hwStats.HardwareOccupancy),
			sanitizeCSVCell(summary),
			sanitizeCSVCell(taskType),
			sanitizeCSVCell(modelName),
			sanitizeCSVCell(runtimeStatus),
			sanitizeCSVCell(npuUtil),
			sanitizeCSVCell(hbmUtil),
			sanitizeCSVCell(issuesCount),
		})
	}
}

type exportHardwareStats struct {
	ProcessMemoryMB    string
	HBMUsageMB         string
	HBMTotalMB         string
	HBMUsagePercent    string
	AICoreUsagePercent string
	HardwareOccupancy  string
}

func buildExportHardwareStats(cards []service.NPUCardInfo) exportHardwareStats {
	if len(cards) == 0 {
		return exportHardwareStats{
			ProcessMemoryMB:    "-",
			HBMUsageMB:         "-",
			HBMTotalMB:         "-",
			HBMUsagePercent:    "-",
			AICoreUsagePercent: "-",
			HardwareOccupancy:  "-",
		}
	}

	var (
		processMemoryMB float64
		hbmUsageMB      float64
		hbmTotalMB      float64
		aicoreSum       float64
		aicoreCount     int
		chipCount       int
	)

	for _, card := range cards {
		processMemoryMB += card.MemoryUsageMB
		chipCount += len(card.Metrics)
		for _, metric := range card.Metrics {
			if metric.HBMUsageMB != nil {
				hbmUsageMB += *metric.HBMUsageMB
			}
			if metric.HBMTotalMB != nil {
				hbmTotalMB += *metric.HBMTotalMB
			}
			if metric.AICoreUsagePercent != nil {
				aicoreSum += *metric.AICoreUsagePercent
				aicoreCount++
			}
		}
	}

	hbmUsagePercent := "-"
	if hbmTotalMB > 0 {
		hbmUsagePercent = fmt.Sprintf("%.2f", hbmUsageMB*100/hbmTotalMB)
	}

	aicoreUsage := "-"
	if aicoreCount > 0 {
		aicoreUsage = fmt.Sprintf("%.2f", aicoreSum/float64(aicoreCount))
	}

	occupancy := fmt.Sprintf("cards=%d,chips=%d", len(cards), chipCount)
	if aicoreUsage != "-" {
		occupancy = fmt.Sprintf("%s,aicore=%s%%", occupancy, aicoreUsage)
	}

	return exportHardwareStats{
		ProcessMemoryMB:    fmt.Sprintf("%.2f", processMemoryMB),
		HBMUsageMB:         fmt.Sprintf("%.2f", hbmUsageMB),
		HBMTotalMB:         fmt.Sprintf("%.2f", hbmTotalMB),
		HBMUsagePercent:    hbmUsagePercent,
		AICoreUsagePercent: aicoreUsage,
		HardwareOccupancy:  occupancy,
	}
}

func extractStartupScript(codes []model.Code) string {
	if len(codes) == 0 {
		return "-"
	}
	code := codes[0]
	paths := make([]string, 0, 2)
	if code.ScriptPath != nil && strings.TrimSpace(*code.ScriptPath) != "" {
		paths = append(paths, strings.TrimSpace(*code.ScriptPath))
	}
	if code.ShScriptPath != nil && strings.TrimSpace(*code.ShScriptPath) != "" {
		paths = append(paths, strings.TrimSpace(*code.ShScriptPath))
	}
	if len(paths) == 0 {
		return "-"
	}
	return strings.Join(paths, " | ")
}

func dedupeStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func valueOrDash[T ~string](v *T) string {
	if v == nil || *v == "" {
		return "-"
	}
	return string(*v)
}

func formatTimeMs(v *int64) string {
	if v == nil || *v <= 0 {
		return "-"
	}
	sec := *v / 1000
	nsec := (*v % 1000) * int64(time.Millisecond)
	return time.Unix(sec, nsec).Format("2006-01-02 15:04:05")
}

func sanitizeCSVCell(v string) string {
	if v == "" {
		return ""
	}
	if strings.HasPrefix(v, "=") || strings.HasPrefix(v, "+") || strings.HasPrefix(v, "-") || strings.HasPrefix(v, "@") {
		return "'" + v
	}
	return v
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
