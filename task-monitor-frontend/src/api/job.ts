import apiClient from './client';
import type { JobDetail, JobListParams, JobListResponse, JobStats, GroupedJobListResponse } from '@/types/job';

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
  getJobById: async (jobId: string): Promise<JobDetail> => {
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
};
