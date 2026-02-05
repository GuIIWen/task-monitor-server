import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Space } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { StatusBadge, LoadingSpinner } from '@/components/Common';
import { useNode } from '@/hooks';
import { formatISOTime } from '@/utils';

const NodeDetail: React.FC = () => {
  const { nodeId } = useParams<{ nodeId: string }>();
  const navigate = useNavigate();
  const { data: node, isLoading } = useNode(nodeId!);

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (!node) {
    return <Card>节点不存在</Card>;
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Button
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/nodes')}
      >
        返回列表
      </Button>

      <Card title="节点详情">
        <Descriptions bordered column={2}>
          <Descriptions.Item label="节点ID">{node.nodeId}</Descriptions.Item>
          <Descriptions.Item label="主机ID">{node.hostId || '-'}</Descriptions.Item>
          <Descriptions.Item label="主机名">{node.hostname || '-'}</Descriptions.Item>
          <Descriptions.Item label="IP地址">{node.ipAddress || '-'}</Descriptions.Item>
          <Descriptions.Item label="NPU数量">{node.npuCount || 0}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <StatusBadge status={node.status} type="node" />
          </Descriptions.Item>
          <Descriptions.Item label="最后心跳">
            {formatISOTime(node.lastHeartbeat)}
          </Descriptions.Item>
          <Descriptions.Item label="创建时间">
            {formatISOTime(node.createdAt)}
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </Space>
  );
};

export default NodeDetail;
