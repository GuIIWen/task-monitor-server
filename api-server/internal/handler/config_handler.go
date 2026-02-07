package handler

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/task-monitor/api-server/internal/config"
	"github.com/task-monitor/api-server/internal/service"
	"github.com/task-monitor/api-server/internal/utils"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	llmService service.LLMServiceInterface
	config     *config.Config
	configPath string
	mu         sync.Mutex
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(llmService service.LLMServiceInterface, cfg *config.Config, configPath string) *ConfigHandler {
	return &ConfigHandler{
		llmService: llmService,
		config:     cfg,
		configPath: configPath,
	}
}

// GetLLMConfig 获取LLM配置（API Key脱敏）
func (h *ConfigHandler) GetLLMConfig(c *gin.Context) {
	cfg := h.llmService.GetConfig()
	utils.SuccessResponse(c, cfg)
}

// UpdateLLMConfigRequest 更新LLM配置请求
type UpdateLLMConfigRequest struct {
	Enabled  *bool   `json:"enabled"`
	Endpoint *string `json:"endpoint"`
	APIKey   *string `json:"api_key"`
	Model    *string `json:"model"`
	Timeout  *int    `json:"timeout"`
}

// UpdateLLMConfig 更新LLM配置
func (h *ConfigHandler) UpdateLLMConfig(c *gin.Context) {
	var req UpdateLLMConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// 合并更新：只更新请求中提供的字段
	llmCfg := h.config.LLM
	if req.Enabled != nil {
		llmCfg.Enabled = *req.Enabled
	}
	if req.Endpoint != nil {
		llmCfg.Endpoint = *req.Endpoint
	}
	if req.APIKey != nil {
		llmCfg.APIKey = *req.APIKey
	}
	if req.Model != nil {
		llmCfg.Model = *req.Model
	}
	if req.Timeout != nil {
		llmCfg.Timeout = *req.Timeout
	}

	// 更新内存中的配置
	h.config.LLM = llmCfg
	h.llmService.UpdateConfig(llmCfg)

	// 持久化到文件
	if err := config.SaveConfig(h.configPath, h.config); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "config updated in memory but failed to save to file: "+err.Error())
		return
	}

	utils.SuccessResponse(c, h.llmService.GetConfig())
}
