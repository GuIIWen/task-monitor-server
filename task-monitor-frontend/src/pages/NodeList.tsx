import React from 'react';
import { Table, Space, Button } from 'antd';
import { useNavigate } from 'react-router-dom';
import { StatusBadge } from '@/components/Common';
import { useNodes } from '@/hooks';
import { formatISOTime } from '@/utils';
import type { Node } from '@/types/node';

const NodeList: React.FC = () => {
  const navigate = useNavigate();
  const { data, isLoading } = useNodes();

  const columns = [
    {
      title: '节点ID',
      dataIndex: 'nodeId',
      key: 'nodeId',
    },
    {
      title: '主机名',
      dataIndex: 'hostname',
      key: 'hostname',
      render: (text: string) => text || '-',
    },
    {
      title: 'IP地址',
      dataIndex: 'ipAddress',
      key: 'ipAddress',
      render: (text: string) => text || '-',
    },
    {
      title: 'NPU数量',
      dataIndex: 'npuCount',
      key: 'npuCount',
      render: (count: number) => count || 0,
    },
    {
      title: '卡型号',
      dataIndex: 'npuModel',
      key: 'npuModel',
      render: (text: string) => text || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: Node['status']) => (
        <StatusBadge status={status} type="node" />
      ),
    },
    {
      title: '最后心跳',
      dataIndex: 'lastHeartbeat',
      key: 'lastHeartbeat',
      render: (time: string) => formatISOTime(time),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Node) => (
        <Space>
          <Button
            type="link"
            onClick={() => navigate(`/nodes/${record.nodeId}`)}
          >
            查看详情
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <Table
      columns={columns}
      dataSource={data || []}
      loading={isLoading}
      rowKey="nodeId"
      pagination={false}
    />
  );
};

export default NodeList;
