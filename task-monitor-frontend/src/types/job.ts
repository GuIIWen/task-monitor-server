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

export interface NPUMetricInfo {
  npuId: number | null;
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
}

export interface NPUCardInfo {
  npuId: number;
  memoryUsageMb: number;
  metrics: NPUMetricInfo[];
}

export interface JobDetailResponse {
  job: Job;
  npuCards: NPUCardInfo[];
  relatedJobs: Job[];
}

export interface Parameter {
  id: number;
  jobId: string | null;
  parameterRaw: string | null;
  parameterData: string | null;
  parameterSource: string | null;
  configFilePath: string | null;
  configFileContent: string | null;
  envVars: string | null;
  timestamp: string;
}

export interface JobDetail extends Job {
  node?: any;
  parameters?: any[];
  code?: any[];
  latestMetrics?: any;
}

export interface JobListParams {
  status?: string[];
  type?: string[];
  framework?: string[];
  nodeId?: string;
  cardCount?: (number | string)[];
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
  cardCount: number | null;
}

export interface JobAnalysisTaskType {
  category: 'training' | 'inference' | 'unknown';
  subCategory: 'pre-training' | 'fine-tuning' | 'rlhf' | 'evaluation' | 'serving' | 'batch-inference' | null;
  inferenceFramework: string | null;
  evidence: string | null;
}

export interface JobAnalysisModelInfo {
  modelName: string | null;
  modelSize: string | null;
  precision: string | null;
  parallelStrategy: string | null;
}

export interface JobAnalysisRuntimeAnalysis {
  duration: string;
  status: 'normal' | 'long-running' | 'just-started' | 'completed';
  description: string;
}

export interface JobAnalysisParameterItem {
  parameter: string;
  value: string;
  assessment: 'normal' | 'warning' | 'abnormal';
  reason: string;
}

export interface JobAnalysisParameterCheck {
  status: 'normal' | 'warning' | 'abnormal';
  items: JobAnalysisParameterItem[];
}

export interface JobAnalysisResourceAssessment {
  npuUtilization: 'high' | 'medium' | 'low' | 'idle';
  hbmUtilization: 'high' | 'medium' | 'low';
  description: string;
}

export interface JobAnalysisIssue {
  severity: 'critical' | 'warning' | 'info';
  category: string;
  description: string;
  suggestion: string;
}

export interface JobAnalysis {
  summary: string;
  taskType: JobAnalysisTaskType;
  modelInfo: JobAnalysisModelInfo | null;
  runtimeAnalysis: JobAnalysisRuntimeAnalysis | null;
  parameterCheck: JobAnalysisParameterCheck | null;
  resourceAssessment: JobAnalysisResourceAssessment;
  issues: JobAnalysisIssue[];
}

export interface JobAnalysisWithStatus {
  status: 'analyzing' | 'completed' | 'failed';
  result: JobAnalysis | null;
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
