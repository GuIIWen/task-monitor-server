package service

import (
	"github.com/task-monitor/api-server/internal/config"
	"github.com/task-monitor/api-server/internal/model"
)

// NodeServiceInterface defines the interface for node service operations
type NodeServiceInterface interface {
	GetNodes() ([]model.Node, error)
	GetNodeByID(nodeID string) (*model.Node, error)
	GetNodesByStatus(status string) ([]model.Node, error)
	GetNodeStats() (map[string]int64, error)
}

// JobGroup 作业分组（按 node_id + pgid 分组）
type JobGroup struct {
	MainJob   model.Job   `json:"mainJob"`
	ChildJobs []model.Job `json:"childJobs"`
	CardCount *int        `json:"cardCount"` // nil 表示 unknown
}

// NPUCardInfo 进程占用的 NPU 卡信息
type NPUCardInfo struct {
	NpuID         int              `json:"npuId"`
	MemoryUsageMB float64          `json:"memoryUsageMb"`
	Metric        *model.NPUMetric `json:"metric"`
}

// JobDetailResponse 作业详情响应
type JobDetailResponse struct {
	Job         model.Job   `json:"job"`
	NPUCards    []NPUCardInfo `json:"npuCards"`
	RelatedJobs []model.Job   `json:"relatedJobs"`
}

// JobAnalysisTaskType 作业类型分析
type JobAnalysisTaskType struct {
	Category           string  `json:"category"`
	SubCategory        *string `json:"subCategory"`
	InferenceFramework *string `json:"inferenceFramework"`
	Evidence           *string `json:"evidence"`
}

// JobAnalysisModelInfo 模型信息分析
type JobAnalysisModelInfo struct {
	ModelName        *string `json:"modelName"`
	ModelSize        *string `json:"modelSize"`
	Precision        *string `json:"precision"`
	ParallelStrategy *string `json:"parallelStrategy"`
}

// JobAnalysisRuntimeAnalysis 运行时长分析
type JobAnalysisRuntimeAnalysis struct {
	Duration    string `json:"duration"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

// JobAnalysisParameterItem 参数检查项
type JobAnalysisParameterItem struct {
	Parameter  string `json:"parameter"`
	Value      string `json:"value"`
	Assessment string `json:"assessment"`
	Reason     string `json:"reason"`
}

// JobAnalysisParameterCheck 参数合理性检查
type JobAnalysisParameterCheck struct {
	Status string                     `json:"status"`
	Items  []JobAnalysisParameterItem `json:"items"`
}

// JobAnalysisResourceAssessment 资源评估
type JobAnalysisResourceAssessment struct {
	NpuUtilization string `json:"npuUtilization"`
	HbmUtilization string `json:"hbmUtilization"`
	Description    string `json:"description"`
}

// JobAnalysisIssue 问题项
type JobAnalysisIssue struct {
	Severity    string `json:"severity"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
}

// JobAnalysisResponse LLM分析结果
type JobAnalysisResponse struct {
	Summary            string                        `json:"summary"`
	TaskType           JobAnalysisTaskType            `json:"taskType"`
	ModelInfo          *JobAnalysisModelInfo           `json:"modelInfo"`
	RuntimeAnalysis    *JobAnalysisRuntimeAnalysis     `json:"runtimeAnalysis"`
	ParameterCheck     *JobAnalysisParameterCheck      `json:"parameterCheck"`
	ResourceAssessment JobAnalysisResourceAssessment   `json:"resourceAssessment"`
	Issues             []JobAnalysisIssue             `json:"issues"`
	Suggestions        []string                       `json:"suggestions"`
}

// LLMServiceInterface LLM服务接口
type LLMServiceInterface interface {
	AnalyzeJob(jobID string) (*JobAnalysisResponse, error)
	GetConfig() config.LLMConfig
	UpdateConfig(cfg config.LLMConfig)
}

// JobServiceInterface defines the interface for job service operations
type JobServiceInterface interface {
	GetJobByID(jobID string) (*model.Job, error)
	GetJobDetail(jobID string) (*JobDetailResponse, error)
	GetJobsByNodeID(nodeID string) ([]model.Job, error)
	GetJobsByStatus(status string) ([]model.Job, error)
	GetAllJobs() ([]model.Job, error)
	GetJobs(nodeID string, statuses []string, jobTypes []string, frameworks []string, sortBy, sortOrder string, page, pageSize int) ([]model.Job, int64, error)
	GetGroupedJobs(nodeID string, statuses []string, jobTypes []string, frameworks []string, cardCounts []int, sortBy, sortOrder string, page, pageSize int) ([]JobGroup, int64, error)
	GetDistinctCardCounts() ([]int, error)
	GetJobParameters(jobID string) ([]model.Parameter, error)
	GetJobCode(jobID string) ([]model.Code, error)
	GetJobStats() (map[string]int64, error)
}

// AuthServiceInterface 认证服务接口
type AuthServiceInterface interface {
	Login(username, password string) (string, error)
	ParseToken(tokenString string) (uint, string, error)
	GetUserByID(id uint) (*model.User, error)
	ListUsers() ([]model.User, error)
	CreateUser(username, password string) (*model.User, error)
	ChangePassword(userID uint, newPassword string) error
	DeleteUser(userID uint, currentUserID uint) error
}
