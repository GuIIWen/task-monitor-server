# NPU作业监控系统 - API Server

## 项目概述

本项目是NPU作业监控系统的API Server，采用独立架构设计：

- **API Server**: 提供RESTful API接口
  - 查询节点、作业、参数、代码等数据
  - 提供统计分析功能
  - 独立的数据模型和数据访问层
  - 完整的单元测试和集成测试

**注意**: Agent Server位于独立项目 `/root/task_monitor/`，负责接收Agent上报的数据并写入数据库。API Server和Agent Server通过MySQL数据库进行数据交互，两个项目完全解耦。

## 技术栈

- **语言**: Go 1.21+
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.0
- **缓存**: Redis 7.0

## 项目结构

```
task-monitor-server/
├── api-server/                 # API Server模块
│   ├── cmd/
│   │   └── api-server/        # API Server入口
│   │       └── main.go
│   ├── internal/
│   │   ├── config/            # 配置管理
│   │   ├── handler/           # HTTP处理器
│   │   ├── service/           # 业务逻辑
│   │   ├── repository/        # 数据访问层
│   │   ├── model/             # 数据模型
│   │   ├── utils/             # 工具函数
│   │   └── middleware/        # 中间件
│   ├── configs/
│   │   └── api-server.yaml    # 配置文件
│   ├── bin/                   # 编译输出目录
│   └── go.mod
│
├── ARCHITECTURE.md             # 架构设计文档
├── API_DESIGN.md              # API接口设计文档
├── FRONTEND_DESIGN.md         # 前端架构设计文档
└── README.md                  # 本文件
```

## 快速开始

### 1. 环境准备

```bash
# 安装Go 1.21+
# 安装MySQL 8.0
```

### 2. 数据库初始化

```bash
# 创建数据库
mysql -u root -p
CREATE DATABASE task_monitor CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 导入表结构（参考 /root/task_monitor/DATABASE.md）
```

### 3. 配置

编辑 `api-server/configs/api-server.yaml`，配置数据库连接信息。

### 4. 编译和运行

```bash
cd api-server

# 使用代理镜像下载依赖
export GOPROXY=https://goproxy.cn,direct
go mod download

# 编译
go build -o bin/api-server ./cmd/api-server

# 运行
./bin/api-server
```

服务将在 `http://localhost:8080` 启动。

### 5. 运行测试

```bash
cd api-server

# 运行所有测试
export GOPROXY=https://goproxy.cn,direct
go test ./...

# 运行特定模块的测试
go test ./internal/repository/ -v
go test ./internal/service/ -v
go test ./internal/handler/ -v
```

## 配置说明

### API Server配置

```yaml
# api-server/configs/api-server.yaml
server:
  port: 8080
  mode: release  # debug, release, test

database:
  host: localhost
  port: 3306
  user: root
  password: password
  database: task_monitor
  max_idle_conns: 10
  max_open_conns: 100
```

## API接口

API Server提供以下RESTful接口：

### 节点相关
- `GET /api/v1/nodes` - 获取节点列表
  - 查询参数: `status` (可选) - 按状态筛选
- `GET /api/v1/nodes/:nodeId` - 获取节点详情

### 作业相关
- `GET /api/v1/jobs` - 获取作业列表
  - 查询参数: `nodeId` 或 `status` (必须提供其中之一)
- `GET /api/v1/jobs/:jobId` - 获取作业详情
- `GET /api/v1/jobs/:jobId/parameters` - 获取作业参数
- `GET /api/v1/jobs/:jobId/code` - 获取作业代码

### 健康检查
- `GET /health` - 健康检查接口

详细的API文档请参考 [API_DESIGN.md](API_DESIGN.md)。

## 开发指南

### 代码架构

API Server采用分层架构：

1. **Handler层** (`internal/handler/`): 处理HTTP请求和响应
2. **Service层** (`internal/service/`): 业务逻辑处理
3. **Repository层** (`internal/repository/`): 数据访问层
4. **Model层** (`internal/model/`): 数据模型定义

所有层都定义了接口，便于单元测试和依赖注入。

### 添加新的API接口

1. 在 `api-server/internal/model/` 定义数据模型（如需要）
2. 在 `api-server/internal/repository/` 添加数据访问方法
3. 在 `api-server/internal/service/` 添加业务逻辑
4. 在 `api-server/internal/handler/` 添加HTTP处理器
5. 在 `cmd/api-server/main.go` 注册路由
6. 编写单元测试和集成测试

### 测试

项目包含完整的测试套件：

- **Repository测试**: 使用sqlmock模拟数据库操作
- **Service测试**: 使用mock接口测试业务逻辑
- **Handler测试**: 使用httptest测试HTTP接口

运行测试：
```bash
# 运行所有测试
go test ./...

# 运行特定模块测试
go test ./internal/repository/ -v
go test ./internal/service/ -v
go test ./internal/handler/ -v

# 查看测试覆盖率
go test ./... -cover
```

## 部署

参考 [ARCHITECTURE.md](ARCHITECTURE.md) 中的部署方案。

## 文档

- [ARCHITECTURE.md](ARCHITECTURE.md) - 架构设计文档
- [API_DESIGN.md](API_DESIGN.md) - API接口设计文档
- [FRONTEND_DESIGN.md](FRONTEND_DESIGN.md) - 前端架构设计文档

## 许可证

MIT License
