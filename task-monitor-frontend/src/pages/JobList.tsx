import React, { useState } from 'react';
import { Table, Space, Button, Tag } from 'antd';
import type { TablePaginationConfig, SorterResult, FilterValue } from 'antd/es/table/interface';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { StatusBadge } from '@/components/Common';
import { useGroupedJobs } from '@/hooks/useJobs';
import { useNodes } from '@/hooks/useNodes';
import { formatTimestamp, JOB_TYPE_MAP } from '@/utils';
import type { Job, JobListParams, JobGroup } from '@/types/job';

const JobList: React.FC = () => {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  // 从 URL 查询参数初始化状态
  const [params, setParams] = useState<JobListParams>(() => {
    const page = parseInt(searchParams.get('page') || '1', 10);
    const pageSize = parseInt(searchParams.get('pageSize') || '20', 10);
    const sortBy = searchParams.get('sortBy') || undefined;
    const sortOrder = (searchParams.get('sortOrder') as 'asc' | 'desc') || undefined;
    const status = searchParams.getAll('status');
    const type = searchParams.getAll('type');
    const framework = searchParams.getAll('framework');
    const nodeId = searchParams.get('nodeId') || undefined;

    return {
      page,
      pageSize,
      sortBy,
      sortOrder,
      status: status.length > 0 ? status : undefined,
      type: type.length > 0 ? type : undefined,
      framework: framework.length > 0 ? framework : undefined,
      nodeId,
    };
  });

  const { data, isLoading } = useGroupedJobs(params);
  const { data: nodesData } = useNodes();

  // 动态生成节点筛选选项
  const nodeFilters = (nodesData || []).map((n: any) => ({
    text: n.hostname || n.nodeId,
    value: n.nodeId,
  }));

  // 动态生成卡数筛选选项
  const cardCountFilters = Array.from(
    new Set((data?.items || []).map((g: JobGroup) => g.cardCount))
  ).sort((a, b) => a - b).map(c => ({ text: String(c), value: c }));
  const handleTableChange = (
    pagination: TablePaginationConfig,
    filters: Record<string, FilterValue | null>,
    sorter: SorterResult<JobGroup> | SorterResult<JobGroup>[]
  ) => {
    const newParams: JobListParams = {
      page: pagination.current || 1,
      pageSize: pagination.pageSize || 20,
    };

    // 处理排序
    if (!Array.isArray(sorter) && sorter.columnKey && sorter.order) {
      newParams.sortBy = sorter.columnKey as string;
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
    if (filters.nodeId && filters.nodeId.length > 0) {
      newParams.nodeId = filters.nodeId[0] as string;
    }

    // 更新状态
    setParams(newParams);

    // 更新 URL 查询参数
    const newSearchParams = new URLSearchParams();
    newSearchParams.set('page', String(newParams.page));
    newSearchParams.set('pageSize', String(newParams.pageSize));

    if (newParams.sortBy) {
      newSearchParams.set('sortBy', newParams.sortBy);
    }
    if (newParams.sortOrder) {
      newSearchParams.set('sortOrder', newParams.sortOrder);
    }
    if (newParams.status) {
      newParams.status.forEach(s => newSearchParams.append('status', s));
    }
    if (newParams.type) {
      newParams.type.forEach(t => newSearchParams.append('type', t));
    }
    if (newParams.framework) {
      newParams.framework.forEach(f => newSearchParams.append('framework', f));
    }
    if (newParams.nodeId) {
      newSearchParams.set('nodeId', newParams.nodeId);
    }

    setSearchParams(newSearchParams);
  };

  const columns = [
    {
      title: '作业名称',
      dataIndex: ['mainJob', 'jobName'],
      key: 'jobName',
      sorter: true,
      sortOrder: params.sortBy === 'jobName' ? (params.sortOrder === 'asc' ? 'ascend' : 'descend') : null,
      render: (text: string, record: JobGroup) => text || record.mainJob.jobId,
    },
    {
      title: '类型',
      dataIndex: ['mainJob', 'jobType'],
      key: 'jobType',
      filters: [
        { text: '训练', value: 'training' },
        { text: '推理', value: 'inference' },
        { text: '测试', value: 'testing' },
        { text: '未知', value: 'unknown' },
      ],
      filteredValue: params.type || null,
      sorter: true,
      sortOrder: params.sortBy === 'jobType' ? (params.sortOrder === 'asc' ? 'ascend' : 'descend') : null,
      render: (type: Job['jobType']) => (
        type ? <Tag>{JOB_TYPE_MAP[type] || type}</Tag> : '-'
      ),
    },
    {
      title: '框架',
      dataIndex: ['mainJob', 'framework'],
      key: 'framework',
      filters: [
        { text: 'PyTorch', value: 'pytorch' },
        { text: 'TensorFlow', value: 'tensorflow' },
        { text: 'MindSpore', value: 'mindspore' },
        { text: '其他', value: 'other' },
      ],
      filteredValue: params.framework || null,
      sorter: true,
      sortOrder: params.sortBy === 'framework' ? (params.sortOrder === 'asc' ? 'ascend' : 'descend') : null,
      render: (text: string) => text ? <Tag color="blue">{text}</Tag> : '-',
    },
    {
      title: '节点',
      dataIndex: ['mainJob', 'nodeId'],
      key: 'nodeId',
      filters: nodeFilters,
      filteredValue: params.nodeId ? [params.nodeId] : null,
      filterMultiple: false,
      sorter: true,
      sortOrder: params.sortBy === 'nodeId' ? (params.sortOrder === 'asc' ? 'ascend' : 'descend') : null,
      render: (text: string) => {
        const node = (nodesData || []).find((n: any) => n.nodeId === text);
        return node?.hostname || text || '-';
      },
    },
    {
      title: '状态',
      dataIndex: ['mainJob', 'status'],
      key: 'status',
      filters: [
        { text: '运行中', value: 'running' },
        { text: '已完成', value: 'completed' },
        { text: '失败', value: 'failed' },
        { text: '已停止', value: 'stopped' },
        { text: '丢失', value: 'lost' },
      ],
      filteredValue: params.status || null,
      sorter: true,
      sortOrder: params.sortBy === 'status' ? (params.sortOrder === 'asc' ? 'ascend' : 'descend') : null,
      render: (status: Job['status']) => (
        <StatusBadge status={status} type="job" />
      ),
    },
    {
      title: '开始时间',
      dataIndex: ['mainJob', 'startTime'],
      key: 'startTime',
      sorter: true,
      sortOrder: params.sortBy === 'startTime' ? (params.sortOrder === 'asc' ? 'ascend' : 'descend') : null,
      render: (time: number) => formatTimestamp(time),
    },
    {
      title: '卡数',
      dataIndex: 'cardCount',
      key: 'cardCount',
      width: 80,
      filters: cardCountFilters,
      onFilter: (value: any, record: JobGroup) => record.cardCount === value,
      sorter: (a: JobGroup, b: JobGroup) => a.cardCount - b.cardCount,
      render: (count: number) => (
        <Tag color={count > 1 ? 'orange' : 'default'}>{count}</Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      render: (_: any, record: JobGroup) => (
        <Space>
          <Button
            type="link"
            onClick={() => navigate(`/jobs/${record.mainJob.jobId}`)}
          >
            查看详情
          </Button>
        </Space>
      ),
    },
  ];

  // 子任务表格列
  const childColumns = [
    {
      title: 'PID',
      dataIndex: 'pid',
      key: 'pid',
    },
    {
      title: '进程名',
      dataIndex: 'processName',
      key: 'processName',
      render: (text: string) => text || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: Job['status']) => <StatusBadge status={status} type="job" />,
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
        <Button type="link" onClick={() => navigate(`/jobs/${record.jobId}`)}>
          查看详情
        </Button>
      ),
    },
  ];

  return (
    <Table<JobGroup>
      columns={columns}
      dataSource={data?.items || []}
      loading={isLoading}
      rowKey={(record) => record.mainJob.jobId}
      onChange={handleTableChange}
      expandable={{
        expandedRowRender: (record) => (
          <Table<Job>
            columns={childColumns}
            dataSource={record.childJobs}
            rowKey="jobId"
            pagination={false}
            size="small"
          />
        ),
        rowExpandable: (record) => record.cardCount > 1,
      }}
      pagination={{
        total: data?.pagination?.total || 0,
        pageSize: data?.pagination?.pageSize || 20,
        current: data?.pagination?.page || 1,
        showSizeChanger: true,
        showTotal: (total) => `共 ${total} 组`,
      }}
    />
  );
};

export default JobList;
