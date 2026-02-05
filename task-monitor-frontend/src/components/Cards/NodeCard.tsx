import React from 'react';
import { Card, Space, Typography } from 'antd';
import { ClusterOutlined } from '@ant-design/icons';
import { StatusBadge } from '@/components/Common';
import { formatISOTime } from '@/utils';
import type { Node } from '@/types/node';

const { Text } = Typography;

interface NodeCardProps {
  node: Node;
  onClick?: () => void;
}

const NodeCard: React.FC<NodeCardProps> = ({ node, onClick }) => {
  return (
    <Card
      hoverable
      onClick={onClick}
      style={{ cursor: onClick ? 'pointer' : 'default' }}
    >
      <Space direction="vertical" style={{ width: '100%' }}>
        <Space>
          <ClusterOutlined style={{ fontSize: '20px' }} />
          <Text strong>{node.hostname || node.nodeId}</Text>
          <StatusBadge status={node.status} type="node" />
        </Space>
        <Space direction="vertical" size="small">
          <Text type="secondary">IP: {node.ipAddress || '-'}</Text>
          <Text type="secondary">NPU数量: {node.npuCount || 0}</Text>
          <Text type="secondary">
            最后心跳: {formatISOTime(node.lastHeartbeat)}
          </Text>
        </Space>
      </Space>
    </Card>
  );
};

export default NodeCard;
