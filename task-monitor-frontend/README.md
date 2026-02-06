# Task Monitor Frontend

Task Monitor 系统的前端应用，用于监控和管理分布式任务执行情况。

## 技术栈

- **React 18** - UI 框架
- **TypeScript 5.x** - 类型安全
- **Vite** - 构建工具
- **Ant Design 5.x** - UI 组件库
- **React Router v6** - 路由管理
- **TanStack Query (React Query)** - 服务端状态管理
- **Zustand** - 客户端状态管理
- **Axios** - HTTP 客户端
- **@ant-design/charts** - 数据可视化
- **dayjs** - 时间处理

## 项目结构

```
src/
├── api/              # API 接口层
│   ├── client.ts     # Axios 客户端配置
│   ├── node.ts       # 节点相关 API
│   ├── job.ts        # 作业相关 API
│   └── metrics.ts    # 指标相关 API
├── components/       # 组件
│   ├── Layout/       # 布局组件
│   ├── Cards/        # 卡片组件
│   ├── Charts/       # 图表组件
│   └── Common/       # 通用组件
├── pages/            # 页面组件
│   ├── Dashboard.tsx # 仪表盘
│   ├── NodeList.tsx  # 节点列表
│   ├── NodeDetail.tsx# 节点详情
│   ├── JobList.tsx   # 作业列表
│   └── JobDetail.tsx # 作业详情
├── hooks/            # 自定义 Hooks
│   ├── useNodes.ts   # 节点数据 Hooks
│   ├── useJobs.ts    # 作业数据 Hooks
│   └── useMetrics.ts # 指标数据 Hooks
├── stores/           # Zustand 状态管理
│   └── useFilterStore.ts # 筛选状态
├── types/            # TypeScript 类型定义
│   ├── api.ts        # API 响应类型
│   ├── node.ts       # 节点类型
│   └── job.ts        # 作业类型
├── utils/            # 工具函数
│   ├── format.ts     # 格式化函数
│   └── constants.ts  # 常量定义
├── styles/           # 样式文件
│   └── global.css    # 全局样式
├── router.tsx        # 路由配置
├── App.tsx           # 应用入口
└── main.tsx          # 主入口文件
```

## 安装依赖

```bash
npm install
```

## 开发

启动开发服务器：

```bash
npm run dev
```

开发服务器将在 `http://localhost:3000` 启动（如果端口被占用会自动使用其他端口）。

## 构建

构建生产版本：

```bash
npm run build
```

构建产物将输出到 `dist/` 目录。

## 预览

预览生产构建：

```bash
npm run preview
```

## API 配置

### 开发环境

开发环境下，Vite 会将 `/api` 请求代理到 `http://localhost:8888`（API Server）。

配置位于 `vite.config.ts`：

```typescript
server: {
  port: 3000,
  proxy: {
    '/api': {
      target: 'http://localhost:8888',
      changeOrigin: true,
    },
  },
}
```

### 生产环境

生产环境下，可以通过环境变量 `VITE_API_BASE_URL` 配置 API 地址：

```bash
VITE_API_BASE_URL=http://your-api-server:8888/api/v1 npm run build
```

## 主要功能

### 1. 仪表盘
- 显示系统总体统计信息
- 节点状态概览
- 作业运行状态概览

### 2. 节点管理
- 节点列表查看（包含节点ID、主机名、IP地址、NPU数量、卡型号等信息）
- 节点详情查看
- 节点状态监控

### 3. 作业管理
- 作业列表查看
- 作业详情查看
- 作业状态监控
- 多维度筛选和排序功能：
  - 作业名称排序（字母顺序）
  - 类型筛选和排序（训练、推理、测试、未知）
  - 框架筛选和排序（PyTorch、TensorFlow、MindSpore、其他）
  - 节点排序
  - 状态筛选和排序（运行中、已完成、失败、已停止、丢失）
  - 开始时间排序

## 开发说明

### 添加新页面

1. 在 `src/pages/` 创建新页面组件
2. 在 `src/router.tsx` 添加路由配置
3. 在 `src/components/Layout/Sidebar.tsx` 添加导航菜单项

### 添加新 API

1. 在 `src/types/` 定义类型
2. 在 `src/api/` 创建 API 函数
3. 在 `src/hooks/` 创建 React Query Hook

### 状态管理

- **服务端状态**：使用 React Query 管理，自动处理缓存、重新获取等
- **客户端状态**：使用 Zustand 管理，如筛选条件、UI 状态等

## 代码规范

- 使用 TypeScript 严格模式
- 遵循 ESLint 规则
- 组件使用函数式组件 + Hooks
- 使用 `import type` 导入类型

## 浏览器支持

支持所有现代浏览器（Chrome、Firefox、Safari、Edge 最新版本）。
