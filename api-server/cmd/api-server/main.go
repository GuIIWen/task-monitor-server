package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/task-monitor/api-server/internal/config"
	"github.com/task-monitor/api-server/internal/handler"
	"github.com/task-monitor/api-server/internal/repository"
	"github.com/task-monitor/api-server/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("configs/api-server.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库
	db, err := config.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 初始化Repository
	nodeRepo := repository.NewNodeRepository(db)
	jobRepo := repository.NewJobRepository(db)
	paramRepo := repository.NewParameterRepository(db)
	codeRepo := repository.NewCodeRepository(db)
	metricsRepo := repository.NewMetricsRepository(db)

	// 初始化Service
	nodeService := service.NewNodeService(nodeRepo)
	jobService := service.NewJobService(jobRepo, paramRepo, codeRepo, metricsRepo)

	// 初始化Handler
	nodeHandler := handler.NewNodeHandler(nodeService)
	jobHandler := handler.NewJobHandler(jobService)

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
		// 节点相关路由
		nodes := api.Group("/nodes")
		{
			nodes.GET("", nodeHandler.GetNodes)
			nodes.GET("/:nodeId", nodeHandler.GetNodeByID)
		}

		// 作业相关路由
		jobs := api.Group("/jobs")
		{
			jobs.GET("", jobHandler.GetJobs)
			jobs.GET("/:jobId", jobHandler.GetJobByID)
			jobs.GET("/:jobId/parameters", jobHandler.GetJobParameters)
			jobs.GET("/:jobId/code", jobHandler.GetJobCode)
		}
	}

	// 启动服务器
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("API Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
