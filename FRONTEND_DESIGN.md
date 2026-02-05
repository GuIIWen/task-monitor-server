# NPU作业监控系统 - 前端架构设计方案

## 文档信息
- **版本**: v1.0.0
- **创建日期**: 2024-02-05
- **设计目标**: 为NPU作业监控系统设计一个直观、美观的管理控制台前端应用

## 一、架构概览

### 1.1 技术栈选型

**核心框架**：
- **React 18** + **TypeScript 5.x** - 类型安全的组件化开发
- **Vite** - 快速的构建工具
- **React Router v6** - 路由管理

**UI组件库**：
- **Ant Design 5.x** - 企业级UI组件库，适合管理控制台
- **Ant Design Charts** - 数据可视化图表库

**状态管理**：
- **Zustand** - 轻量级状态管理（比Redux简单，适合中等规模应用）
- **React Query (TanStack Query)** - 服务端状态管理和数据缓存

**数据请求**：
- **Axios** - HTTP客户端
- **React Query** - 数据获取和缓存策略

**工具库**：
- **dayjs** - 时间处理
- **lodash-es** - 工具函数
- **ahooks** - React Hooks工具库

### 1.2 项目结构

```
task-monitor-frontend/
├── public/                      # 静态资源
├── src/
│   ├── api/                     # API接口定义
│   │   ├── nodes.ts            # 节点相关API
│   │   ├── jobs.ts             # 作业相关API
│   │   ├── metrics.ts          # 指标相关API
│   │   └── index.ts            # API统一导出
│   │
│   ├── components/              # 通用组件
│   │   ├── Layout/             # 布局组件
│   │   │   ├── MainLayout.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── Header.tsx
│   │   ├── Charts/             # 图表组件
│   │   │   ├── NPUUsageChart.tsx
│   │   │   ├── ProcessMetricsChart.tsx
│   │   │   └── TimeSeriesChart.tsx
│   │   ├── Cards/              # 卡片组件
│   │   │   ├── NodeCard.tsx
│   │   │   ├── JobCard.tsx
│   │   │   └── MetricCard.tsx
│   │   └── Common/             # 通用组件
│   │       ├── StatusBadge.tsx
│   │       ├── LoadingSpinner.tsx
│   │       └── EmptyState.tsx
│   │
│   ├── pages/                   # 页面组件
│   │   ├── Dashboard/          # 总览页
│   │   │   ├── index.tsx
│   │   │   ├── ClusterOverview.tsx
│   │   │   └── RecentJobs.tsx
│   │   ├── Nodes/              # 节点管理
│   │   │   ├── NodeList.tsx
│   │   │   └── NodeDetail.tsx
│   │   ├── Jobs/               # 作业管理
│   │   │   ├── JobList.tsx
│   │   │   └── JobDetail.tsx
│   │   └── Monitoring/         # 监控页面
│   │       ├── NPUMonitoring.tsx
│   │       └── ProcessMonitoring.tsx
│   │
│   ├── stores/                  # 状态管理
│   │   ├── useNodeStore.ts
│   │   ├── useJobStore.ts
│   │   └── useUserStore.ts
│   │
│   ├── hooks/                   # 自定义Hooks
│   │   ├── useNodes.ts
│   │   ├── useJobs.ts
│   │   ├── useMetrics.ts
│   │   └── usePolling.ts
│   │
│   ├── types/                   # TypeScript类型定义
│   │   ├── node.ts
│   │   ├── job.ts
│   │   ├── metrics.ts
│   │   └── api.ts
│   │
│   ├── utils/                   # 工具函数
│   │   ├── format.ts           # 格式化函数
│   │   ├── constants.ts        # 常量定义
│   │   └── helpers.ts          # 辅助函数
│   │
│   ├── styles/                  # 样式文件
│   │   ├── global.css
│   │   └── variables.css
│   │
│   ├── App.tsx                  # 根组件
│   ├── main.tsx                 # 入口文件
│   └── router.tsx               # 路由配置
│
├── package.json
├── tsconfig.json
├── vite.config.ts
└── README.md
```

## 二、页面结构设计

### 2.1 导航结构

```
主导航（侧边栏）
├── 📊 总览 (Dashboard)              - /dashboard
├── 🖥️ 节点管理 (Nodes)              - /nodes
│   └── 节点详情                     - /nodes/:nodeId
├── 📋 作业管理 (Jobs)                - /jobs
│   └── 作业详情                      - /jobs/:jobId
├── 📈 实时监控 (Monitoring)          - /monitoring
│   ├── NPU监控                      - /monitoring/npu
│   └── 进程监控                      - /monitoring/process
└── 📚 历史分析 (History)             - /history
    ├── 作业历史                      - /history/jobs
    └── 状态变更                      - /history/status
```

### 2.2 核心页面设计

#### 2.2.1 总览页 (Dashboard)

**布局结构**：
```
┌─────────────────────────────────────────────────────────┐
│ 顶部统计卡片区                                            │
│ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐    │
│ │总节点数   │ │运行作业   │ │健康NPU   │ │平均负载   │    │
│ │   12     │ │   45     │ │  96/96   │ │  65%     │    │
│ └──────────┘ └──────────┘ └──────────┘ └──────────┘    │
├─────────────────────────────────────────────────────────┤
│ 中部可视化区                                              │
│ ┌────────────────────┐ ┌────────────────────┐          │
│ │ 集群NPU使用率趋势   │ │ 作业类型分布饼图    │          │
│ │ (折线图)           │ │ (饼图)             │          │
│ └────────────────────┘ └────────────────────┘          │
├─────────────────────────────────────────────────────────┤
│ 底部列表区                                                │
│ ┌─────────────────────────────────────────────────┐    │
│ │ 最近运行的作业 (表格)                             │    │
│ │ - 作业名称 | 节点 | 状态 | 框架 | 开始时间        │    │
│ └─────────────────────────────────────────────────┘    │
│ ┌─────────────────────────────────────────────────┐    │
│ │ 节点状态概览 (卡片列表)                           │    │
│ │ [节点卡片1] [节点卡片2] [节点卡片3] ...          │    │
│ └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

**功能特性**：
- 实时显示集群整体状态
- 快速访问最近的作业
- 节点健康状态一览
- 资源使用趋势可视化

#### 2.2.2 节点列表页 (Node List)

**布局结构**：
```
┌─────────────────────────────────────────────────────────┐
│ 筛选和搜索栏                                              │
│ [搜索框] [状态筛选] [NPU数量筛选] [刷新按钮]              │
├─────────────────────────────────────────────────────────┤
│ 节点卡片网格 (Grid Layout)                               │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐    │
│ │ Node-01      │ │ Node-02      │ │ Node-03      │    │
│ │ ● Active     │ │ ● Active     │ │ ○ Inactive   │    │
│ │ IP: 192...   │ │ IP: 192...   │ │ IP: 192...   │    │
│ │ NPU: 8/8     │ │ NPU: 8/8     │ │ NPU: 0/8     │    │
│ │ Jobs: 3      │ │ Jobs: 5      │ │ Jobs: 0      │    │
│ │ Avg: 75%     │ │ Avg: 82%     │ │ Avg: 0%      │    │
│ └──────────────┘ └──────────────┘ └──────────────┘    │
└─────────────────────────────────────────────────────────┘
```

**功能特性**：
- 卡片式展示，直观显示节点状态
- 支持按状态、NPU数量筛选
- 搜索节点名称或IP
- 点击卡片进入节点详情

#### 2.2.3 节点详情页 (Node Detail)

**布局结构**：
```
┌─────────────────────────────────────────────────────────┐
│ 面包屑导航: 节点管理 > Node-01                            │
├─────────────────────────────────────────────────────────┤
│ 节点基本信息卡片                                          │
│ ┌─────────────────────────────────────────────────┐    │
│ │ Node-01 [● Active]  [最后心跳: 2秒前]           │    │
│ │ Hostname: gpu-node-01                            │    │
│ │ IP: 192.168.1.100                                │    │
│ │ NPU数量: 8                                       │    │
│ └─────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────┤
│ Tab切换区                                                │
│ [NPU监控] [运行作业] [历史记录]                          │
│                                                         │
│ Tab 1: NPU监控                                          │
│ ┌────────────────────┐ ┌────────────────────┐          │
│ │ NPU使用率实时图     │ │ NPU温度/功率图      │          │
│ └────────────────────┘ └────────────────────┘          │
│ ┌─────────────────────────────────────────────────┐    │
│ │ NPU设备列表 (表格)                               │    │
│ │ NPU ID | 名称 | 使用率 | 温度 | 功率 | 内存      │    │
│ │ 0      | ...  | 85%   | 65°C | 250W | 16G/32G  │    │
│ └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

**功能特性**：
- 节点基本信息展示
- NPU设备实时监控
- 运行作业列表
- 历史数据查询

#### 2.2.4 作业列表页 (Job List)

**布局结构**：
```
┌─────────────────────────────────────────────────────────┐
│ 高级筛选区                                                │
│ [搜索] [状态] [类型] [框架] [节点] [时间范围] [导出]      │
├─────────────────────────────────────────────────────────┤
│ 作业表格 (支持排序、分页)                                 │
│ ┌─────────────────────────────────────────────────┐    │
│ │ ☑ | 作业名 | 节点 | 类型 | 状态 | 框架 | 时间 | 操作│    │
│ │ ☐ | train_model.py | Node-01 | Training |        │    │
│ │     ● Running | PyTorch | 2h30m | [详情][停止]  │    │
│ │ ☐ | inference.py | Node-02 | Inference |         │    │
│ │     ● Running | Transformers | 1h15m | [详情]    │    │
│ └─────────────────────────────────────────────────┘    │
│ [批量操作] [每页显示: 20] [分页: 1/10]                   │
└─────────────────────────────────────────────────────────┘
```

**功能特性**：
- 多维度筛选和搜索
- 批量操作（停止、删除）
- 导出作业数据
- 实时状态更新

#### 2.2.5 作业详情页 (Job Detail)

**布局结构**：
```
┌─────────────────────────────────────────────────────────┐
│ 面包屑: 作业管理 > train_model.py                        │
├─────────────────────────────────────────────────────────┤
│ 作业头部信息                                              │
│ ┌─────────────────────────────────────────────────┐    │
│ │ train_model.py  [● Running]  [停止] [重启]      │    │
│ │ 作业ID: abc123def456                            │    │
│ │ 节点: Node-01 | 类型: Training | 框架: PyTorch   │    │
│ │ PID: 12345 | 开始时间: 2024-02-05 10:30:00      │    │
│ │ 运行时长: 2小时30分钟                            │    │
│ └─────────────────────────────────────────────────┘    │
├─────────────────────────────────────────────────────────┤
│ Tab切换区                                                │
│ [概览] [参数配置] [代码信息] [资源监控] [状态历史]        │
│                                                         │
│ Tab 1: 概览                                             │
│ ┌──────────────┐ ┌──────────────┐ ┌──────────────┐    │
│ │ CPU使用率     │ │ 内存使用      │ │ NPU使用率     │    │
│ │ 85%          │ │ 4.2GB/16GB   │ │ 90%          │    │
│ └──────────────┘ └──────────────┘ └──────────────┘    │
│ ┌─────────────────────────────────────────────────┐    │
│ │ 命令行                                           │    │
│ │ python train.py --batch_size 32 --lr 0.001 ...  │    │
│ └─────────────────────────────────────────────────┘    │
│                                                         │
│ Tab 2: 参数配置                                         │
│ ┌─────────────────────────────────────────────────┐    │
│ │ 命令行参数 (JSON格式化显示)                       │    │
│ │ {                                                │    │
│ │   "batch_size": 32,                              │    │
│ │   "learning_rate": 0.001,                        │    │
│ │   "epochs": 100                                  │    │
│ │ }                                                │    │
│ └─────────────────────────────────────────────────┘    │
│ ┌─────────────────────────────────────────────────┐    │
│ │ 配置文件 (可展开/折叠)                            │    │
│ │ config.yaml [查看] [下载]                        │    │
│ └─────────────────────────────────────────────────┘    │
│                                                         │
│ Tab 3: 代码信息                                         │
│ ┌─────────────────────────────────────────────────┐    │
│ │ 脚本路径: /workspace/train.py                    │    │
│ │ [查看代码] [下载]                                 │    │
│ └─────────────────────────────────────────────────┘    │
│                                                         │
│ Tab 4: 资源监控                                         │
│ ┌────────────────────┐ ┌────────────────────┐          │
│ │ CPU/内存趋势图      │ │ NPU使用率趋势图     │          │
│ └────────────────────┘ └────────────────────┘          │
│                                                         │
│ Tab 5: 状态历史                                         │
│ ┌─────────────────────────────────────────────────┐    │
│ │ 时间轴展示状态变更                                │    │
│ │ 2024-02-05 10:30:00 - Running (agent_report)    │    │
│ │ 2024-02-05 12:00:00 - Paused (manual)           │    │
│ │ 2024-02-05 12:05:00 - Running (manual)          │    │
│ └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
```

**功能特性**：
- 作业完整信息展示
- 参数和配置文件查看
- 代码内容查看和下载
- 资源使用趋势图表
- 状态变更历史追踪

## 三、数据流和状态管理

### 3.1 数据流架构

```
┌─────────────────────────────────────────────────────────┐
│                    React Components                      │
│  (Dashboard, NodeList, JobDetail, etc.)                 │
└────────────┬────────────────────────────┬───────────────┘
             │                            │
             │ 使用                        │ 使用
             ▼                            ▼
┌────────────────────────┐    ┌────────────────────────┐
│   React Query          │    │   Zustand Store        │
│  (服务端状态)           │    │  (客户端状态)           │
│                        │    │                        │
│ - useNodes()           │    │ - 用户偏好设置          │
│ - useJobs()            │    │ - UI状态(侧边栏展开等)  │
│ - useMetrics()         │    │ - 筛选条件缓存          │
│ - 自动缓存和刷新        │    │ - 临时数据              │
└────────────┬───────────┘    └────────────────────────┘
             │
             │ 调用
             ▼
┌────────────────────────┐
│      API Layer         │
│   (Axios + 拦截器)      │
│                        │
│ - 请求/响应拦截         │
│ - 错误处理              │
│ - Token管理             │
└────────────┬───────────┘
             │
             │ HTTP请求
             ▼
┌────────────────────────┐
│   Backend API Server   │
│   (Go Server)          │
└────────────────────────┘
```

### 3.2 React Query 使用策略

**查询配置**：
```typescript
// hooks/useNodes.ts
export const useNodes = () => {
  return useQuery({
    queryKey: ['nodes'],
    queryFn: fetchNodes,
    staleTime: 30000,        // 30秒内数据视为新鲜
    cacheTime: 300000,       // 缓存5分钟
    refetchInterval: 60000,  // 每60秒自动刷新
    refetchOnWindowFocus: true,
  });
};

// hooks/useJobDetail.ts
export const useJobDetail = (jobId: string) => {
  return useQuery({
    queryKey: ['job', jobId],
    queryFn: () => fetchJobDetail(jobId),
    enabled: !!jobId,        // 只有jobId存在时才查询
    staleTime: 10000,        // 10秒
  });
};
```

**数据更新策略**：
- **节点列表**: 60秒自动刷新
- **作业列表**: 30秒自动刷新
- **作业详情**: 手动刷新 + 窗口聚焦时刷新
- **实时指标**: 10秒自动刷新

### 3.3 Zustand Store 设计

```typescript
// stores/useUserStore.ts
interface UserState {
  theme: 'light' | 'dark';
  sidebarCollapsed: boolean;
  language: 'zh' | 'en';
  setTheme: (theme: 'light' | 'dark') => void;
  toggleSidebar: () => void;
}

// stores/useFilterStore.ts
interface FilterState {
  jobFilters: {
    status?: string[];
    type?: string[];
    framework?: string[];
    nodeId?: string;
    timeRange?: [Date, Date];
  };
  setJobFilters: (filters: Partial<FilterState['jobFilters']>) => void;
  clearJobFilters: () => void;
}
```

## 四、API接口设计

### 4.1 API基础配置

```typescript
// api/client.ts
import axios from 'axios';

const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    // 添加认证token
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => response.data,
  (error) => {
    // 统一错误处理
    if (error.response?.status === 401) {
      // 跳转到登录页
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);
```

### 4.2 API接口定义

**节点相关API**：
```typescript
// api/nodes.ts
export interface NodeListParams {
  status?: 'active' | 'inactive' | 'error';
  page?: number;
  pageSize?: number;
}

export interface NodeListResponse {
  nodes: Node[];
  total: number;
  page: number;
  pageSize: number;
}

// 获取节点列表
export const fetchNodes = async (params?: NodeListParams): Promise<NodeListResponse> => {
  return apiClient.get('/nodes', { params });
};

// 获取节点详情
export const fetchNodeDetail = async (nodeId: string): Promise<Node> => {
  return apiClient.get(`/nodes/${nodeId}`);
};

// 获取节点的NPU指标
export const fetchNodeNPUMetrics = async (
  nodeId: string,
  timeRange?: { start: Date; end: Date }
): Promise<NPUMetric[]> => {
  return apiClient.get(`/nodes/${nodeId}/npu-metrics`, {
    params: {
      start: timeRange?.start?.toISOString(),
      end: timeRange?.end?.toISOString(),
    },
  });
};
```

**作业相关API**：
```typescript
// api/jobs.ts
export interface JobListParams {
  status?: string[];
  type?: string[];
  framework?: string[];
  nodeId?: string;
  startTime?: Date;
  endTime?: Date;
  page?: number;
  pageSize?: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

// 获取作业列表
export const fetchJobs = async (params?: JobListParams): Promise<JobListResponse> => {
  return apiClient.get('/jobs', { params });
};

// 获取作业详情
export const fetchJobDetail = async (jobId: string): Promise<JobDetail> => {
  return apiClient.get(`/jobs/${jobId}`);
};

// 获取作业参数
export const fetchJobParameters = async (jobId: string): Promise<Parameter[]> => {
  return apiClient.get(`/jobs/${jobId}/parameters`);
};

// 获取作业代码
export const fetchJobCode = async (jobId: string): Promise<Code[]> => {
  return apiClient.get(`/jobs/${jobId}/code`);
};

// 获取作业进程指标
export const fetchJobProcessMetrics = async (
  jobId: string,
  timeRange?: { start: Date; end: Date }
): Promise<ProcessMetric[]> => {
  return apiClient.get(`/jobs/${jobId}/process-metrics`, {
    params: {
      start: timeRange?.start?.toISOString(),
      end: timeRange?.end?.toISOString(),
    },
  });
};

// 停止作业
export const stopJob = async (jobId: string): Promise<void> => {
  return apiClient.post(`/jobs/${jobId}/stop`);
};

// 批量停止作业
export const batchStopJobs = async (jobIds: string[]): Promise<void> => {
  return apiClient.post('/jobs/batch-stop', { jobIds });
};
```

**监控指标API**：
```typescript
// api/metrics.ts
// 获取集群整体统计
export const fetchClusterStats = async (): Promise<ClusterStats> => {
  return apiClient.get('/metrics/cluster-stats');
};

// 获取NPU指标
export const fetchNPUMetrics = async (
  nodeId?: string,
  timeRange?: { start: Date; end: Date }
): Promise<NPUMetric[]> => {
  return apiClient.get('/metrics/npu', {
    params: {
      nodeId,
      start: timeRange?.start?.toISOString(),
      end: timeRange?.end?.toISOString(),
    },
  });
};
```

## 五、TypeScript类型定义

### 5.1 核心数据类型

```typescript
// types/node.ts
export interface Node {
  nodeId: string;
  hostId: string | null;
  hostname: string | null;
  ipAddress: string | null;
  npuCount: number | null;
  status: 'active' | 'inactive' | 'error' | null;
  lastHeartbeat: string | null;  // ISO 8601 格式
  createdAt: string;
  updatedAt: string;
}

// types/job.ts
export type JobType = 'training' | 'inference' | 'testing' | 'unknown';
export type JobStatus = 'running' | 'completed' | 'failed' | 'stopped' | 'lost';

export interface Job {
  jobId: string;
  nodeId: string | null;
  hostId: string | null;
  jobName: string | null;
  jobType: JobType | null;
  pid: number | null;
  ppid: number | null;
  pgid: number | null;
  processName: string | null;
  commandLine: string | null;
  framework: string | null;
  modelFormat: string | null;
  status: JobStatus | null;
  startTime: number | null;  // Unix timestamp (ms)
  endTime: number | null;
  cwd: string | null;
  createdAt: string;
  updatedAt: string | null;
}

export interface JobDetail extends Job {
  node?: Node;
  parameters?: Parameter[];
  code?: Code[];
  latestMetrics?: ProcessMetric;
}

// types/parameter.ts
export interface Parameter {
  id: number;
  jobId: string;
  parameterRaw: string | null;
  parameterData: Record<string, any> | null;
  parameterSource: string | null;
  configFilePath: string | null;
  configFileContent: string | null;
  envVars: Record<string, string> | null;
  timestamp: string;
}

// types/code.ts
export interface Code {
  id: number;
  jobId: string;
  scriptPath: string | null;
  scriptContent: string | null;
  importedLibraries: string | null;
  configFiles: string | null;
  shScriptPath: string | null;
  shScriptContent: string | null;
  timestamp: string;
}

// types/metrics.ts
export interface NPUMetric {
  id: number;
  nodeId: string;
  npuId: number;
  name: string | null;
  health: string | null;
  powerW: number | null;
  tempC: number | null;
  aicoreUsagePercent: number | null;
  memoryUsageMb: number | null;
  memoryTotalMb: number | null;
  hbmUsageMb: number | null;
  hbmTotalMb: number | null;
  busId: string | null;
  timestamp: string;
}

export interface ProcessMetric {
  id: number;
  jobId: string;
  pid: number;
  cpuPercent: number | null;
  memoryMb: number | null;
  threadCount: number | null;
  openFiles: number | null;
  status: string | null;
  timestamp: string;
}

export interface ClusterStats {
  totalNodes: number;
  activeNodes: number;
  totalJobs: number;
  runningJobs: number;
  totalNPUs: number;
  healthyNPUs: number;
  avgNPUUsage: number;
  jobTypeDistribution: Record<JobType, number>;
  frameworkDistribution: Record<string, number>;
}
```

### 5.2 API响应类型

```typescript
// types/api.ts
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

export interface ApiError {
  code: number;
  message: string;
  details?: any;
}
```

## 六、UI/UX设计规范

### 6.1 设计系统

**色彩方案**：
```css
/* styles/variables.css */
:root {
  /* 主色调 - 蓝色系 */
  --primary-color: #1890ff;
  --primary-hover: #40a9ff;
  --primary-active: #096dd9;

  /* 状态色 */
  --success-color: #52c41a;    /* 成功/运行中 */
  --warning-color: #faad14;    /* 警告 */
  --error-color: #ff4d4f;      /* 错误/失败 */
  --info-color: #1890ff;       /* 信息 */

  /* 中性色 */
  --text-primary: rgba(0, 0, 0, 0.85);
  --text-secondary: rgba(0, 0, 0, 0.65);
  --text-disabled: rgba(0, 0, 0, 0.25);
  --border-color: #d9d9d9;
  --background-color: #f0f2f5;

  /* 卡片和容器 */
  --card-background: #ffffff;
  --card-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  --card-radius: 8px;

  /* 间距 */
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  --spacing-xl: 32px;
}
```

**状态颜色映射**：
```typescript
// utils/constants.ts
export const STATUS_COLORS = {
  // 节点状态
  node: {
    active: '#52c41a',    // 绿色
    inactive: '#d9d9d9',  // 灰色
    error: '#ff4d4f',     // 红色
  },
  // 作业状态
  job: {
    running: '#52c41a',   // 绿色
    completed: '#1890ff', // 蓝色
    failed: '#ff4d4f',    // 红色
    stopped: '#faad14',   // 橙色
    lost: '#d9d9d9',      // 灰色
  },
  // NPU健康状态
  npu: {
    OK: '#52c41a',
    Warning: '#faad14',
    Error: '#ff4d4f',
  },
};

export const STATUS_LABELS = {
  job: {
    running: '运行中',
    completed: '已完成',
    failed: '失败',
    stopped: '已停止',
    lost: '失联',
  },
  node: {
    active: '在线',
    inactive: '离线',
    error: '错误',
  },
};
```

### 6.2 组件设计规范

**卡片组件**：
```typescript
// components/Cards/NodeCard.tsx
interface NodeCardProps {
  node: Node;
  onClick?: () => void;
}

// 设计要点：
// - 卡片高度固定，宽度响应式
// - 状态指示器使用圆点 + 颜色
// - 关键指标突出显示
// - 悬停效果：阴影加深 + 轻微上移
// - 点击效果：缩放动画
```

**状态徽章**：
```typescript
// components/Common/StatusBadge.tsx
interface StatusBadgeProps {
  status: JobStatus | NodeStatus;
  type: 'job' | 'node';
}

// 设计要点：
// - 使用Ant Design的Badge组件
// - 根据状态显示不同颜色
// - 包含状态文本
// - 支持小尺寸和大尺寸
```

**图表组件**：
```typescript
// components/Charts/NPUUsageChart.tsx
// 设计要点：
// - 使用Ant Design Charts
// - 响应式设计，自适应容器大小
// - 支持时间范围选择
// - 支持多NPU对比
// - 工具提示显示详细数据
// - 支持导出图表数据
```

### 6.3 响应式设计

**断点定义**：
```typescript
// utils/constants.ts
export const BREAKPOINTS = {
  xs: 480,   // 手机
  sm: 576,   // 手机横屏
  md: 768,   // 平板
  lg: 992,   // 桌面
  xl: 1200,  // 大桌面
  xxl: 1600, // 超大桌面
};
```

**布局适配**：
- **节点卡片网格**：
  - xl: 4列
  - lg: 3列
  - md: 2列
  - sm: 1列

- **统计卡片**：
  - xl/lg: 4列
  - md: 2列
  - sm: 1列

## 七、性能优化策略

### 7.1 代码分割和懒加载

```typescript
// router.tsx
import { lazy, Suspense } from 'react';
import { LoadingSpinner } from '@/components/Common/LoadingSpinner';

// 路由级别的代码分割
const Dashboard = lazy(() => import('@/pages/Dashboard'));
const NodeList = lazy(() => import('@/pages/Nodes/NodeList'));
const NodeDetail = lazy(() => import('@/pages/Nodes/NodeDetail'));
const JobList = lazy(() => import('@/pages/Jobs/JobList'));
const JobDetail = lazy(() => import('@/pages/Jobs/JobDetail'));

export const routes = [
  {
    path: '/dashboard',
    element: (
      <Suspense fallback={<LoadingSpinner />}>
        <Dashboard />
      </Suspense>
    ),
  },
  // ... 其他路由
];
```

### 7.2 数据缓存策略

**React Query缓存配置**：
```typescript
// main.tsx
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30000,           // 30秒内数据视为新鲜
      cacheTime: 300000,          // 缓存5分钟
      retry: 2,                   // 失败重试2次
      refetchOnWindowFocus: true, // 窗口聚焦时重新获取
      refetchOnReconnect: true,   // 重新连接时重新获取
    },
  },
});
```

**本地存储缓存**：
```typescript
// utils/cache.ts
// 使用localStorage缓存用户偏好设置
export const cacheManager = {
  set: (key: string, value: any, ttl?: number) => {
    const item = {
      value,
      expiry: ttl ? Date.now() + ttl : null,
    };
    localStorage.setItem(key, JSON.stringify(item));
  },

  get: (key: string) => {
    const itemStr = localStorage.getItem(key);
    if (!itemStr) return null;

    const item = JSON.parse(itemStr);
    if (item.expiry && Date.now() > item.expiry) {
      localStorage.removeItem(key);
      return null;
    }
    return item.value;
  },
};
```

### 7.3 虚拟滚动

对于大数据量列表（如作业列表），使用虚拟滚动：

```typescript
// components/Jobs/JobList.tsx
import { useVirtualizer } from '@tanstack/react-virtual';

// 只渲染可见区域的行，提升性能
const JobList = ({ jobs }: { jobs: Job[] }) => {
  const parentRef = useRef<HTMLDivElement>(null);

  const virtualizer = useVirtualizer({
    count: jobs.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 60, // 每行高度
    overscan: 5,            // 预渲染5行
  });

  // ... 渲染逻辑
};
```

### 7.4 图片和资源优化

```typescript
// vite.config.ts
export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor': ['react', 'react-dom', 'react-router-dom'],
          'antd': ['antd', '@ant-design/icons'],
          'charts': ['@ant-design/charts'],
        },
      },
    },
    chunkSizeWarningLimit: 1000,
  },
  // 图片压缩
  plugins: [
    imagemin({
      gifsicle: { optimizationLevel: 3 },
      mozjpeg: { quality: 80 },
      pngquant: { quality: [0.8, 0.9] },
    }),
  ],
});
```

### 7.5 防抖和节流

```typescript
// hooks/useDebounce.ts
import { useEffect, useState } from 'react';

export const useDebounce = <T,>(value: T, delay: number): T => {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => clearTimeout(handler);
  }, [value, delay]);

  return debouncedValue;
};

// 使用示例：搜索框防抖
const JobList = () => {
  const [searchTerm, setSearchTerm] = useState('');
  const debouncedSearchTerm = useDebounce(searchTerm, 500);

  const { data } = useJobs({ search: debouncedSearchTerm });
  // ...
};
```

## 八、开发规范和最佳实践

### 8.1 代码组织规范

**组件文件结构**：
```typescript
// components/Cards/NodeCard.tsx
import React from 'react';
import { Card, Badge, Typography } from 'antd';
import type { Node } from '@/types/node';
import styles from './NodeCard.module.css';

// 1. 类型定义
interface NodeCardProps {
  node: Node;
  onClick?: () => void;
}

// 2. 常量定义
const STATUS_CONFIG = {
  active: { color: 'success', text: '在线' },
  inactive: { color: 'default', text: '离线' },
  error: { color: 'error', text: '错误' },
};

// 3. 组件定义
export const NodeCard: React.FC<NodeCardProps> = ({ node, onClick }) => {
  // 4. Hooks
  const [isHovered, setIsHovered] = useState(false);

  // 5. 事件处理函数
  const handleClick = () => {
    onClick?.();
  };

  // 6. 渲染逻辑
  return (
    <Card
      className={styles.nodeCard}
      onClick={handleClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      {/* 组件内容 */}
    </Card>
  );
};

// 7. 默认导出（如果需要）
export default NodeCard;
```

### 8.2 命名规范

**文件命名**：
- 组件文件：PascalCase（如 `NodeCard.tsx`）
- 工具函数：camelCase（如 `formatDate.ts`）
- 类型定义：camelCase（如 `node.ts`）
- 样式文件：kebab-case（如 `node-card.module.css`）

**变量命名**：
- 组件：PascalCase（如 `NodeCard`）
- 函数：camelCase（如 `fetchNodes`）
- 常量：UPPER_SNAKE_CASE（如 `API_BASE_URL`）
- 类型/接口：PascalCase（如 `NodeCardProps`）

### 8.3 TypeScript最佳实践

```typescript
// ✅ 好的实践
interface User {
  id: string;
  name: string;
  email?: string; // 可选属性使用 ?
}

const fetchUser = async (id: string): Promise<User> => {
  // 明确的返回类型
};

// ❌ 避免使用 any
const data: any = await fetchData(); // 不推荐

// ✅ 使用具体类型
const data: User = await fetchData();

// ✅ 使用类型守卫
const isNode = (obj: any): obj is Node => {
  return 'nodeId' in obj && 'hostname' in obj;
};
```

### 8.4 错误处理

```typescript
// hooks/useJobs.ts
export const useJobs = (params?: JobListParams) => {
  return useQuery({
    queryKey: ['jobs', params],
    queryFn: () => fetchJobs(params),
    onError: (error: ApiError) => {
      // 统一错误处理
      message.error(`获取作业列表失败: ${error.message}`);
      console.error('Failed to fetch jobs:', error);
    },
  });
};

// 组件中的错误处理
const JobList = () => {
  const { data, error, isLoading } = useJobs();

  if (error) {
    return <ErrorState message="加载失败，请重试" onRetry={() => refetch()} />;
  }

  if (isLoading) {
    return <LoadingSpinner />;
  }

  return <JobTable jobs={data.items} />;
};
```

### 8.5 测试规范

```typescript
// components/Cards/NodeCard.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { NodeCard } from './NodeCard';

describe('NodeCard', () => {
  const mockNode: Node = {
    nodeId: 'node-1',
    hostname: 'test-node',
    status: 'active',
    // ... 其他属性
  };

  it('renders node information correctly', () => {
    render(<NodeCard node={mockNode} />);
    expect(screen.getByText('test-node')).toBeInTheDocument();
  });

  it('calls onClick when clicked', () => {
    const handleClick = jest.fn();
    render(<NodeCard node={mockNode} onClick={handleClick} />);

    fireEvent.click(screen.getByRole('button'));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });
});
```

## 九、部署和构建

### 9.1 环境变量配置

```bash
# .env.development
VITE_API_BASE_URL=http://localhost:8080/api
VITE_APP_TITLE=NPU作业监控系统（开发环境）

# .env.production
VITE_API_BASE_URL=https://api.example.com/api
VITE_APP_TITLE=NPU作业监控系统
```

### 9.2 构建配置

```typescript
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    outDir: 'dist',
    sourcemap: false,
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true, // 生产环境移除console
      },
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
```

### 9.3 Docker部署

```dockerfile
# Dockerfile
FROM node:18-alpine as builder

WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

```nginx
# nginx.conf
server {
    listen 80;
    server_name localhost;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://backend:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 十、开发流程和工具

### 10.1 开发环境搭建

```bash
# 1. 克隆项目
git clone <repository-url>
cd task-monitor-frontend

# 2. 安装依赖
npm install

# 3. 启动开发服务器
npm run dev

# 4. 构建生产版本
npm run build

# 5. 预览生产构建
npm run preview
```

### 10.2 推荐的VSCode插件

```json
// .vscode/extensions.json
{
  "recommendations": [
    "dbaeumer.vscode-eslint",
    "esbenp.prettier-vscode",
    "bradlc.vscode-tailwindcss",
    "dsznajder.es7-react-js-snippets",
    "formulahendry.auto-rename-tag",
    "christian-kohler.path-intellisense"
  ]
}
```

### 10.3 代码质量工具

**ESLint配置**：
```json
// .eslintrc.json
{
  "extends": [
    "eslint:recommended",
    "plugin:react/recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:react-hooks/recommended"
  ],
  "rules": {
    "react/react-in-jsx-scope": "off",
    "@typescript-eslint/no-explicit-any": "warn",
    "no-console": ["warn", { "allow": ["warn", "error"] }]
  }
}
```

**Prettier配置**：
```json
// .prettierrc
{
  "semi": true,
  "trailingComma": "es5",
  "singleQuote": true,
  "printWidth": 100,
  "tabWidth": 2
}
```

### 10.4 Git工作流

```bash
# 功能开发流程
git checkout -b feature/node-detail-page
# 开发功能...
git add .
git commit -m "feat: 添加节点详情页面"
git push origin feature/node-detail-page
# 创建Pull Request

# 提交信息规范
# feat: 新功能
# fix: 修复bug
# docs: 文档更新
# style: 代码格式调整
# refactor: 重构
# test: 测试相关
# chore: 构建/工具相关
```

## 十一、项目依赖清单

### 11.1 核心依赖

```json
{
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.0",
    "antd": "^5.12.0",
    "@ant-design/icons": "^5.2.6",
    "@ant-design/charts": "^2.0.0",
    "@tanstack/react-query": "^5.14.0",
    "zustand": "^4.4.7",
    "axios": "^1.6.2",
    "dayjs": "^1.11.10",
    "lodash-es": "^4.17.21",
    "ahooks": "^3.7.8"
  },
  "devDependencies": {
    "@types/react": "^18.2.43",
    "@types/react-dom": "^18.2.17",
    "@types/lodash-es": "^4.17.12",
    "@typescript-eslint/eslint-plugin": "^6.13.2",
    "@typescript-eslint/parser": "^6.13.2",
    "@vitejs/plugin-react": "^4.2.1",
    "eslint": "^8.55.0",
    "eslint-plugin-react": "^7.33.2",
    "eslint-plugin-react-hooks": "^4.6.0",
    "prettier": "^3.1.0",
    "typescript": "^5.3.3",
    "vite": "^5.0.7"
  }
}
```

## 十二、总结和下一步计划

### 12.1 设计亮点

1. **清晰的架构分层**：组件、页面、API、状态管理分离明确
2. **类型安全**：全面使用TypeScript，减少运行时错误
3. **性能优化**：代码分割、虚拟滚动、数据缓存等多重优化
4. **用户体验**：响应式设计、状态反馈、错误处理完善
5. **可维护性**：统一的代码规范、清晰的文件结构

### 12.2 开发优先级

**第一阶段（核心功能）**：
1. ✅ 搭建项目基础架构
2. ✅ 实现主布局和路由
3. ✅ 开发总览页（Dashboard）
4. ✅ 开发节点列表和详情页
5. ✅ 开发作业列表和详情页

**第二阶段（增强功能）**：
1. 实现实时监控页面
2. 添加历史数据分析
3. 实现数据导出功能
4. 添加用户偏好设置

**第三阶段（优化完善）**：
1. 性能优化和测试
2. 添加单元测试和E2E测试
3. 完善错误处理和边界情况
4. 文档完善和部署

### 12.3 技术债务和改进方向

1. **国际化支持**：添加i18n支持多语言
2. **主题切换**：支持亮色/暗色主题
3. **权限管理**：添加用户角色和权限控制
4. **实时通知**：WebSocket实时推送重要事件
5. **数据可视化增强**：更丰富的图表类型和交互

### 12.4 预期效果

完成后的前端应用将具备：
- 🎨 **美观的界面**：现代化的设计风格，符合企业级应用标准
- 🚀 **流畅的体验**：快速响应，加载时间<2秒
- 📊 **直观的数据展示**：清晰的图表和统计信息
- 🔍 **强大的查询能力**：多维度筛选和搜索
- 📱 **响应式设计**：适配不同屏幕尺寸
- 🛡️ **稳定可靠**：完善的错误处理和边界情况处理

---

## 附录

### A. 参考资料

- [React官方文档](https://react.dev/)
- [TypeScript官方文档](https://www.typescriptlang.org/)
- [Ant Design组件库](https://ant.design/)
- [React Query文档](https://tanstack.com/query/latest)
- [Zustand状态管理](https://github.com/pmndrs/zustand)

### B. 相关文档

- [DATABASE.md](task_monitor_go/DATABASE.md) - 数据库设计文档
- [README.md](README.md) - 项目说明文档
- [DESIGN.md](DESIGN.md) - 系统架构设计文档

---

**文档版本**: v1.0.0
**最后更新**: 2024-02-05
**维护者**: Task Monitor Frontend Team

