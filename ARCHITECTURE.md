# NPU作业监控系统 - 分离架构设计

## 文档信息
- **版本**: v1.0.0
- **创建日期**: 2024-02-05
- **架构类型**: Agent-Server分离架构

## 一、架构概览

### 1.1 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                      计算节点集群                             │
│                                                             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Node 1     │  │   Node 2     │  │   Node N     │     │
│  │              │  │              │  │              │     │
│  │  ┌────────┐  │  │  ┌────────┐  │  │  ┌────────┐  │     │
│  │  │ Agent  │  │  │  │ Agent  │  │  │  │ Agent  │  │     │
│  │  └────┬───┘  │  │  └────┬───┘  │  │  └────┬───┘  │     │
│  │       │      │  │       │      │  │       │      │     │
│  │  ┌────▼───┐  │  │  ┌────▼───┐  │  │  ┌────▼───┐  │     │
│  │  │ Jobs   │  │  │  │ Jobs   │  │  │  │ Jobs   │  │     │
│  │  │ NPUs   │  │  │  │ NPUs   │  │  │  │ NPUs   │  │     │
│  │  └────────┘  │  │  └────────┘  │  │  └────────┘  │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
│         │                 │                 │             │
└─────────┼─────────────────┼─────────────────┼─────────────┘
          │                 │                 │
          │ HTTP POST       │                 │
          │ (内网)          │                 │
          ▼                 ▼                 ▼
┌─────────────────────────────────────────────────────────────┐
│                    Agent Server (内网)                       │
│                    Port: 8081                               │
│                                                             │
│  ┌──────────────────────────────────────────────────┐      │
│  │  Agent API Endpoints                             │      │
│  │  - POST /agent/v1/heartbeat    (节点心跳)        │      │
│  │  - POST /agent/v1/jobs         (作业上报)        │      │
│  │  - POST /agent/v1/metrics      (指标上报)        │      │
│  └──────────────────────────────────────────────────┘      │
│                         │                                   │
│                         │ 写入数据                           │
│                         ▼                                   │
│  ┌──────────────────────────────────────────────────┐      │
│  │              MySQL Database                      │      │
│  │  - nodes, jobs, parameters, code                 │      │
│  │  - npu_metrics, process_metrics                  │      │
│  └──────────────────────────────────────────────────┘      │
│                         │                                   │
└─────────────────────────┼───────────────────────────────────┘
                          │ 读取数据
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    API Server (外网)                         │
│                    Port: 8080                               │
│                                                             │
│  ┌──────────────────────────────────────────────────┐      │
│  │  RESTful API Endpoints                           │      │
│  │  - GET  /api/v1/nodes          (节点查询)        │      │
│  │  - GET  /api/v1/jobs           (作业查询)        │      │
│  │  - GET  /api/v1/metrics        (指标查询)        │      │
│  │  - GET  /api/v1/stats          (统计分析)        │      │
│  └──────────────────────────────────────────────────┘      │
│                         ▲                                   │
└─────────────────────────┼───────────────────────────────────┘
                          │ HTTP GET
                          │ (外网/内网)
                          │
                  ┌───────┴────────┐
                  │                │
          ┌───────▼──────┐  ┌──────▼──────┐
          │   Frontend   │  │  Third-party│
          │   (React)    │  │   Clients   │
          └──────────────┘  └─────────────┘
```

### 1.2 架构特点

**职责分离**：
- **Agent Server**: 专注于接收Agent上报的数据，高吞吐写入
- **API Server**: 专注于提供查询服务，高并发读取

**安全隔离**：
- **Agent Server**: 部署在内网，只接受Agent连接
- **API Server**: 可暴露外网，提供前端访问

**独立扩展**：
- 根据Agent数量扩展Agent Server
- 根据查询负载扩展API Server

**数据库共享**：
- 两个服务共享同一个MySQL数据库
- Agent Server负责写入，API Server负责读取
- 读写分离，避免锁竞争

## 二、Agent Server设计

### 2.1 职责定义

**核心职责**：
1. 接收Agent心跳，更新节点状态
2. 接收作业信息上报，写入数据库
3. 接收监控指标上报，写入数据库
4. 实现Job Monitor机制，检测失联作业

**非职责**：
- 不提供查询接口（由API Server负责）
- 不处理前端请求
- 不进行复杂的数据分析

### 2.2 接口设计

**基础路径**: `http://agent-server:8081/agent/v1`

#### 2.2.1 节点心跳

**接口**: `POST /agent/v1/heartbeat`

**请求体**：
```json
{
  "nodeId": "a1b2c3d4e5f6",
  "hostId": "host-001",
  "hostname": "gpu-node-01",
  "ipAddress": "192.168.1.100",
  "npuCount": 8,
  "timestamp": "2024-02-05T10:30:00.000Z"
}
```

**响应**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "nodeId": "a1b2c3d4e5f6",
    "status": "active"
  }
}
```

#### 2.2.2 作业上报

**接口**: `POST /agent/v1/jobs`

**请求体**：
```json
{
  "nodeId": "a1b2c3d4e5f6",
  "jobs": [
    {
      "jobId": "abc123def456",
      "jobName": "train_model.py",
      "jobType": "training",
      "pid": 12345,
      "ppid": 1000,
      "pgid": 12345,
      "processName": "python",
      "commandLine": "python train.py --batch_size 32",
      "framework": "torch",
      "status": "running",
      "startTime": 1707120000000,
      "cwd": "/workspace"
    }
  ],
  "timestamp": "2024-02-05T10:30:00.000Z"
}
```

**响应**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "received": 1,
    "inserted": 1,
    "updated": 0
  }
}
```

#### 2.2.3 参数上报

**接口**: `POST /agent/v1/parameters`

**请求体**：
```json
{
  "jobId": "abc123def456",
  "parameterRaw": "--batch_size 32 --lr 0.001",
  "parameterData": {
    "batch_size": 32,
    "learning_rate": 0.001
  },
  "parameterSource": "command_line",
  "configFilePath": "/workspace/config.yaml",
  "configFileContent": "batch_size: 32\n...",
  "envVars": {
    "CUDA_VISIBLE_DEVICES": "0,1,2,3"
  },
  "timestamp": "2024-02-05T10:30:00.000Z"
}
```

#### 2.2.4 代码上报

**接口**: `POST /agent/v1/code`

**请求体**：
```json
{
  "jobId": "abc123def456",
  "scriptPath": "/workspace/train.py",
  "scriptContent": "import torch\n...",
  "importedLibraries": "torch,transformers,numpy",
  "configFiles": "config.yaml,model_config.json",
  "timestamp": "2024-02-05T10:30:00.000Z"
}
```

#### 2.2.5 NPU指标上报

**接口**: `POST /agent/v1/npu-metrics`

**请求体**：
```json
{
  "nodeId": "a1b2c3d4e5f6",
  "metrics": [
    {
      "npuId": 0,
      "name": "Ascend 910",
      "health": "OK",
      "powerW": 250.5,
      "tempC": 65.2,
      "aicoreUsagePercent": 85.3,
      "memoryUsageMb": 16384,
      "memoryTotalMb": 32768
    }
  ],
  "timestamp": "2024-02-05T10:30:00.000Z"
}
```

#### 2.2.6 进程指标上报

**接口**: `POST /agent/v1/process-metrics`

**请求体**：
```json
{
  "jobId": "abc123def456",
  "metrics": {
    "pid": 12345,
    "cpuPercent": 85.5,
    "memoryMb": 4096,
    "threadCount": 8,
    "openFiles": 128,
    "status": "running"
  },
  "timestamp": "2024-02-05T10:30:00.000Z"
}
```

### 2.3 项目结构

```
task-monitor-agent-server/
├── cmd/
│   └── agent-server/
│       └── main.go              # 入口文件
│
├── internal/
│   ├── handler/                 # HTTP处理器
│   │   ├── heartbeat.go        # 心跳处理
│   │   ├── job.go              # 作业上报处理
│   │   ├── metrics.go          # 指标上报处理
│   │   └── parameter.go        # 参数上报处理
│   │
│   ├── service/                 # 业务逻辑层
│   │   ├── node_service.go     # 节点服务
│   │   ├── job_service.go      # 作业服务
│   │   └── metrics_service.go  # 指标服务
│   │
│   ├── repository/              # 数据访问层
│   │   ├── node_repo.go
│   │   ├── job_repo.go
│   │   └── metrics_repo.go
│   │
│   ├── monitor/                 # Job Monitor
│   │   └── job_monitor.go      # 检测失联作业
│   │
│   └── middleware/              # 中间件
│       ├── auth.go             # Agent认证
│       ├── logger.go           # 日志记录
│       └── recovery.go         # 错误恢复
│
├── configs/
│   └── agent-server.yaml        # 配置文件
│
└── go.mod
```

### 2.4 核心功能实现

#### 2.4.1 批量写入优化

**问题**：Agent频繁上报数据，单条写入效率低

**解决方案**：使用批量写入和缓冲队列

```go
// internal/service/metrics_service.go
type MetricsService struct {
    repo   *repository.MetricsRepository
    buffer chan *model.NPUMetric
}

func (s *MetricsService) Start() {
    // 启动批量写入协程
    go s.batchInsertWorker()
}

func (s *MetricsService) batchInsertWorker() {
    ticker := time.NewTicker(5 * time.Second)
    batch := make([]*model.NPUMetric, 0, 1000)
    
    for {
        select {
        case metric := <-s.buffer:
            batch = append(batch, metric)
            if len(batch) >= 1000 {
                s.repo.BatchInsert(batch)
                batch = batch[:0]
            }
        case <-ticker.C:
            if len(batch) > 0 {
                s.repo.BatchInsert(batch)
                batch = batch[:0]
            }
        }
    }
}
```

#### 2.4.2 Job Monitor机制

**功能**：检测长时间未更新的作业，标记为lost状态

```go
// internal/monitor/job_monitor.go
type JobMonitor struct {
    jobRepo *repository.JobRepository
    interval time.Duration
    timeout  time.Duration
}

func (m *JobMonitor) Start() {
    ticker := time.NewTicker(m.interval)
    
    for range ticker.C {
        // 查找超时未更新的作业
        jobs, err := m.jobRepo.FindStaleJobs(m.timeout)
        if err != nil {
            log.Error("Failed to find stale jobs", err)
            continue
        }
        
        // 标记为lost状态
        for _, job := range jobs {
            m.jobRepo.UpdateStatus(job.JobID, "lost", "job_monitor")
        }
    }
}
```

### 2.5 配置文件

```yaml
# configs/agent-server.yaml
server:
  port: 8081
  mode: release  # debug, release
  
database:
  host: localhost
  port: 3306
  user: root
  password: password
  database: task_monitor
  max_open_conns: 100
  max_idle_conns: 10
  
auth:
  enabled: true
  token: "agent-secret-token"  # Agent认证Token
  
batch:
  buffer_size: 10000
  flush_interval: 5s
  batch_size: 1000
  
job_monitor:
  enabled: true
  check_interval: 60s
  timeout: 300s  # 5分钟未更新标记为lost
  
log:
  level: info
  file: /var/log/agent-server.log
```

## 三、API Server设计

### 3.1 职责定义

**核心职责**：
1. 提供RESTful API接口给前端
2. 查询节点、作业、指标等数据
3. 提供统计分析功能
4. 实现数据导出功能

**非职责**：
- 不接收Agent上报（由Agent Server负责）
- 不直接写入数据库（只读取）

### 3.2 项目结构

```
task-monitor-api-server/
├── cmd/
│   └── api-server/
│       └── main.go              # 入口文件
│
├── internal/
│   ├── handler/                 # HTTP处理器
│   │   ├── node.go             # 节点接口
│   │   ├── job.go              # 作业接口
│   │   ├── metrics.go          # 指标接口
│   │   └── stats.go            # 统计接口
│   │
│   ├── service/                 # 业务逻辑层
│   │   ├── node_service.go
│   │   ├── job_service.go
│   │   ├── metrics_service.go
│   │   └── stats_service.go
│   │
│   ├── repository/              # 数据访问层
│   │   ├── node_repo.go
│   │   ├── job_repo.go
│   │   └── metrics_repo.go
│   │
│   └── middleware/              # 中间件
│       ├── auth.go             # JWT认证
│       ├── cors.go             # CORS处理
│       ├── logger.go           # 日志记录
│       └── cache.go            # 缓存中间件
│
├── configs/
│   └── api-server.yaml          # 配置文件
│
└── go.mod
```

### 3.3 核心功能实现

#### 3.3.1 查询缓存

**问题**：频繁查询数据库，性能压力大

**解决方案**：使用Redis缓存热点数据

```go
// internal/service/node_service.go
type NodeService struct {
    repo  *repository.NodeRepository
    cache *redis.Client
}

func (s *NodeService) GetNodes(ctx context.Context, params *NodeListParams) (*NodeListResponse, error) {
    // 尝试从缓存获取
    cacheKey := fmt.Sprintf("nodes:list:%v", params)
    cached, err := s.cache.Get(ctx, cacheKey).Result()
    if err == nil {
        var result NodeListResponse
        json.Unmarshal([]byte(cached), &result)
        return &result, nil
    }
    
    // 缓存未命中，查询数据库
    result, err := s.repo.FindNodes(params)
    if err != nil {
        return nil, err
    }
    
    // 写入缓存，TTL 30秒
    data, _ := json.Marshal(result)
    s.cache.Set(ctx, cacheKey, data, 30*time.Second)
    
    return result, nil
}
```

#### 3.3.2 分页查询优化

```go
// internal/repository/job_repo.go
func (r *JobRepository) FindJobs(params *JobListParams) (*JobListResponse, error) {
    query := r.db.Model(&model.Job{})
    
    // 筛选条件
    if len(params.Status) > 0 {
        query = query.Where("status IN ?", params.Status)
    }
    if len(params.Type) > 0 {
        query = query.Where("job_type IN ?", params.Type)
    }
    if params.NodeID != "" {
        query = query.Where("node_id = ?", params.NodeID)
    }
    
    // 计算总数
    var total int64
    query.Count(&total)
    
    // 分页查询
    offset := (params.Page - 1) * params.PageSize
    var jobs []model.Job
    query.Offset(offset).Limit(params.PageSize).Find(&jobs)
    
    return &JobListResponse{
        Items: jobs,
        Pagination: Pagination{
            Page:       params.Page,
            PageSize:   params.PageSize,
            Total:      total,
            TotalPages: (total + int64(params.PageSize) - 1) / int64(params.PageSize),
        },
    }, nil
}
```

### 3.4 配置文件

```yaml
# configs/api-server.yaml
server:
  port: 8080
  mode: release
  
database:
  host: localhost
  port: 3306
  user: root
  password: password
  database: task_monitor
  max_open_conns: 50
  max_idle_conns: 10
  
redis:
  enabled: true
  host: localhost
  port: 6379
  password: ""
  db: 0
  
auth:
  jwt_secret: "api-jwt-secret"
  token_expire: 3600  # 1小时
  
cors:
  enabled: true
  allowed_origins:
    - "http://localhost:3000"
    - "https://monitor.example.com"
  
log:
  level: info
  file: /var/log/api-server.log
```

## 四、部署方案

### 4.1 部署架构

```
┌─────────────────────────────────────────────────────────────┐
│                      生产环境部署                             │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    计算节点集群 (内网)                        │
│  Node1 (Agent) → Node2 (Agent) → ... → NodeN (Agent)       │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP POST
                         │ Port: 8081
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  Agent Server (内网)                         │
│  - 接收Agent上报                                             │
│  - 批量写入数据库                                             │
│  - Job Monitor                                              │
│  - 部署: Docker/Systemd                                     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ MySQL
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  MySQL Database                             │
│  - 主从复制 (可选)                                           │
│  - 定期备份                                                  │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ Read
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  API Server (外网/内网)                      │
│  - 提供RESTful API                                          │
│  - Redis缓存                                                │
│  - JWT认证                                                  │
│  - 部署: Docker/Kubernetes                                  │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP GET
                         │ Port: 8080
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  Nginx (反向代理)                            │
│  - SSL终止                                                  │
│  - 负载均衡                                                  │
│  - 访问控制                                                  │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTPS
                         ▼
                  ┌──────────────┐
                  │   Frontend   │
                  │   (React)    │
                  └──────────────┘
```

### 4.2 Docker部署

**Agent Server Dockerfile**:
```dockerfile
# Dockerfile.agent-server
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o agent-server ./cmd/agent-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/agent-server .
COPY --from=builder /app/configs/agent-server.yaml ./configs/

EXPOSE 8081
CMD ["./agent-server"]
```

**API Server Dockerfile**:
```dockerfile
# Dockerfile.api-server
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o api-server ./cmd/api-server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/api-server .
COPY --from=builder /app/configs/api-server.yaml ./configs/

EXPOSE 8080
CMD ["./api-server"]
```

**Docker Compose**:
```yaml
# docker-compose.yml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: task_monitor
    volumes:
      - mysql_data:/var/lib/mysql
    ports:
      - "3306:3306"
    networks:
      - task-monitor

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    networks:
      - task-monitor

  agent-server:
    build:
      context: .
      dockerfile: Dockerfile.agent-server
    ports:
      - "8081:8081"
    depends_on:
      - mysql
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306
    networks:
      - task-monitor

  api-server:
    build:
      context: .
      dockerfile: Dockerfile.api-server
    ports:
      - "8080:8080"
    depends_on:
      - mysql
      - redis
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    networks:
      - task-monitor

volumes:
  mysql_data:

networks:
  task-monitor:
    driver: bridge
```

### 4.3 Systemd部署

**Agent Server Service**:
```ini
# /etc/systemd/system/agent-server.service
[Unit]
Description=Task Monitor Agent Server
After=network.target mysql.service

[Service]
Type=simple
User=taskmonitor
WorkingDirectory=/opt/task-monitor/agent-server
ExecStart=/opt/task-monitor/agent-server/agent-server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

**API Server Service**:
```ini
# /etc/systemd/system/api-server.service
[Unit]
Description=Task Monitor API Server
After=network.target mysql.service redis.service

[Service]
Type=simple
User=taskmonitor
WorkingDirectory=/opt/task-monitor/api-server
ExecStart=/opt/task-monitor/api-server/api-server
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

## 五、数据流和通信

### 5.1 Agent → Agent Server

**通信方式**: HTTP POST
**频率**: 
- 心跳: 每30秒
- 作业上报: 每30秒
- 指标上报: 每30秒

**数据流**:
```
Agent (Node1)
  ├─ 采集作业信息 → POST /agent/v1/jobs
  ├─ 采集NPU指标 → POST /agent/v1/npu-metrics
  ├─ 采集进程指标 → POST /agent/v1/process-metrics
  └─ 发送心跳 → POST /agent/v1/heartbeat
```

### 5.2 Frontend → API Server

**通信方式**: HTTP GET/POST
**认证**: JWT Token

**数据流**:
```
Frontend
  ├─ 查询节点列表 → GET /api/v1/nodes
  ├─ 查询作业列表 → GET /api/v1/jobs
  ├─ 查询作业详情 → GET /api/v1/jobs/{id}
  ├─ 查询NPU指标 → GET /api/v1/metrics/npu
  └─ 查询统计数据 → GET /api/v1/stats/cluster
```

### 5.3 数据库访问模式

**Agent Server (写)**:
- 批量插入作业、参数、代码、指标数据
- 更新节点心跳时间
- 更新作业状态

**API Server (读)**:
- 查询节点、作业、指标数据
- 聚合统计数据
- 使用Redis缓存热点数据

## 六、监控和运维

### 6.1 服务监控

**关键指标**:
- Agent Server:
  - 接收请求QPS
  - 数据库写入延迟
  - 批量写入队列长度
  - Job Monitor检测到的失联作业数

- API Server:
  - API请求QPS
  - 响应时间P50/P95/P99
  - 缓存命中率
  - 数据库查询延迟

**监控工具**:
- Prometheus: 指标采集
- Grafana: 可视化展示
- Alertmanager: 告警通知

### 6.2 日志管理

**日志级别**:
- DEBUG: 开发调试
- INFO: 正常运行日志
- WARN: 警告信息
- ERROR: 错误信息

**日志收集**:
- 使用ELK Stack (Elasticsearch + Logstash + Kibana)
- 或使用Loki + Grafana

### 6.3 备份策略

**数据库备份**:
- 全量备份: 每天凌晨2点
- 增量备份: 每小时
- 备份保留: 7天

**配置备份**:
- 配置文件纳入Git管理
- 定期备份到远程存储

## 七、总结

### 7.1 架构优势

1. **职责清晰**: Agent Server专注写入，API Server专注查询
2. **安全隔离**: Agent Server内网部署，API Server可暴露外网
3. **独立扩展**: 根据负载独立扩展两个服务
4. **性能优化**: 批量写入、查询缓存、读写分离
5. **易于维护**: 模块化设计，便于开发和调试

### 7.2 技术栈

**后端**:
- 语言: Go 1.21+
- 框架: Gin
- ORM: GORM
- 数据库: MySQL 8.0
- 缓存: Redis 7.0

**前端**:
- 框架: React 18 + TypeScript
- UI库: Ant Design 5.x
- 状态管理: React Query + Zustand

### 7.3 下一步工作

**第一阶段 (核心功能)**:
1. 实现Agent Server核心接口
2. 实现API Server核心接口
3. 完善数据库操作层
4. 实现Job Monitor机制

**第二阶段 (优化完善)**:
1. 添加Redis缓存
2. 实现批量写入优化
3. 添加单元测试
4. 性能测试和优化

**第三阶段 (部署上线)**:
1. 编写部署文档
2. 配置监控告警
3. 生产环境部署
4. 运维文档完善

---

## 附录

### A. 相关文档

- [FRONTEND_DESIGN.md](FRONTEND_DESIGN.md) - 前端架构设计
- [API_DESIGN.md](API_DESIGN.md) - API接口设计
- [DATABASE.md](task_monitor_go/DATABASE.md) - 数据库设计
- [README.md](README.md) - 项目说明

### B. 端口规划

| 服务 | 端口 | 说明 |
|------|------|------|
| Agent Server | 8081 | Agent上报接口 |
| API Server | 8080 | RESTful API接口 |
| MySQL | 3306 | 数据库 |
| Redis | 6379 | 缓存 |
| Frontend | 3000 | 前端开发服务器 |
| Nginx | 80/443 | 反向代理 |

---

**文档版本**: v1.0.0
**最后更新**: 2024-02-05
**维护者**: Task Monitor Team
