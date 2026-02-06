// 作业类型定义
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
  node?: any;  // 临时使用any，后续会引入Node类型
  parameters?: any[];
  code?: any[];
  latestMetrics?: any;
}

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

export interface JobListResponse {
  jobs: Job[];
  total: number;
  page: number;
  pageSize: number;
}

export interface JobStats {
  total: number;
  running: number;
  completed: number;
  failed: number;
  stopped: number;
  lost: number;
}

export interface JobGroup {
  mainJob: Job;
  childJobs: Job[];
  cardCount: number;
}

export interface GroupedJobListResponse {
  items: JobGroup[];
  pagination: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
}
