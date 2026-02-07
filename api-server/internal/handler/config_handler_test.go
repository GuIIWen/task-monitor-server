package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/task-monitor/api-server/internal/config"
)

func TestConfigHandler_GetLLMConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLLM := new(MockLLMService)
	cfg := &config.Config{}
	handler := NewConfigHandler(mockLLM, cfg, "/tmp/test-config.yaml")

	mockLLM.On("GetConfig").Return(config.LLMConfig{
		Enabled:  true,
		Endpoint: "http://localhost:8000/v1",
		APIKey:   "****1234",
		Model:    "qwen2.5",
		Timeout:  60,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/config/llm", nil)

	handler.GetLLMConfig(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])

	data := response["data"].(map[string]interface{})
	assert.Equal(t, true, data["enabled"])
	assert.Equal(t, "****1234", data["api_key"])
	assert.Equal(t, "qwen2.5", data["model"])
	mockLLM.AssertExpectations(t)
}

func TestConfigHandler_UpdateLLMConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	os.WriteFile(configPath, []byte("server:\n  port: 8080\n"), 0644)

	mockLLM := new(MockLLMService)
	cfg := &config.Config{
		Server: config.ServerConfig{Port: 8080},
		LLM: config.LLMConfig{
			Enabled:  false,
			Endpoint: "http://old:8000/v1",
			Model:    "old-model",
			Timeout:  30,
		},
	}
	h := NewConfigHandler(mockLLM, cfg, configPath)

	updatedCfg := config.LLMConfig{
		Enabled:  true,
		Endpoint: "http://new:8000/v1",
		APIKey:   "new-key",
		Model:    "new-model",
		Timeout:  120,
	}
	mockLLM.On("UpdateConfig", updatedCfg).Return()
	mockLLM.On("GetConfig").Return(config.LLMConfig{
		Enabled:  true,
		Endpoint: "http://new:8000/v1",
		APIKey:   "****-key",
		Model:    "new-model",
		Timeout:  120,
	})

	body, _ := json.Marshal(map[string]interface{}{
		"enabled":  true,
		"endpoint": "http://new:8000/v1",
		"api_key":  "new-key",
		"model":    "new-model",
		"timeout":  120,
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/v1/config/llm", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateLLMConfig(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])

	// 验证文件已持久化
	savedData, err := os.ReadFile(configPath)
	assert.NoError(t, err)
	assert.Contains(t, string(savedData), "new-model")

	mockLLM.AssertExpectations(t)
}

func TestConfigHandler_UpdateLLMConfig_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockLLM := new(MockLLMService)
	cfg := &config.Config{}
	h := NewConfigHandler(mockLLM, cfg, "/tmp/test.yaml")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/api/v1/config/llm", bytes.NewReader([]byte("invalid")))
	c.Request.Header.Set("Content-Type", "application/json")

	h.UpdateLLMConfig(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
