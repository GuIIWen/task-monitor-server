import React from 'react';
import { Table, Space, Button } from 'antd';
import type { FilterValue } from 'antd/es/table/interface';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { StatusBadge } from '@/components/Common';
import { useNodes } from '@/hooks';
import { formatISOTime } from '@/utils';
import type { Node, NodeStatus } from '@/types/node';

const NodeList: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const statusFilter = searchParams.get('status') as NodeStatus | null;
  const { data, isLoading } = useNodes(statusFilter ? { status: statusFilter } : undefined);

  const handleFilterChange = (
    _pagination: any,
    filters: Record<string, FilterValue | null>,
  ) => {
    const newStatus = filters.status?.[0] as string | undefined;
    if (newStatus) {
      setSearchParams({ status: newStatus });
    } else {
      setSearchParams({});
    }
  };

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
      filters: [
        { text: '活跃', value: 'active' },
        { text: '离线', value: 'inactive' },
        { text: '错误', value: 'error' },
      ],
      filteredValue: statusFilter ? [statusFilter] : null,
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
      onChange={handleFilterChange}
      pagination={false}
    />
  );
};

export default NodeList;
