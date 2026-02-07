package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/task-monitor/api-server/internal/config"
)

// LLMService LLM分析服务
type LLMService struct {
	jobService JobServiceInterface
	httpClient *http.Client
	config     config.LLMConfig
	mu         sync.RWMutex
}

// NewLLMService 创建LLM服务
func NewLLMService(jobService JobServiceInterface, cfg config.LLMConfig) *LLMService {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 60
	}
	return &LLMService{
		jobService: jobService,
		httpClient: &http.Client{Timeout: time.Duration(timeout) * time.Second},
		config:     cfg,
	}
}

// GetConfig 获取当前LLM配置（API Key脱敏）
func (s *LLMService) GetConfig() config.LLMConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg := s.config
	if len(cfg.APIKey) > 4 {
		cfg.APIKey = "****" + cfg.APIKey[len(cfg.APIKey)-4:]
	} else if cfg.APIKey != "" {
		cfg.APIKey = "****"
	}
	return cfg
}

// UpdateConfig 更新LLM配置
func (s *LLMService) UpdateConfig(cfg config.LLMConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.config = cfg
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 60
	}
	s.httpClient = &http.Client{Timeout: time.Duration(timeout) * time.Second}
}

// chatMessage OpenAI chat message
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest OpenAI chat completions request
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

// chatResponse OpenAI chat completions response
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// AnalyzeJob 分析作业
func (s *LLMService) AnalyzeJob(jobID string) (*JobAnalysisResponse, error) {
	s.mu.RLock()
	enabled := s.config.Enabled
	s.mu.RUnlock()
	if !enabled {
		return nil, fmt.Errorf("LLM service is not enabled")
	}

	// 1. 聚合作业数据
	userPrompt, err := s.buildUserPrompt(jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// 2. 调用LLM
	content, err := s.callLLM(systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to call LLM: %w", err)
	}

	// 3. 解析返回的JSON
	result, err := s.parseResponse(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return result, nil
}

// buildUserPrompt 聚合作业数据构建用户提示词
func (s *LLMService) buildUserPrompt(jobID string) (string, error) {
	var sb strings.Builder

	// 获取作业详情
	detail, err := s.jobService.GetJobDetail(jobID)
	if err != nil {
		return "", fmt.Errorf("get job detail: %w", err)
	}
	job := detail.Job

	sb.WriteString("## 作业基本信息\n")
	sb.WriteString(fmt.Sprintf("- 作业ID: %s\n", job.JobID))
	if job.JobName != nil {
		sb.WriteString(fmt.Sprintf("- 作业名称: %s\n", *job.JobName))
	}
	if job.JobType != nil {
		sb.WriteString(fmt.Sprintf("- 作业类型: %s\n", *job.JobType))
	}
	if job.Framework != nil {
		sb.WriteString(fmt.Sprintf("- 框架: %s\n", *job.Framework))
	}
	if job.Status != nil {
		sb.WriteString(fmt.Sprintf("- 状态: %s\n", *job.Status))
	}
	if job.ProcessName != nil {
		sb.WriteString(fmt.Sprintf("- 进程名称: %s\n", *job.ProcessName))
	}
	if job.CommandLine != nil {
		sb.WriteString(fmt.Sprintf("- 命令行: %s\n", *job.CommandLine))
	}
	if job.CWD != nil {
		sb.WriteString(fmt.Sprintf("- 工作目录: %s\n", *job.CWD))
	}

	// NPU卡信息
	sb.WriteString(fmt.Sprintf("\n## NPU 卡信息 (共 %d 张)\n", len(detail.NPUCards)))
	for _, card := range detail.NPUCards {
		sb.WriteString(fmt.Sprintf("- NPU %d: 进程显存 %.1f MB", card.NpuID, card.MemoryUsageMB))
		if card.Metric != nil {
			m := card.Metric
			if m.AICoreUsagePercent != nil {
				sb.WriteString(fmt.Sprintf(", AICore使用率 %.1f%%", *m.AICoreUsagePercent))
			}
			if m.HBMUsageMB != nil && m.HBMTotalMB != nil {
				sb.WriteString(fmt.Sprintf(", HBM %.0f/%.0f MB", *m.HBMUsageMB, *m.HBMTotalMB))
			}
			if m.PowerW != nil {
				sb.WriteString(fmt.Sprintf(", 功率 %.1fW", *m.PowerW))
			}
			if m.TempC != nil {
				sb.WriteString(fmt.Sprintf(", 温度 %.1f°C", *m.TempC))
			}
		}
		sb.WriteString("\n")
	}

	// 关联进程
	if len(detail.RelatedJobs) > 0 {
		sb.WriteString(fmt.Sprintf("\n## 关联进程 (共 %d 个)\n", len(detail.RelatedJobs)))
		for _, rj := range detail.RelatedJobs {
			name := "-"
			if rj.ProcessName != nil {
				name = *rj.ProcessName
			}
			sb.WriteString(fmt.Sprintf("- PID %v, 进程名: %s\n", safeInt64(rj.PID), name))
		}
	}

	// 获取参数和环境变量
	params, err := s.jobService.GetJobParameters(jobID)
	if err == nil && len(params) > 0 {
		p := params[0]
		if p.ParameterData != nil && *p.ParameterData != "" {
			sb.WriteString("\n## 参数配置\n")
			sb.WriteString("```json\n")
			sb.WriteString(truncateStr(*p.ParameterData, 3000))
			sb.WriteString("\n```\n")
		}
		if p.ConfigFileContent != nil && *p.ConfigFileContent != "" {
			sb.WriteString("\n## 配置文件内容\n")
			if p.ConfigFilePath != nil {
				sb.WriteString(fmt.Sprintf("路径: %s\n", *p.ConfigFilePath))
			}
			sb.WriteString("```\n")
			sb.WriteString(truncateStr(*p.ConfigFileContent, 3000))
			sb.WriteString("\n```\n")
		}
		if p.EnvVars != nil && *p.EnvVars != "" {
			sb.WriteString("\n## 环境变量（已过滤敏感信息）\n")
			sb.WriteString(filterSensitiveEnvVars(*p.EnvVars))
			sb.WriteString("\n")
		}
	}

	// 获取脚本代码
	codes, err := s.jobService.GetJobCode(jobID)
	if err == nil && len(codes) > 0 {
		c := codes[0]
		if c.ScriptContent != nil && *c.ScriptContent != "" {
			sb.WriteString("\n## 启动脚本代码\n")
			if c.ScriptPath != nil {
				sb.WriteString(fmt.Sprintf("路径: %s\n", *c.ScriptPath))
			}
			sb.WriteString("```python\n")
			sb.WriteString(truncateStr(*c.ScriptContent, 5000))
			sb.WriteString("\n```\n")
		}
		if c.ShScriptContent != nil && *c.ShScriptContent != "" {
			sb.WriteString("\n## Shell 启动脚本\n")
			if c.ShScriptPath != nil {
				sb.WriteString(fmt.Sprintf("路径: %s\n", *c.ShScriptPath))
			}
			sb.WriteString("```bash\n")
			sb.WriteString(truncateStr(*c.ShScriptContent, 3000))
			sb.WriteString("\n```\n")
		}
	}

	return sb.String(), nil
}

// callLLM 调用OpenAI兼容接口
func (s *LLMService) callLLM(sysPrompt, userPrompt string) (string, error) {
	s.mu.RLock()
	cfg := s.config
	s.mu.RUnlock()

	reqBody := chatRequest{
		Model: cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: sysPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.3,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	endpoint := strings.TrimRight(cfg.Endpoint, "/") + "/chat/completions"
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLM API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned empty choices")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// parseResponse 解析LLM返回的JSON
func (s *LLMService) parseResponse(content string) (*JobAnalysisResponse, error) {
	// 尝试提取JSON块（LLM可能返回markdown包裹的JSON）
	jsonStr := extractJSON(content)

	var result JobAnalysisResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("parse JSON: %w (raw: %s)", err, truncateStr(content, 500))
	}
	return &result, nil
}

// extractJSON 从可能包含markdown代码块的文本中提取JSON
func extractJSON(s string) string {
	// 尝试找 ```json ... ``` 块
	if idx := strings.Index(s, "```json"); idx != -1 {
		start := idx + 7
		if end := strings.Index(s[start:], "```"); end != -1 {
			return strings.TrimSpace(s[start : start+end])
		}
	}
	// 尝试找 ``` ... ``` 块
	if idx := strings.Index(s, "```"); idx != -1 {
		start := idx + 3
		// 跳过可能的语言标识行
		if nlIdx := strings.Index(s[start:], "\n"); nlIdx != -1 {
			start = start + nlIdx + 1
		}
		if end := strings.Index(s[start:], "```"); end != -1 {
			return strings.TrimSpace(s[start : start+end])
		}
	}
	// 尝试找第一个 { 到最后一个 }
	first := strings.Index(s, "{")
	last := strings.LastIndex(s, "}")
	if first != -1 && last > first {
		return s[first : last+1]
	}
	return s
}

// truncateStr 截断字符串
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (内容已截断)"
}

// filterSensitiveEnvVars 过滤敏感环境变量
func filterSensitiveEnvVars(envJSON string) string {
	var envMap map[string]string
	if err := json.Unmarshal([]byte(envJSON), &envMap); err != nil {
		return "(解析失败)"
	}

	sensitiveKeys := []string{"PASSWORD", "SECRET", "TOKEN", "KEY", "CREDENTIAL", "AUTH"}
	var sb strings.Builder
	for k, v := range envMap {
		upper := strings.ToUpper(k)
		isSensitive := false
		for _, sk := range sensitiveKeys {
			if strings.Contains(upper, sk) {
				isSensitive = true
				break
			}
		}
		if isSensitive {
			sb.WriteString(fmt.Sprintf("- %s=***\n", k))
		} else {
			sb.WriteString(fmt.Sprintf("- %s=%s\n", k, v))
		}
	}
	return sb.String()
}

// safeInt64 安全获取int64指针值
func safeInt64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

// systemPrompt LLM系统提示词
const systemPrompt = `你是一个专业的 NPU（华为昇腾）作业分析助手。用户会提供一个在昇腾 NPU 上运行的作业的详细信息，包括基本信息、NPU 资源使用、脚本代码、参数配置和环境变量。

请你综合分析这些信息，并严格按照以下 JSON 格式返回结果，不要输出任何其他内容：

{
  "summary": "100字以内的作业概要描述",
  "taskType": {
    "category": "training 或 inference 或 unknown",
    "subCategory": "pre-training / fine-tuning / rlhf / evaluation / serving / batch-inference 或 null",
    "inferenceFramework": "推理框架名称如 vLLM / TGI / MindIE 或 null"
  },
  "modelInfo": {
    "modelName": "模型名称或null",
    "modelSize": "模型大小如7B/13B/70B或null",
    "precision": "精度如fp16/bf16/int8或null",
    "parallelStrategy": "并行策略如TP=8/PP=2或null"
  },
  "resourceAssessment": {
    "npuUtilization": "high / medium / low / idle",
    "hbmUtilization": "high / medium / low",
    "description": "资源使用情况的简要描述"
  },
  "issues": [
    {
      "severity": "critical / warning / info",
      "category": "问题分类",
      "description": "问题描述",
      "suggestion": "改进建议"
    }
  ],
  "suggestions": ["优化建议1", "优化建议2"]
}

分析要点：
1. 根据命令行、脚本内容、框架信息判断作业类型（训练/推理）和子类型
2. 从参数、环境变量、脚本中提取模型名称、大小、精度、并行策略
3. 根据 NPU AICore 使用率和 HBM 使用情况评估资源利用率
4. 识别潜在问题：资源浪费、配置不当、性能瓶颈等
5. 给出针对性的优化建议

如果某些信息无法确定，对应字段填 null。modelInfo 整体可以为 null。
issues 和 suggestions 数组可以为空但不能为 null。`
