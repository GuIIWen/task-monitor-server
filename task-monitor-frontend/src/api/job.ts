import apiClient from './client';
import type { JobDetailResponse, JobListParams, JobListResponse, JobStats, GroupedJobListResponse, Parameter, JobAnalysis } from '@/types/job';

/**
 * 作业相关API
 */
export const jobApi = {
  /**
   * 获取作业列表
   */
  getJobs: async (params?: JobListParams): Promise<JobListResponse> => {
    return apiClient.get('/jobs', { params });
  },

  /**
   * 获取作业详情
   */
  getJobById: async (jobId: string): Promise<JobDetailResponse> => {
    return apiClient.get(`/jobs/${jobId}`);
  },

  /**
   * 获取作业统计信息
   */
  getJobStats: async (): Promise<JobStats> => {
    return apiClient.get('/jobs/stats');
  },

  /**
   * 获取作业代码
   */
  getJobCode: async (jobId: string): Promise<any[]> => {
    return apiClient.get(`/jobs/${jobId}/code`);
  },

  /**
   * 获取作业参数（含环境变量）
   */
  getJobParameters: async (jobId: string): Promise<Parameter[]> => {
    return apiClient.get(`/jobs/${jobId}/parameters`);
  },

  /**
   * 获取分组作业列表（按 node_id+pgid 分组）
   */
  getGroupedJobs: async (params?: JobListParams): Promise<GroupedJobListResponse> => {
    return apiClient.get('/jobs/grouped', { params });
  },

  /**
   * 获取所有去重的卡数值
   */
  getDistinctCardCounts: async (): Promise<number[]> => {
    return apiClient.get('/jobs/grouped/card-counts');
  },

  /**
   * AI分析作业
   */
  analyzeJob: async (jobId: string): Promise<JobAnalysis> => {
    return apiClient.post(`/jobs/${jobId}/analyze`, null, { timeout: 180000 });
  },

  /**
   * 获取已保存的AI分析结果
   */
  getJobAnalysis: async (jobId: string): Promise<JobAnalysis | null> => {
    return apiClient.get(`/jobs/${jobId}/analysis`);
  },

  batchAnalyze: async (jobIds: string[]): Promise<{ batchId: string }> => {
    return apiClient.post('/jobs/batch-analyze', { jobIds });
  },

  getBatchAnalyzeProgress: async (batchId: string): Promise<{
    status: string; total: number; current: number; success: number; failed: number;
  }> => {
    return apiClient.get(`/jobs/batch-analyze/${batchId}`);
  },

  /**
   * 批量获取分析摘要
   */
  getBatchAnalyses: async (jobIds: string[]): Promise<Record<string, JobAnalysis>> => {
    return apiClient.get('/jobs/analyses/batch', { params: { jobIds } });
  },
};
