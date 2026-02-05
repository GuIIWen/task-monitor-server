import { useQuery } from '@tanstack/react-query';
import { metricsApi } from '@/api';
import { REFRESH_INTERVAL } from '@/utils';

/**
 * 获取作业的最新指标
 */
export const useLatestMetrics = (jobId: string) => {
  return useQuery({
    queryKey: ['metrics', 'latest', jobId],
    queryFn: () => metricsApi.getLatestMetrics(jobId),
    enabled: !!jobId,
    refetchInterval: REFRESH_INTERVAL,
  });
};

/**
 * 获取作业的历史指标
 */
export const useMetricsHistory = (
  jobId: string,
  params?: {
    startTime?: number;
    endTime?: number;
    limit?: number;
  }
) => {
  return useQuery({
    queryKey: ['metrics', 'history', jobId, params],
    queryFn: () => metricsApi.getMetricsHistory(jobId, params),
    enabled: !!jobId,
  });
};
