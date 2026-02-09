import React, { useState, useRef, useCallback, useEffect } from 'react';
import { Table, Space, Button, Tag, Modal, Progress } from 'antd';
import { CheckCircleOutlined, WarningOutlined, LoadingOutlined, MinusOutlined, ExpandOutlined, CloseOutlined } from '@ant-design/icons';
import type { TablePaginationConfig, SorterResult, FilterValue } from 'antd/es/table/interface';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { StatusBadge } from '@/components/Common';
import { useGroupedJobs, useDistinctCardCounts } from '@/hooks/useJobs';
import { useNodes } from '@/hooks/useNodes';
import { jobApi } from '@/api/job';
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
    const cardCount = searchParams.getAll('cardCount').map(v => v === 'unknown' ? v : Number(v)).filter(v => v === 'unknown' || !isNaN(v as number));

    return {
      page,
      pageSize,
      sortBy,
      sortOrder,
      status: status.length > 0 ? status : undefined,
      type: type.length > 0 ? type : undefined,
      framework: framework.length > 0 ? framework : undefined,
      nodeId,
      cardCount: cardCount.length > 0 ? cardCount : undefined,
    };
  });

  const [selectedRowKeys, setSelectedRowKeys] = useState<string[]>([]);
  const analyzingRef = useRef(false);

  // 批量分析浮动进度状态
  const STORAGE_KEY = 'batch-analyze-progress';
  type AnalyzeProgress = {
    status: 'running' | 'done' | 'interrupted';
    current: number;
    total: number;
    success: number;
    failed: number;
  };
  const [analyzeProgress, setAnalyzeProgress] = useState<AnalyzeProgress | null>(() => {
    try {
      const saved = sessionStorage.getItem(STORAGE_KEY);
      if (!saved) return null;
      const parsed = JSON.parse(saved) as AnalyzeProgress;
      if (parsed.status === 'running') {
        return { ...parsed, status: 'interrupted' };
      }
      return parsed;
    } catch {
      return null;
    }
  });
  const [progressMinimized, setProgressMinimized] = useState(false);

  // 持久化进度到 sessionStorage
  useEffect(() => {
    if (analyzeProgress) {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(analyzeProgress));
    } else {
      sessionStorage.removeItem(STORAGE_KEY);
    }
  }, [analyzeProgress]);

  // 分析进行中拦截页面刷新/关闭
  useEffect(() => {
    if (analyzeProgress?.status !== 'running') return;
    const handler = (e: BeforeUnloadEvent) => {
      e.preventDefault();
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, [analyzeProgress?.status]);

  const { data, isLoading } = useGroupedJobs(params);
  const { data: nodesData } = useNodes();
  const { data: cardCountsData } = useDistinctCardCounts();

  // 动态生成节点筛选选项
  const nodeFilters = (nodesData || []).map((n: any) => ({
    text: n.hostname || n.nodeId,
    value: n.nodeId,
  }));

  // 动态生成卡数筛选选项
  const cardCountFilters = [
    { text: 'unknown', value: 'unknown' },
    ...(cardCountsData || [])
      .sort((a: number, b: number) => a - b)
      .map((c: number) => ({ text: String(c), value: c })),
  ];
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

    // 卡数筛选（走后端）
    if (filters.cardCount && filters.cardCount.length > 0) {
      newParams.cardCount = filters.cardCount as (number | string)[];
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
    if (newParams.cardCount) {
      newParams.cardCount.forEach(c => newSearchParams.append('cardCount', String(c)));
    }

    setSearchParams(newSearchParams);
  };

  const dismissProgress = useCallback(() => {
    setAnalyzeProgress(null);
    setProgressMinimized(false);
  }, []);

  const handleBatchAnalyze = () => {
    if (analyzingRef.current) return;
    const jobIds = [...selectedRowKeys];

    Modal.confirm({
      title: '批量分析确认',
      content: `确定要对选中的 ${jobIds.length} 个作业进行 AI 分析吗？`,
      okText: '确定',
      cancelText: '取消',
      onOk: () => {
        analyzingRef.current = true;
        setSelectedRowKeys([]);
        setProgressMinimized(false);
        setAnalyzeProgress({ status: 'running', current: 0, total: jobIds.length, success: 0, failed: 0 });

        const run = async () => {
          let success = 0;
          let failed = 0;
          const total = jobIds.length;

          for (let i = 0; i < total; i++) {
            setAnalyzeProgress({ status: 'running', current: i, total, success, failed });
            try {
              await jobApi.analyzeJob(jobIds[i]);
              success++;
            } catch {
              failed++;
            }
          }

          analyzingRef.current = false;
          setAnalyzeProgress({ status: 'done', current: total, total, success, failed });
        };

        run();
      },
    });
  };

  const columns = [
    {
      title: '作业名称',
      dataIndex: ['mainJob', 'jobName'],
      key: 'jobName',
      sorter: true,
      sortOrder: params.sortBy === 'jobName' ? (params.sortOrder === 'asc' ? 'ascend' as const : 'descend' as const) : null,
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
      sortOrder: params.sortBy === 'jobType' ? (params.sortOrder === 'asc' ? 'ascend' as const : 'descend' as const) : null,
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
      sortOrder: params.sortBy === 'framework' ? (params.sortOrder === 'asc' ? 'ascend' as const : 'descend' as const) : null,
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
      sortOrder: params.sortBy === 'nodeId' ? (params.sortOrder === 'asc' ? 'ascend' as const : 'descend' as const) : null,
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
      sortOrder: params.sortBy === 'status' ? (params.sortOrder === 'asc' ? 'ascend' as const : 'descend' as const) : null,
      render: (status: Job['status']) => (
        <StatusBadge status={status} type="job" />
      ),
    },
    {
      title: '开始时间',
      dataIndex: ['mainJob', 'startTime'],
      key: 'startTime',
      sorter: true,
      sortOrder: params.sortBy === 'startTime' ? (params.sortOrder === 'asc' ? 'ascend' as const : 'descend' as const) : null,
      render: (time: number) => formatTimestamp(time),
    },
    {
      title: '卡数',
      dataIndex: 'cardCount',
      key: 'cardCount',
      width: 80,
      filters: cardCountFilters,
      filteredValue: params.cardCount || null,
      sorter: true,
      sortOrder: params.sortBy === 'cardCount' ? (params.sortOrder === 'asc' ? 'ascend' as const : 'descend' as const) : null,
      render: (count: number | null) => (
        count === null
          ? <Tag color="default">unknown</Tag>
          : <Tag color={count > 1 ? 'orange' : 'default'}>{count}</Tag>
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
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span>
          {selectedRowKeys.length > 0 ? `已选择 ${selectedRowKeys.length} 项` : ''}
        </span>
        <Button
          type="primary"
          disabled={selectedRowKeys.length === 0}
          onClick={handleBatchAnalyze}
        >
          批量分析
        </Button>
      </div>
      <Table<JobGroup>
        columns={columns}
        dataSource={data?.items || []}
        loading={isLoading}
        rowKey={(record) => record.mainJob.jobId}
        rowSelection={{
          selectedRowKeys,
          onChange: (keys) => setSelectedRowKeys(keys as string[]),
        }}
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
        rowExpandable: (record) => record.childJobs.length > 0,
      }}
      pagination={{
        total: data?.pagination?.total || 0,
        pageSize: data?.pagination?.pageSize || 20,
        current: data?.pagination?.page || 1,
        showSizeChanger: true,
        showTotal: (total) => `共 ${total} 组`,
      }}
    />

    {/* 批量分析浮动进度 */}
    {analyzeProgress && (
      progressMinimized ? (
        <div
          onClick={() => setProgressMinimized(false)}
          style={{
            position: 'fixed', bottom: 24, right: 24, zIndex: 1050,
            background: analyzeProgress.status === 'running' ? '#1890ff' : (analyzeProgress.failed > 0 || analyzeProgress.status === 'interrupted' ? '#faad14' : '#52c41a'),
            color: '#fff', borderRadius: 20, padding: '8px 16px',
            cursor: 'pointer', boxShadow: '0 4px 12px rgba(0,0,0,0.2)',
            display: 'flex', alignItems: 'center', gap: 8, fontSize: 13,
          }}
        >
          {analyzeProgress.status === 'running'
            ? <><LoadingOutlined /> 分析中 {analyzeProgress.current}/{analyzeProgress.total}</>
            : analyzeProgress.status === 'interrupted'
              ? <><WarningOutlined /> 分析已中断 {analyzeProgress.success}/{analyzeProgress.total}</>
              : <><ExpandOutlined /> 分析完成 {analyzeProgress.success}/{analyzeProgress.total}</>
          }
        </div>
      ) : (
        <div style={{
          position: 'fixed', bottom: 24, right: 24, zIndex: 1050,
          width: 320, background: '#fff', borderRadius: 8,
          boxShadow: '0 6px 20px rgba(0,0,0,0.15)', overflow: 'hidden',
        }}>
          <div style={{
            padding: '10px 16px', display: 'flex', justifyContent: 'space-between', alignItems: 'center',
            background: analyzeProgress.status === 'running' ? '#1890ff' : (analyzeProgress.failed > 0 || analyzeProgress.status === 'interrupted' ? '#faad14' : '#52c41a'),
            color: '#fff',
          }}>
            <span style={{ fontWeight: 500, display: 'flex', alignItems: 'center', gap: 8 }}>
              {analyzeProgress.status === 'running'
                ? <><LoadingOutlined /> 批量分析进行中</>
                : analyzeProgress.status === 'interrupted'
                  ? <><WarningOutlined /> 批量分析已中断</>
                  : analyzeProgress.failed > 0
                    ? <><WarningOutlined /> 批量分析完成（部分失败）</>
                    : <><CheckCircleOutlined /> 批量分析完成</>
              }
            </span>
            <span style={{ display: 'flex', gap: 8 }}>
              <MinusOutlined onClick={() => setProgressMinimized(true)} style={{ cursor: 'pointer' }} />
              {analyzeProgress.status !== 'running' && (
                <CloseOutlined onClick={dismissProgress} style={{ cursor: 'pointer' }} />
              )}
            </span>
          </div>
          <div style={{ padding: '12px 16px' }}>
            <Progress
              percent={analyzeProgress.total > 0 ? Math.round((analyzeProgress.current / analyzeProgress.total) * 100) : 0}
              size="small"
              status={analyzeProgress.status === 'running' ? 'active' : (analyzeProgress.failed > 0 || analyzeProgress.status === 'interrupted' ? 'exception' : 'success')}
            />
            <div style={{ marginTop: 8, color: '#666', fontSize: 13 }}>
              {analyzeProgress.status === 'running'
                ? <>进度 {analyzeProgress.current}/{analyzeProgress.total}，成功 {analyzeProgress.success}，失败 {analyzeProgress.failed}</>
                : analyzeProgress.status === 'interrupted'
                  ? <>页面刷新导致中断，已完成 {analyzeProgress.success}/{analyzeProgress.total}</>
                  : <>共 {analyzeProgress.total} 个作业：{analyzeProgress.success} 个成功{analyzeProgress.failed > 0 ? `，${analyzeProgress.failed} 个失败` : ''}</>
              }
            </div>
          </div>
        </div>
      )
    )}
    </div>
  );
};

export default JobList;
