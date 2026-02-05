import { useQuery } from '@tanstack/react-query';
import { nodeApi } from '@/api';
import type { NodeListParams } from '@/types/node';
import { REFRESH_INTERVAL } from '@/utils';

/**
 * 获取节点列表
 */
export const useNodes = (params?: NodeListParams) => {
  return useQuery({
    queryKey: ['nodes', params],
    queryFn: () => nodeApi.getNodes(params),
    refetchInterval: REFRESH_INTERVAL,
  });
};

/**
 * 获取节点详情
 */
export const useNode = (nodeId: string) => {
  return useQuery({
    queryKey: ['node', nodeId],
    queryFn: () => nodeApi.getNodeById(nodeId),
    enabled: !!nodeId,
    refetchInterval: REFRESH_INTERVAL,
  });
};

/**
 * 获取节点统计信息
 */
export const useNodeStats = () => {
  return useQuery({
    queryKey: ['nodeStats'],
    queryFn: () => nodeApi.getNodeStats(),
    refetchInterval: REFRESH_INTERVAL,
  });
};
