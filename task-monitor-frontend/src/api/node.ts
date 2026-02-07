import apiClient from './client';
import type { Node, NodeListParams, NodeStats } from '@/types/node';

/**
 * 节点相关API
 */
export const nodeApi = {
  /**
   * 获取节点列表
   */
  getNodes: async (params?: NodeListParams): Promise<Node[]> => {
    return apiClient.get('/nodes', { params });
  },

  /**
   * 获取节点详情
   */
  getNodeById: async (nodeId: string): Promise<Node> => {
    return apiClient.get(`/nodes/${nodeId}`);
  },

  /**
   * 获取节点统计信息
   */
  getNodeStats: async (): Promise<NodeStats> => {
    return apiClient.get('/nodes/stats');
  },
};
