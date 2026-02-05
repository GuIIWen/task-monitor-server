// 节点类型定义
export interface Node {
  nodeId: string;
  hostId: string | null;
  hostname: string | null;
  ipAddress: string | null;
  npuCount: number | null;
  status: 'active' | 'inactive' | 'error' | null;
  lastHeartbeat: string | null;  // ISO 8601 格式
  createdAt: string;
  updatedAt: string;
}

export type NodeStatus = 'active' | 'inactive' | 'error';

export interface NodeListParams {
  status?: NodeStatus;
  page?: number;
  pageSize?: number;
}

export interface NodeListResponse {
  nodes: Node[];
  total: number;
  page: number;
  pageSize: number;
}

export interface NodeStats {
  total: number;
  active: number;
  inactive: number;
  error: number;
}
