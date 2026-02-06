import { useQuery } from '@tanstack/react-query';
import { jobApi } from '@/api';
import type { JobListParams } from '@/types/job';
import { REFRESH_INTERVAL } from '@/utils';

/**
 * 获取作业列表
 */
export const useJobs = (params?: JobListParams) => {
  return useQuery({
    queryKey: ['jobs', params],
    queryFn: () => jobApi.getJobs(params),
    refetchInterval: REFRESH_INTERVAL,
  });
};

/**
 * 获取作业详情
 */
export const useJob = (jobId: string) => {
  return useQuery({
    queryKey: ['job', jobId],
    queryFn: () => jobApi.getJobById(jobId),
    enabled: !!jobId,
    refetchInterval: REFRESH_INTERVAL,
  });
};

/**
 * 获取作业统计信息
 */
export const useJobStats = () => {
  return useQuery({
    queryKey: ['jobStats'],
    queryFn: () => jobApi.getJobStats(),
    refetchInterval: REFRESH_INTERVAL,
  });
};

/**
 * 获取作业代码
 */
export const useJobCode = (jobId: string) => {
  return useQuery({
    queryKey: ['jobCode', jobId],
    queryFn: () => jobApi.getJobCode(jobId),
    enabled: !!jobId,
  });
};

/**
 * 获取分组作业列表
 */
export const useGroupedJobs = (params?: JobListParams) => {
  return useQuery({
    queryKey: ['groupedJobs', params],
    queryFn: () => jobApi.getGroupedJobs(params),
    refetchInterval: REFRESH_INTERVAL,
  });
};
