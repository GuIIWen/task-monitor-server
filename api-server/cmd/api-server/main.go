package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/task-monitor/api-server/internal/config"
	"github.com/task-monitor/api-server/internal/handler"
	"github.com/task-monitor/api-server/internal/middleware"
	"github.com/task-monitor/api-server/internal/repository"
	"github.com/task-monitor/api-server/internal/service"
)

func main() {
	// 支持命令行参数和环境变量指定配置文件路径
	configPath := flag.String("config", "", "配置文件路径")
	flag.Parse()

	// 优先级：命令行参数 > 环境变量 > 默认路径
	if *configPath == "" {
		*configPath = os.Getenv("API_SERVER_CONFIG")
	}
	if *configPath == "" {
		*configPath = "configs/api-server.yaml"
	}

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", *configPath, err)
	}

	// 初始化数据库
	db, err := config.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 自动建表并创建默认用户
	if err := config.AutoMigrateAndSeed(db); err != nil {
		log.Fatalf("Failed to migrate and seed: %v", err)
	}

	// 初始化Repository
	nodeRepo := repository.NewNodeRepository(db)
	jobRepo := repository.NewJobRepository(db)
	paramRepo := repository.NewParameterRepository(db)
	codeRepo := repository.NewCodeRepository(db)
	metricsRepo := repository.NewMetricsRepository(db)
	userRepo := repository.NewUserRepository(db)

	// 初始化Service
	nodeService := service.NewNodeService(nodeRepo)
	jobService := service.NewJobService(jobRepo, paramRepo, codeRepo, metricsRepo)
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireMinutes)

	// 初始化LLM Service（始终创建，可通过页面启用/禁用）
	jobAnalysisRepo := repository.NewJobAnalysisRepository(db)
	llmService := service.NewLLMService(jobService, jobAnalysisRepo, cfg.LLM)
	if cfg.LLM.Enabled {
		log.Println("LLM service enabled")
	}

	// 初始化Handler
	nodeHandler := handler.NewNodeHandler(nodeService)
	jobHandler := handler.NewJobHandler(jobService, llmService, cfg.LLM.BatchConcurrency)
	configHandler := handler.NewConfigHandler(llmService, cfg, *configPath)
	authHandler := handler.NewAuthHandler(authService)

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建路由
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API路由组
	api := r.Group("/api/v1")
	{
		// 公开路由（不需要认证）
		api.POST("/auth/login", authHandler.Login)

		// 节点（只读）
		api.GET("/nodes", nodeHandler.GetNodes)
		api.GET("/nodes/stats", nodeHandler.GetNodeStats)
		api.GET("/nodes/:nodeId", nodeHandler.GetNodeByID)

		// 作业（只读）
		api.GET("/jobs", jobHandler.GetJobs)
		api.GET("/jobs/grouped", jobHandler.GetGroupedJobs)
		api.GET("/jobs/grouped/card-counts", jobHandler.GetDistinctCardCounts)
		api.GET("/jobs/stats", jobHandler.GetJobStats)
		api.GET("/jobs/batch-analyze/:batchId", jobHandler.GetBatchAnalyzeProgress)
		api.GET("/jobs/analyses/batch", jobHandler.GetBatchAnalyses)
		api.GET("/jobs/analyses/export", jobHandler.ExportAnalysesCSV)
		api.GET("/jobs/:jobId", jobHandler.GetJobByID)
		api.GET("/jobs/:jobId/parameters", jobHandler.GetJobParameters)
		api.GET("/jobs/:jobId/code", jobHandler.GetJobCode)
		api.GET("/jobs/:jobId/analysis", jobHandler.GetJobAnalysis)

		// 配置（只读）
		api.GET("/config/llm", configHandler.GetLLMConfig)

		// === 以下路由需要认证 ===
		authed := api.Group("")
		authed.Use(middleware.JWTAuth(authService))

		authed.GET("/auth/me", authHandler.GetCurrentUser)

		// 用户管理
		authed.GET("/users", authHandler.ListUsers)
		authed.POST("/users", authHandler.CreateUser)
		authed.PUT("/users/:id/password", authHandler.ChangePassword)
		authed.DELETE("/users/:id", authHandler.DeleteUser)

		// 作业分析（写操作）
		authed.POST("/jobs/batch-analyze", jobHandler.BatchAnalyze)
		authed.POST("/jobs/batch-analyze/:batchId/cancel", jobHandler.CancelBatchAnalyze)
		authed.POST("/jobs/:jobId/analyze", jobHandler.AnalyzeJob)

		// 配置修改
		authed.PUT("/config/llm", configHandler.UpdateLLMConfig)
	}

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("API Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
