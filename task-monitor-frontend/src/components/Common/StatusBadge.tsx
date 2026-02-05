import React from 'react';
import { Tag } from 'antd';
import type { JobStatus, NodeStatus } from '@/types';

interface StatusBadgeProps {
  status: JobStatus | NodeStatus | null;
  type?: 'job' | 'node';
}

const StatusBadge: React.FC<StatusBadgeProps> = ({ status, type = 'job' }) => {
  if (!status) return <Tag>未知</Tag>;

  const getColor = () => {
    if (type === 'job') {
      switch (status as JobStatus) {
        case 'running':
          return 'processing';
        case 'completed':
          return 'success';
        case 'failed':
          return 'error';
        case 'stopped':
          return 'default';
        case 'lost':
          return 'warning';
        default:
          return 'default';
      }
    } else {
      switch (status as NodeStatus) {
        case 'active':
          return 'success';
        case 'inactive':
          return 'default';
        case 'error':
          return 'error';
        default:
          return 'default';
      }
    }
  };

  const getText = () => {
    if (type === 'job') {
      const jobStatusMap: Record<JobStatus, string> = {
        running: '运行中',
        completed: '已完成',
        failed: '失败',
        stopped: '已停止',
        lost: '丢失',
      };
      return jobStatusMap[status as JobStatus] || status;
    } else {
      const nodeStatusMap: Record<NodeStatus, string> = {
        active: '活跃',
        inactive: '离线',
        error: '错误',
      };
      return nodeStatusMap[status as NodeStatus] || status;
    }
  };

  return <Tag color={getColor()}>{getText()}</Tag>;
};

export default StatusBadge;
