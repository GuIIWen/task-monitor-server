# NPU作业监控系统 - API接口设计文档

## 文档信息
- **版本**: v1.0.0
- **创建日期**: 2024-02-05
- **API版本**: v1
- **基础路径**: `/api/v1`

## 一、API设计原则

### 1.1 RESTful规范

- 使用标准HTTP方法：GET（查询）、POST（创建）、PUT（更新）、DELETE（删除）
- 使用复数名词表示资源：`/nodes`、`/jobs`、`/metrics`
- 使用路径参数表示资源ID：`/nodes/{nodeId}`
- 使用查询参数进行筛选和分页：`?status=active&page=1`

### 1.2 统一响应格式

**成功响应**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    // 实际数据
  }
}
```

**分页响应**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 100,
      "totalPages": 5
    }
  }
}
```

**错误响应**：
```json
{
  "code": 400,
  "message": "Invalid parameters",
  "error": {
    "field": "nodeId",
    "reason": "nodeId is required"
  }
}
```

### 1.3 HTTP状态码

- `200 OK` - 请求成功
- `201 Created` - 资源创建成功
- `400 Bad Request` - 请求参数错误
- `401 Unauthorized` - 未授权
- `403 Forbidden` - 禁止访问
- `404 Not Found` - 资源不存在
- `500 Internal Server Error` - 服务器内部错误

## 二、节点管理API

### 2.1 获取节点列表

**接口**: `GET /api/v1/nodes`

**描述**: 获取所有节点的列表，支持筛选和分页

**请求参数**：
| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| status | string | 否 | 节点状态筛选 | active, inactive, error |
| page | integer | 否 | 页码，默认1 | 1 |
| pageSize | integer | 否 | 每页数量，默认20 | 20 |
| sortBy | string | 否 | 排序字段 | hostname, lastHeartbeat |
| sortOrder | string | 否 | 排序方向 | asc, desc |

**请求示例**：
```bash
GET /api/v1/nodes?status=active&page=1&pageSize=20
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "nodeId": "a1b2c3d4e5f6",
        "hostId": "host-001",
        "hostname": "gpu-node-01",
        "ipAddress": "192.168.1.100",
        "npuCount": 8,
        "status": "active",
        "lastHeartbeat": "2024-02-05T10:30:00.000Z",
        "createdAt": "2024-02-01T08:00:00.000Z",
        "updatedAt": "2024-02-05T10:30:00.000Z",
        "stats": {
          "runningJobs": 3,
          "avgNpuUsage": 75.5,
          "healthyNpus": 8
        }
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 12,
      "totalPages": 1
    }
  }
}
```

### 2.2 获取节点详情

**接口**: `GET /api/v1/nodes/{nodeId}`

**描述**: 获取指定节点的详细信息

**路径参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| nodeId | string | 是 | 节点ID |

**请求示例**：
```bash
GET /api/v1/nodes/a1b2c3d4e5f6
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "nodeId": "a1b2c3d4e5f6",
    "hostId": "host-001",
    "hostname": "gpu-node-01",
    "ipAddress": "192.168.1.100",
    "npuCount": 8,
    "status": "active",
    "lastHeartbeat": "2024-02-05T10:30:00.000Z",
    "createdAt": "2024-02-01T08:00:00.000Z",
    "updatedAt": "2024-02-05T10:30:00.000Z"
  }
}
```

### 2.3 获取节点的NPU指标

**接口**: `GET /api/v1/nodes/{nodeId}/npu-metrics`

**描述**: 获取指定节点的NPU设备指标数据

**路径参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| nodeId | string | 是 | 节点ID |

**查询参数**：
| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| startTime | string | 否 | 开始时间（ISO 8601） | 2024-02-05T00:00:00Z |
| endTime | string | 否 | 结束时间（ISO 8601） | 2024-02-05T23:59:59Z |
| npuId | integer | 否 | 指定NPU设备ID | 0 |
| interval | string | 否 | 数据聚合间隔 | 1m, 5m, 1h |

**请求示例**：
```bash
GET /api/v1/nodes/a1b2c3d4e5f6/npu-metrics?startTime=2024-02-05T00:00:00Z&interval=5m
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
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
        "memoryTotalMb": 32768,
        "hbmUsageMb": 24576,
        "hbmTotalMb": 32768,
        "busId": "0000:01:00.0",
        "timestamp": "2024-02-05T10:30:00.000Z"
      }
    ]
  }
}
```

### 2.4 获取节点的运行作业

**接口**: `GET /api/v1/nodes/{nodeId}/jobs`

**描述**: 获取指定节点上运行的作业列表

**路径参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| nodeId | string | 是 | 节点ID |

**查询参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| status | string | 否 | 作业状态筛选 | running, completed |
| page | integer | 否 | 页码 | 1 |
| pageSize | integer | 否 | 每页数量 | 20 |

**请求示例**：
```bash
GET /api/v1/nodes/a1b2c3d4e5f6/jobs?status=running
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "jobId": "abc123def456",
        "jobName": "train_model.py",
        "jobType": "training",
        "status": "running",
        "framework": "torch",
        "startTime": 1707120000000,
        "pid": 12345
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 3,
      "totalPages": 1
    }
  }
}
```

## 三、作业管理API

### 3.1 获取作业列表

**接口**: `GET /api/v1/jobs`

**描述**: 获取所有作业的列表，支持多维度筛选和分页

**请求参数**：
| 参数名 | 类型 | 必填 | 说明 | 示例 |
|--------|------|------|------|------|
| status | string[] | 否 | 作业状态筛选（多选） | running,completed |
| type | string[] | 否 | 作业类型筛选（多选） | training,inference |
| framework | string[] | 否 | 框架筛选（多选） | torch,transformers |
| nodeId | string | 否 | 节点ID筛选 | a1b2c3d4e5f6 |
| startTime | string | 否 | 开始时间范围（起） | 2024-02-05T00:00:00Z |
| endTime | string | 否 | 开始时间范围（止） | 2024-02-05T23:59:59Z |
| search | string | 否 | 搜索关键词（作业名） | train |
| page | integer | 否 | 页码，默认1 | 1 |
| pageSize | integer | 否 | 每页数量，默认20 | 20 |
| sortBy | string | 否 | 排序字段 | startTime, jobName |
| sortOrder | string | 否 | 排序方向 | asc, desc |

**请求示例**：
```bash
GET /api/v1/jobs?status=running&type=training&page=1&pageSize=20&sortBy=startTime&sortOrder=desc
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "jobId": "abc123def456",
        "nodeId": "a1b2c3d4e5f6",
        "jobName": "train_model.py",
        "jobType": "training",
        "pid": 12345,
        "ppid": 1000,
        "pgid": 12345,
        "processName": "python",
        "framework": "torch",
        "modelFormat": "pt",
        "status": "running",
        "startTime": 1707120000000,
        "endTime": null,
        "cwd": "/workspace",
        "createdAt": "2024-02-05T10:30:00.000Z",
        "updatedAt": "2024-02-05T12:30:00.000Z",
        "node": {
          "hostname": "gpu-node-01",
          "ipAddress": "192.168.1.100"
        },
        "latestMetrics": {
          "cpuPercent": 85.5,
          "memoryMb": 4096,
          "timestamp": "2024-02-05T12:30:00.000Z"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 45,
      "totalPages": 3
    }
  }
}
```

### 3.2 获取作业详情

**接口**: `GET /api/v1/jobs/{jobId}`

**描述**: 获取指定作业的详细信息

**路径参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| jobId | string | 是 | 作业ID |

**请求示例**：
```bash
GET /api/v1/jobs/abc123def456
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "jobId": "abc123def456",
    "nodeId": "a1b2c3d4e5f6",
    "hostId": "host-001",
    "jobName": "train_model.py",
    "jobType": "training",
    "pid": 12345,
    "ppid": 1000,
    "pgid": 12345,
    "processName": "python",
    "commandLine": "python train.py --batch_size 32 --learning_rate 0.001",
    "framework": "torch",
    "modelFormat": "pt",
    "status": "running",
    "startTime": 1707120000000,
    "endTime": null,
    "cwd": "/workspace",
    "createdAt": "2024-02-05T10:30:00.000Z",
    "updatedAt": "2024-02-05T12:30:00.000Z",
    "node": {
      "nodeId": "a1b2c3d4e5f6",
      "hostname": "gpu-node-01",
      "ipAddress": "192.168.1.100"
    }
  }
}
```

### 3.3 获取作业参数

**接口**: `GET /api/v1/jobs/{jobId}/parameters`

**描述**: 获取指定作业的参数配置信息

**路径参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| jobId | string | 是 | 作业ID |

**请求示例**：
```bash
GET /api/v1/jobs/abc123def456/parameters
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "jobId": "abc123def456",
        "parameterRaw": "--batch_size 32 --learning_rate 0.001 --epochs 100",
        "parameterData": {
          "batch_size": 32,
          "learning_rate": 0.001,
          "epochs": 100
        },
        "parameterSource": "command_line",
        "configFilePath": "/workspace/config.yaml",
        "configFileContent": "batch_size: 32\nlearning_rate: 0.001\n...",
        "envVars": {
          "CUDA_VISIBLE_DEVICES": "0,1,2,3",
          "PYTHONPATH": "/workspace"
        },
        "timestamp": "2024-02-05T10:30:00.000Z"
      }
    ]
  }
}
```

### 3.4 获取作业代码信息

**接口**: `GET /api/v1/jobs/{jobId}/code`

**描述**: 获取指定作业的代码信息

**路径参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| jobId | string | 是 | 作业ID |

**请求示例**：
```bash
GET /api/v1/jobs/abc123def456/code
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "jobId": "abc123def456",
        "scriptPath": "/workspace/train.py",
        "scriptContent": "import torch\nimport transformers\n...",
        "importedLibraries": "torch,transformers,numpy,pandas",
        "configFiles": "config.yaml,model_config.json",
        "shScriptPath": "/workspace/run.sh",
        "shScriptContent": "#!/bin/bash\npython train.py\n...",
        "timestamp": "2024-02-05T10:30:00.000Z"
      }
    ]
  }
}
```

### 3.5 获取作业进程指标

**接口**: `GET /api/v1/jobs/{jobId}/process-metrics`

**描述**: 获取指定作业的进程资源使用指标

**路径参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| jobId | string | 是 | 作业ID |

**查询参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| startTime | string | 否 | 开始时间 | 2024-02-05T00:00:00Z |
| endTime | string | 否 | 结束时间 | 2024-02-05T23:59:59Z |
| interval | string | 否 | 数据聚合间隔 | 1m, 5m, 1h |

**请求示例**：
```bash
GET /api/v1/jobs/abc123def456/process-metrics?startTime=2024-02-05T10:00:00Z&interval=5m
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "jobId": "abc123def456",
    "metrics": [
      {
        "id": 1,
        "jobId": "abc123def456",
        "pid": 12345,
        "cpuPercent": 85.5,
        "memoryMb": 4096,
        "threadCount": 8,
        "openFiles": 128,
        "status": "running",
        "timestamp": "2024-02-05T10:30:00.000Z"
      }
    ]
  }
}
```

### 3.6 获取作业状态历史

**接口**: `GET /api/v1/jobs/{jobId}/status-history`

**描述**: 获取指定作业的状态变更历史

**路径参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| jobId | string | 是 | 作业ID |

**请求示例**：
```bash
GET /api/v1/jobs/abc123def456/status-history
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "items": [
      {
        "id": 1,
        "jobId": "abc123def456",
        "oldStatus": null,
        "newStatus": "running",
        "reason": "agent_report",
        "changedAt": "2024-02-05T10:30:00.000Z"
      },
      {
        "id": 2,
        "jobId": "abc123def456",
        "oldStatus": "running",
        "newStatus": "lost",
        "reason": "job_monitor",
        "changedAt": "2024-02-05T12:00:00.000Z"
      }
    ]
  }
}
```

### 3.7 停止作业

**接口**: `POST /api/v1/jobs/{jobId}/stop`

**描述**: 停止指定的作业

**路径参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| jobId | string | 是 | 作业ID |

**请求示例**：
```bash
POST /api/v1/jobs/abc123def456/stop
```

**响应示例**：
```json
{
  "code": 200,
  "message": "Job stopped successfully",
  "data": {
    "jobId": "abc123def456",
    "status": "stopped"
  }
}
```

### 3.8 批量停止作业

**接口**: `POST /api/v1/jobs/batch-stop`

**描述**: 批量停止多个作业

**请求体**：
```json
{
  "jobIds": ["abc123def456", "def456ghi789"]
}
```

**请求示例**：
```bash
POST /api/v1/jobs/batch-stop
Content-Type: application/json

{
  "jobIds": ["abc123def456", "def456ghi789"]
}
```

**响应示例**：
```json
{
  "code": 200,
  "message": "Batch stop completed",
  "data": {
    "success": ["abc123def456"],
    "failed": [
      {
        "jobId": "def456ghi789",
        "reason": "Job not found"
      }
    ]
  }
}
```

## 四、监控指标API

### 4.1 获取NPU指标

**接口**: `GET /api/v1/metrics/npu`

**描述**: 获取NPU设备的监控指标数据

**请求参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| nodeId | string | 否 | 节点ID筛选 |
| npuId | integer | 否 | NPU设备ID筛选 |
| startTime | string | 否 | 开始时间 |
| endTime | string | 否 | 结束时间 |
| interval | string | 否 | 数据聚合间隔 |

**请求示例**：
```bash
GET /api/v1/metrics/npu?nodeId=a1b2c3d4e5f6&startTime=2024-02-05T10:00:00Z&interval=5m
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "metrics": [
      {
        "id": 1,
        "nodeId": "a1b2c3d4e5f6",
        "npuId": 0,
        "name": "Ascend 910",
        "health": "OK",
        "powerW": 250.5,
        "tempC": 65.2,
        "aicoreUsagePercent": 85.3,
        "memoryUsageMb": 16384,
        "memoryTotalMb": 32768,
        "hbmUsageMb": 24576,
        "hbmTotalMb": 32768,
        "busId": "0000:01:00.0",
        "timestamp": "2024-02-05T10:30:00.000Z"
      }
    ]
  }
}
```

### 4.2 获取NPU进程信息

**接口**: `GET /api/v1/metrics/npu-processes`

**描述**: 获取运行在NPU上的进程信息

**请求参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| nodeId | string | 否 | 节点ID筛选 |
| npuId | integer | 否 | NPU设备ID筛选 |
| startTime | string | 否 | 开始时间 |
| endTime | string | 否 | 结束时间 |

**请求示例**：
```bash
GET /api/v1/metrics/npu-processes?nodeId=a1b2c3d4e5f6&npuId=0
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "processes": [
      {
        "id": 1,
        "nodeId": "a1b2c3d4e5f6",
        "npuId": 0,
        "pid": 12345,
        "processName": "python",
        "memoryUsageMb": 8192,
        "timestamp": "2024-02-05T10:30:00.000Z"
      }
    ]
  }
}
```

## 五、统计分析API

### 5.1 获取集群整体统计

**接口**: `GET /api/v1/stats/cluster`

**描述**: 获取集群的整体统计信息

**请求示例**：
```bash
GET /api/v1/stats/cluster
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "totalNodes": 12,
    "activeNodes": 10,
    "inactiveNodes": 2,
    "totalJobs": 45,
    "runningJobs": 38,
    "completedJobs": 5,
    "failedJobs": 2,
    "totalNPUs": 96,
    "healthyNPUs": 94,
    "avgNPUUsage": 65.5,
    "avgNPUTemp": 62.3,
    "avgNPUPower": 245.8,
    "jobTypeDistribution": {
      "training": 30,
      "inference": 12,
      "testing": 3
    },
    "frameworkDistribution": {
      "torch": 25,
      "transformers": 15,
      "mindspore": 5
    },
    "timestamp": "2024-02-05T12:30:00.000Z"
  }
}
```

### 5.2 获取节点统计

**接口**: `GET /api/v1/stats/nodes`

**描述**: 获取所有节点的统计信息

**请求示例**：
```bash
GET /api/v1/stats/nodes
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "nodes": [
      {
        "nodeId": "a1b2c3d4e5f6",
        "hostname": "gpu-node-01",
        "status": "active",
        "totalJobs": 3,
        "runningJobs": 3,
        "totalNPUs": 8,
        "healthyNPUs": 8,
        "avgNPUUsage": 75.5,
        "avgNPUTemp": 65.2,
        "avgNPUPower": 250.5
      }
    ]
  }
}
```

### 5.3 获取作业统计

**接口**: `GET /api/v1/stats/jobs`

**描述**: 获取作业的统计信息

**请求参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| startTime | string | 否 | 统计开始时间 |
| endTime | string | 否 | 统计结束时间 |
| groupBy | string | 否 | 分组维度：type, framework, node |

**请求示例**：
```bash
GET /api/v1/stats/jobs?groupBy=type
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 45,
    "byType": {
      "training": 30,
      "inference": 12,
      "testing": 3
    },
    "byStatus": {
      "running": 38,
      "completed": 5,
      "failed": 2
    },
    "byFramework": {
      "torch": 25,
      "transformers": 15,
      "mindspore": 5
    }
  }
}
```

### 5.4 获取趋势数据

**接口**: `GET /api/v1/stats/trends`

**描述**: 获取指定时间范围内的趋势数据

**请求参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| metric | string | 是 | 指标类型：npu_usage, job_count, node_count |
| startTime | string | 是 | 开始时间 |
| endTime | string | 是 | 结束时间 |
| interval | string | 否 | 数据间隔：1m, 5m, 1h, 1d |

**请求示例**：
```bash
GET /api/v1/stats/trends?metric=npu_usage&startTime=2024-02-05T00:00:00Z&endTime=2024-02-05T23:59:59Z&interval=1h
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "metric": "npu_usage",
    "interval": "1h",
    "dataPoints": [
      {
        "timestamp": "2024-02-05T00:00:00.000Z",
        "value": 45.5
      },
      {
        "timestamp": "2024-02-05T01:00:00.000Z",
        "value": 52.3
      }
    ]
  }
}
```

## 六、其他辅助API

### 6.1 数据导出

**接口**: `POST /api/v1/export/jobs`

**描述**: 导出作业数据为CSV格式

**请求体**：
```json
{
  "jobIds": ["abc123def456", "def456ghi789"],
  "fields": ["jobName", "status", "framework", "startTime"],
  "format": "csv"
}
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "downloadUrl": "/api/v1/downloads/jobs_20240205.csv",
    "expiresAt": "2024-02-05T13:00:00.000Z"
  }
}
```

### 6.2 全局搜索

**接口**: `GET /api/v1/search`

**描述**: 全局搜索节点、作业等资源

**请求参数**：
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| q | string | 是 | 搜索关键词 |
| type | string | 否 | 资源类型：node, job, all |
| limit | integer | 否 | 返回数量限制 |

**请求示例**：
```bash
GET /api/v1/search?q=train&type=job&limit=10
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "jobs": [
      {
        "jobId": "abc123def456",
        "jobName": "train_model.py",
        "status": "running"
      }
    ],
    "nodes": [],
    "total": 1
  }
}
```

## 七、错误码定义

| 错误码 | HTTP状态码 | 说明 |
|--------|-----------|------|
| 200 | 200 | 请求成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未授权，需要登录 |
| 403 | 403 | 禁止访问，权限不足 |
| 404 | 404 | 资源不存在 |
| 409 | 409 | 资源冲突 |
| 500 | 500 | 服务器内部错误 |
| 503 | 503 | 服务暂时不可用 |

**错误响应示例**：
```json
{
  "code": 404,
  "message": "Job not found",
  "error": {
    "field": "jobId",
    "value": "invalid-job-id",
    "reason": "The specified job does not exist"
  }
}
```

## 八、认证和授权

### 8.1 认证方式

**Bearer Token认证**：
```bash
GET /api/v1/nodes
Authorization: Bearer <token>
```

### 8.2 获取Token

**接口**: `POST /api/v1/auth/login`

**请求体**：
```json
{
  "username": "admin",
  "password": "password"
}
```

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 3600,
    "refreshToken": "refresh_token_here"
  }
}
```

## 九、API使用示例

### 9.1 获取集群概览

```bash
# 1. 获取集群统计
curl -X GET "http://localhost:8080/api/v1/stats/cluster" \
  -H "Authorization: Bearer <token>"

# 2. 获取节点列表
curl -X GET "http://localhost:8080/api/v1/nodes?status=active" \
  -H "Authorization: Bearer <token>"

# 3. 获取运行中的作业
curl -X GET "http://localhost:8080/api/v1/jobs?status=running&page=1&pageSize=20" \
  -H "Authorization: Bearer <token>"
```

### 9.2 查看作业详情

```bash
# 1. 获取作业基本信息
curl -X GET "http://localhost:8080/api/v1/jobs/abc123def456" \
  -H "Authorization: Bearer <token>"

# 2. 获取作业参数
curl -X GET "http://localhost:8080/api/v1/jobs/abc123def456/parameters" \
  -H "Authorization: Bearer <token>"

# 3. 获取作业进程指标
curl -X GET "http://localhost:8080/api/v1/jobs/abc123def456/process-metrics?startTime=2024-02-05T10:00:00Z" \
  -H "Authorization: Bearer <token>"
```

### 9.3 监控NPU资源

```bash
# 1. 获取节点的NPU指标
curl -X GET "http://localhost:8080/api/v1/nodes/a1b2c3d4e5f6/npu-metrics?interval=5m" \
  -H "Authorization: Bearer <token>"

# 2. 获取NPU进程信息
curl -X GET "http://localhost:8080/api/v1/metrics/npu-processes?nodeId=a1b2c3d4e5f6" \
  -H "Authorization: Bearer <token>"
```

## 十、总结

### 10.1 API设计特点

1. **RESTful规范**：遵循REST API设计最佳实践
2. **统一响应格式**：所有接口返回格式一致，便于前端处理
3. **完善的筛选和分页**：支持多维度筛选和分页查询
4. **清晰的错误处理**：统一的错误码和错误信息
5. **灵活的数据聚合**：支持时间间隔聚合，减少数据传输量

### 10.2 实现建议

1. **使用Go标准库和Gin框架**：高性能的HTTP服务
2. **数据库查询优化**：使用索引、分页、缓存等优化手段
3. **API版本管理**：通过路径前缀管理API版本
4. **接口文档自动生成**：使用Swagger/OpenAPI生成文档
5. **API限流和熔断**：防止接口被滥用

### 10.3 下一步工作

1. **实现核心API接口**：优先实现节点和作业管理API
2. **添加单元测试**：确保API接口的正确性
3. **性能测试**：测试高并发场景下的性能
4. **API文档生成**：使用Swagger生成交互式文档
5. **前后端联调**：与前端团队协作完成接口对接

---

## 附录

### A. 相关文档

- [FRONTEND_DESIGN.md](FRONTEND_DESIGN.md) - 前端架构设计文档
- [DATABASE.md](task_monitor_go/DATABASE.md) - 数据库设计文档
- [README.md](README.md) - 项目说明文档

### B. API测试工具

- **Postman**: 图形化API测试工具
- **curl**: 命令行HTTP客户端
- **httpie**: 更友好的命令行HTTP客户端

---

**文档版本**: v1.0.0
**最后更新**: 2024-02-05
**维护者**: Task Monitor Backend Team
