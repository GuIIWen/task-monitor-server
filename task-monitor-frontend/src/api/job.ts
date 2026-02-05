import apiClient from './client';
import type { JobDetail, JobListParams, JobListResponse, JobStats } from '@/types/job';

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
};
