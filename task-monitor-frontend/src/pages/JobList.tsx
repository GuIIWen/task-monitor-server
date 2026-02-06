import React, { useState } from 'react';
import { Table, Space, Button, Tag } from 'antd';
import type { TablePaginationConfig, SorterResult, FilterValue } from 'antd/es/table/interface';
import { useNavigate } from 'react-router-dom';
import { StatusBadge } from '@/components/Common';
import { useJobs } from '@/hooks';
import { formatTimestamp, JOB_TYPE_MAP } from '@/utils';
import type { Job, JobListParams } from '@/types/job';

const JobList: React.FC = () => {
  const navigate = useNavigate();
  const [params, setParams] = useState<JobListParams>({
    page: 1,
    pageSize: 20,
  });
  const { data, isLoading } = useJobs(params);

  // 处理表格变化（分页、排序、筛选）
  const handleTableChange = (
    pagination: TablePaginationConfig,
    filters: Record<string, FilterValue | null>,
    sorter: SorterResult<Job> | SorterResult<Job>[]
  ) => {
    const newParams: JobListParams = {
      page: pagination.current || 1,
      pageSize: pagination.pageSize || 20,
    };

    // 处理排序
    if (!Array.isArray(sorter) && sorter.field && sorter.order) {
      newParams.sortBy = sorter.field as string;
      newParams.sortOrder = sorter.order === 'ascend' ? 'asc' : 'desc';
    }

    // 处理筛选
    if (filters.status) {
      newParams.status = filters.status as string[];
    }
    if (filters.jobType) {
      newParams.type = filters.jobType as string[];
    }
    if (filters.framework) {
      newParams.framework = filters.framework as string[];
    }

    setParams(newParams);
  };

  const columns = [
    {
      title: '作业名称',
      dataIndex: 'jobName',
      key: 'jobName',
      sorter: true,
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
      sorter: true,
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
      sorter: true,
      render: (text: string) => text ? <Tag color="blue">{text}</Tag> : '-',
    },
    {
      title: '节点',
      dataIndex: 'nodeId',
      key: 'nodeId',
      sorter: true,
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
      sorter: true,
      render: (status: Job['status']) => (
        <StatusBadge status={status} type="job" />
      ),
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      key: 'startTime',
      sorter: true,
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
      onChange={handleTableChange}
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
