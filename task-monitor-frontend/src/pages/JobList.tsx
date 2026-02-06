import React from 'react';
import { Table, Space, Button, Tag } from 'antd';
import { useNavigate } from 'react-router-dom';
import { StatusBadge } from '@/components/Common';
import { useJobs } from '@/hooks';
import { formatTimestamp, JOB_TYPE_MAP } from '@/utils';
import type { Job } from '@/types/job';

const JobList: React.FC = () => {
  const navigate = useNavigate();
  const { data, isLoading } = useJobs();

  const columns = [
    {
      title: '作业名称',
      dataIndex: 'jobName',
      key: 'jobName',
      render: (text: string, record: Job) => text || record.jobId,
    },
    {
      title: '类型',
      dataIndex: 'jobType',
      key: 'jobType',
      render: (type: Job['jobType']) => (
        type ? <Tag>{JOB_TYPE_MAP[type] || type}</Tag> : '-'
      ),
    },
    {
      title: '框架',
      dataIndex: 'framework',
      key: 'framework',
      render: (text: string) => text ? <Tag color="blue">{text}</Tag> : '-',
    },
    {
      title: '节点',
      dataIndex: 'nodeId',
      key: 'nodeId',
      render: (text: string) => text || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      filters: [
        { text: '运行中', value: 'running' },
        { text: '已完成', value: 'completed' },
        { text: '失败', value: 'failed' },
        { text: '已停止', value: 'stopped' },
        { text: '丢失', value: 'lost' },
      ],
      onFilter: (value: string | number | boolean, record: Job) => {
        return record.status === value;
      },
      sorter: (a: Job, b: Job) => {
        const statusOrder: Record<string, number> = {
          running: 1,
          failed: 2,
          stopped: 3,
          completed: 4,
          lost: 5,
        };
        const aOrder = a.status ? statusOrder[a.status] || 999 : 999;
        const bOrder = b.status ? statusOrder[b.status] || 999 : 999;
        return aOrder - bOrder;
      },
      render: (status: Job['status']) => (
        <StatusBadge status={status} type="job" />
      ),
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      key: 'startTime',
      render: (time: number) => formatTimestamp(time),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: Job) => (
        <Space>
          <Button
            type="link"
            onClick={() => navigate(`/jobs/${record.jobId}`)}
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
      dataSource={data?.items || []}
      loading={isLoading}
      rowKey="jobId"
      pagination={{
        total: data?.pagination?.total || 0,
        pageSize: data?.pagination?.pageSize || 20,
        current: data?.pagination?.page || 1,
        showSizeChanger: true,
        showTotal: (total) => `共 ${total} 条`,
      }}
    />
  );
};

export default JobList;
