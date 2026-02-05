import apiClient from './client';

/**
 * 指标相关API
 */
export const metricsApi = {
  /**
   * 获取作业的最新指标
   */
  getLatestMetrics: async (jobId: string) => {
    return apiClient.get(`/jobs/${jobId}/metrics/latest`);
  },

  /**
   * 获取作业的历史指标
   */
  getMetricsHistory: async (jobId: string, params?: {
    startTime?: number;
    endTime?: number;
    limit?: number;
  }) => {
    return apiClient.get(`/jobs/${jobId}/metrics`, { params });
  },
};
