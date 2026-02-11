package handler

import (
	"errors"
	"net/http"
	"strings"
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
	Enabled          *bool                    `json:"enabled"`
	Endpoint         *string                  `json:"endpoint"`
	APIKey           *string                  `json:"api_key"`
	Model            *string                  `json:"model"`
	Timeout          *int                     `json:"timeout"`
	BatchConcurrency *int                     `json:"batch_concurrency"`
	DefaultModelID   *string                  `json:"default_model_id"`
	Models           *[]config.LLMModelConfig `json:"models"`
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
	if req.APIKey != nil && *req.APIKey != "" && !strings.HasPrefix(*req.APIKey, "****") {
		llmCfg.APIKey = *req.APIKey
	}
	if req.Model != nil {
		llmCfg.Model = *req.Model
	}
	if req.Timeout != nil {
		llmCfg.Timeout = *req.Timeout
	}
	if req.BatchConcurrency != nil {
		llmCfg.BatchConcurrency = *req.BatchConcurrency
	}
	if req.Models != nil {
		mergedModels, err := mergeModelConfig(*req.Models, llmCfg.Models)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
			return
		}
		llmCfg.Models = mergedModels
	}
	if req.DefaultModelID != nil {
		llmCfg.DefaultModelID = strings.TrimSpace(*req.DefaultModelID)
	}

	if err := validateDefaultModelID(llmCfg); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
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

func mergeModelConfig(newModels, oldModels []config.LLMModelConfig) ([]config.LLMModelConfig, error) {
	oldByID := make(map[string]config.LLMModelConfig, len(oldModels))
	for _, m := range oldModels {
		oldByID[m.ID] = m
	}

	seen := make(map[string]struct{}, len(newModels))
	merged := make([]config.LLMModelConfig, 0, len(newModels))
	for _, m := range newModels {
		m.ID = strings.TrimSpace(m.ID)
		m.Name = strings.TrimSpace(m.Name)
		m.Endpoint = strings.TrimSpace(m.Endpoint)
		m.Model = strings.TrimSpace(m.Model)

		if m.ID == "" {
			return nil, errors.New("model id is required")
		}
		if _, ok := seen[m.ID]; ok {
			return nil, errors.New("duplicate model id: " + m.ID)
		}
		seen[m.ID] = struct{}{}

		if strings.HasPrefix(m.APIKey, "****") {
			if old, ok := oldByID[m.ID]; ok {
				m.APIKey = old.APIKey
			}
		}
		merged = append(merged, m)
	}

	return merged, nil
}

func validateDefaultModelID(cfg config.LLMConfig) error {
	if strings.TrimSpace(cfg.DefaultModelID) == "" || len(cfg.Models) == 0 {
		return nil
	}

	for _, m := range cfg.Models {
		if m.ID != cfg.DefaultModelID {
			continue
		}
		if !m.Enabled {
			return errors.New("default model must be enabled")
		}
		return nil
	}

	return errors.New("default model not found")
}
