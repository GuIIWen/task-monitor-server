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
	authService := service.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.ExpireHour)

	// 初始化LLM Service（始终创建，可通过页面启用/禁用）
	jobAnalysisRepo := repository.NewJobAnalysisRepository(db)
	llmService := service.NewLLMService(jobService, jobAnalysisRepo, cfg.LLM)
	if cfg.LLM.Enabled {
		log.Println("LLM service enabled")
	}

	// 初始化Handler
	nodeHandler := handler.NewNodeHandler(nodeService)
	jobHandler := handler.NewJobHandler(jobService, llmService)
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
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
		}

		// 需要认证的路由
		api.Use(middleware.JWTAuth(authService))

		// 当前用户
		api.GET("/auth/me", authHandler.GetCurrentUser)

		// 用户管理
		users := api.Group("/users")
		{
			users.GET("", authHandler.ListUsers)
			users.POST("", authHandler.CreateUser)
			users.PUT("/:id/password", authHandler.ChangePassword)
			users.DELETE("/:id", authHandler.DeleteUser)
		}

		// 节点相关路由
		nodes := api.Group("/nodes")
		{
			nodes.GET("", nodeHandler.GetNodes)
			nodes.GET("/stats", nodeHandler.GetNodeStats)
			nodes.GET("/:nodeId", nodeHandler.GetNodeByID)
		}

		// 作业相关路由
		jobs := api.Group("/jobs")
		{
			jobs.GET("", jobHandler.GetJobs)
			jobs.GET("/grouped", jobHandler.GetGroupedJobs)
			jobs.GET("/grouped/card-counts", jobHandler.GetDistinctCardCounts)
			jobs.GET("/stats", jobHandler.GetJobStats)
			jobs.POST("/batch-analyze", jobHandler.BatchAnalyze)
			jobs.GET("/batch-analyze/:batchId", jobHandler.GetBatchAnalyzeProgress)
			jobs.GET("/:jobId", jobHandler.GetJobByID)
			jobs.GET("/:jobId/parameters", jobHandler.GetJobParameters)
			jobs.GET("/:jobId/code", jobHandler.GetJobCode)
			jobs.POST("/:jobId/analyze", jobHandler.AnalyzeJob)
			jobs.GET("/:jobId/analysis", jobHandler.GetJobAnalysis)
		}

		// 配置相关路由
		cfgGroup := api.Group("/config")
		{
			cfgGroup.GET("/llm", configHandler.GetLLMConfig)
			cfgGroup.PUT("/llm", configHandler.UpdateLLMConfig)
		}
	}

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("API Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
