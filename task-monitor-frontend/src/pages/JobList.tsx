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
      sorter: (a: Job, b: Job) => {
        const aName = a.jobName || a.jobId;
        const bName = b.jobName || b.jobId;
        return aName.localeCompare(bName);
      },
      render: (text: string, record: Job) => text || record.jobId,
    },
    {
      title: '类型',
      dataIndex: 'jobType',
      key: 'jobType',
      filters: [
        { text: '训练', value: 'training' },
        { text: '推理', value: 'inference' },
        { text: '测试', value: 'testing' },
        { text: '未知', value: 'unknown' },
      ],
      onFilter: (value: string | number | boolean, record: Job) => {
        return record.jobType === value;
      },
      sorter: (a: Job, b: Job) => {
        const aType = a.jobType || 'unknown';
        const bType = b.jobType || 'unknown';
        return aType.localeCompare(bType);
      },
      render: (type: Job['jobType']) => (
        type ? <Tag>{JOB_TYPE_MAP[type] || type}</Tag> : '-'
      ),
    },
    {
      title: '框架',
      dataIndex: 'framework',
      key: 'framework',
      filters: [
        { text: 'PyTorch', value: 'pytorch' },
        { text: 'TensorFlow', value: 'tensorflow' },
        { text: 'MindSpore', value: 'mindspore' },
        { text: '其他', value: 'other' },
      ],
      onFilter: (value: string | number | boolean, record: Job) => {
        if (value === 'other') {
          return record.framework && !['pytorch', 'tensorflow', 'mindspore'].includes(record.framework.toLowerCase());
        }
        return record.framework?.toLowerCase() === value;
      },
      sorter: (a: Job, b: Job) => {
        const aFramework = a.framework || '';
        const bFramework = b.framework || '';
        return aFramework.localeCompare(bFramework);
      },
      render: (text: string) => text ? <Tag color="blue">{text}</Tag> : '-',
    },
    {
      title: '节点',
      dataIndex: 'nodeId',
      key: 'nodeId',
      sorter: (a: Job, b: Job) => {
        const aNode = a.nodeId || '';
        const bNode = b.nodeId || '';
        return aNode.localeCompare(bNode);
      },
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
      sorter: (a: Job, b: Job) => {
        const aTime = a.startTime || 0;
        const bTime = b.startTime || 0;
        return aTime - bTime;
      },
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
