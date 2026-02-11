import React, { useState, useRef, useCallback, useEffect, useMemo } from 'react';
import { Table, Space, Button, Tag, Modal, Progress, Tooltip, Dropdown, message } from 'antd';
import { CheckCircleOutlined, WarningOutlined, LoadingOutlined, MinusOutlined, ExpandOutlined, CloseOutlined } from '@ant-design/icons';
import type { TablePaginationConfig, SorterResult, FilterValue } from 'antd/es/table/interface';
import type { MenuProps } from 'antd';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { StatusBadge } from '@/components/Common';
import { useGroupedJobs, useDistinctCardCounts } from '@/hooks/useJobs';
import { useNodes } from '@/hooks/useNodes';
import { jobApi } from '@/api/job';
import { formatTimestamp, JOB_TYPE_MAP } from '@/utils';
import type { Job, JobListParams, JobGroup, JobAnalysis } from '@/types/job';

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
  const [exportLoading, setExportLoading] = useState(false);

  // 批量分析浮动进度状态
  const BATCH_KEY = 'batch-analyze-id';
  type AnalyzeProgress = {
    status: 'running' | 'done' | 'cancelled';
    current: number;
    total: number;
    success: number;
    failed: number;
    failedItems?: { jobId: string; error: string }[];
  };
  const [analyzeProgress, setAnalyzeProgress] = useState<AnalyzeProgress | null>(null);
  const [progressMinimized, setProgressMinimized] = useState(false);
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const batchIdRef = useRef<string | null>(null);

  // 轮询批量分析进度
  const startPolling = useCallback((batchId: string) => {
    batchIdRef.current = batchId;
    if (pollingRef.current) clearInterval(pollingRef.current);
    pollingRef.current = setInterval(async () => {
      try {
        const progress = await jobApi.getBatchAnalyzeProgress(batchId);
        setAnalyzeProgress(progress);
        if (progress.status !== 'running') {
          if (pollingRef.current) clearInterval(pollingRef.current);
          pollingRef.current = null;
          sessionStorage.removeItem(BATCH_KEY);
        }
      } catch {
        if (pollingRef.current) clearInterval(pollingRef.current);
        pollingRef.current = null;
      }
    }, 2000);
  }, []);

  // 页面加载时恢复轮询
  useEffect(() => {
    const savedBatchId = sessionStorage.getItem(BATCH_KEY);
    if (savedBatchId) {
      batchIdRef.current = savedBatchId;
      jobApi.getBatchAnalyzeProgress(savedBatchId).then((progress) => {
        setAnalyzeProgress(progress);
        if (progress.status === 'running') {
          startPolling(savedBatchId);
        } else {
          sessionStorage.removeItem(BATCH_KEY);
        }
      }).catch(() => {
        sessionStorage.removeItem(BATCH_KEY);
      });
    }
    return () => {
      if (pollingRef.current) clearInterval(pollingRef.current);
    };
  }, [startPolling]);

  const { data, isLoading } = useGroupedJobs(params);
  const { data: nodesData } = useNodes();
  const { data: cardCountsData } = useDistinctCardCounts();

  // 批量拉取当前页所有作业的分析摘要
  const [analysesMap, setAnalysesMap] = useState<Record<string, JobAnalysis>>({});
  const analysesMapRef = useRef(analysesMap);
  analysesMapRef.current = analysesMap;
  useEffect(() => {
    const items = data?.items;
    if (!items || items.length === 0) return;
    const jobIds = items.map((g) => g.mainJob.jobId);
    jobApi.getBatchAnalyses(jobIds).then(setAnalysesMap).catch(() => {});
  }, [data]);

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
    batchIdRef.current = null;
  }, []);

  const handleCancelBatch = useCallback(async () => {
    const batchId = batchIdRef.current;
    if (!batchId) return;
    try {
      await jobApi.cancelBatchAnalyze(batchId);
    } catch {
      // ignore
    }
  }, []);

  const handleBatchAnalyze = () => {
    if (analyzeProgress?.status === 'running') return;
    const jobIds = [...selectedRowKeys];

    Modal.confirm({
      title: '批量分析确认',
      content: (<div>
        <div>{`确定要对选中的 ${jobIds.length} 个作业进行 AI 分析吗？`}</div>
        <div style={{ marginTop: 8, color: '#999', fontSize: 12 }}>提示：批量分析数量较多时，请确保 LLM 超时时间设置足够大，避免请求超时导致分析失败。</div>
      </div>),
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        setSelectedRowKeys([]);
        setProgressMinimized(false);
        setAnalyzeProgress({ status: 'running', current: 0, total: jobIds.length, success: 0, failed: 0 });

        try {
          const { batchId } = await jobApi.batchAnalyze(jobIds);
          sessionStorage.setItem(BATCH_KEY, batchId);
          startPolling(batchId);
        } catch {
          setAnalyzeProgress(null);
        }
      },
    });
  };

  const handleExport = useCallback(async (scope: 'filtered' | 'page' | 'selected') => {
    if (scope === 'selected' && selectedRowKeys.length === 0) {
      message.warning('请先勾选要导出的作业');
      return;
    }

    const exportParams: JobListParams & { scope: 'filtered' | 'page' | 'selected'; jobIds?: string[] } = {
      ...params,
      scope,
    };

    if (scope === 'selected') {
      exportParams.jobIds = selectedRowKeys;
    }
    if (scope === 'filtered') {
      delete exportParams.page;
      delete exportParams.pageSize;
    }

    try {
      setExportLoading(true);
      const blob = await jobApi.exportAnalysesCSV(exportParams);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'ai-analysis-overview_' + new Date().toISOString().slice(0, 19).replace(/[T:]/g, '-') + '.csv';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(url);
      message.success('导出成功');
    } catch (e: any) {
      message.error(e?.message || '导出失败');
    } finally {
      setExportLoading(false);
    }
  }, [params, selectedRowKeys]);

  const exportMenuItems: MenuProps['items'] = [
    { key: 'filtered', label: '导出当前筛选' },
    { key: 'page', label: '导出当前页' },
    { key: 'selected', label: '导出已勾选', disabled: selectedRowKeys.length === 0 },
  ];

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

  // 自定义行组件：在数据行下方渲染分析预览条
  const AnalysisRow = useMemo(() => {
    const Comp = (props: any) => {
      const { 'data-job-id': jobId, ...restProps } = props;
      const analysis = jobId ? analysesMapRef.current[jobId] : null;
      if (!analysis) return <tr {...restProps} />;
      return (
        <>
          <tr {...restProps} />
          <tr className="analysis-preview-row" style={{ background: '#fafbfc' }}>
            <td colSpan={100} style={{ padding: '6px 12px 6px 52px', borderBottom: '1px solid #f0f0f0' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13, color: '#555' }}>
                <div style={{ display: 'flex', gap: 6, alignItems: 'center', flexShrink: 0 }}>
                  {analysis.modelInfo?.modelName && (
                    <Tag color="blue" style={{ margin: 0 }}>
                      {analysis.modelInfo.modelName}
                      {analysis.modelInfo.modelSize ? ` ${analysis.modelInfo.modelSize}` : ''}
                    </Tag>
                  )}
                  {analysis.runtimeAnalysis?.duration && (
                    <Tag style={{ margin: 0 }}>{analysis.runtimeAnalysis.duration}</Tag>
                  )}
                  {(() => {
                    const allIssues = analysis.issues || [];
                    const problems = allIssues.filter((i: any) => i.severity !== 'info');
                    const tips = allIssues.filter((i: any) => i.severity === 'info');
                    const pCount = problems.length;
                    const pColor = pCount === 0 ? 'default' : pCount <= 2 ? 'orange' : 'red';
                    const elements: React.ReactNode[] = [];
                    // 问题标签
                    const problemTag = (
                      <Tag color={pColor} style={{ margin: 0, cursor: pCount > 0 ? 'pointer' : 'default' }}>
                        {pCount}个问题
                      </Tag>
                    );
                    if (pCount === 0) {
                      elements.push(problemTag);
                    } else {
                      elements.push(
                        <Tooltip
                          key="problems"
                          overlayStyle={{ maxWidth: 400 }}
                          title={
                            <div style={{ maxWidth: 360 }}>
                              {problems.map((issue: any, idx: number) => (
                                <div key={idx} style={{ marginBottom: idx < pCount - 1 ? 8 : 0 }}>
                                  <div>
                                    <Tag
                                      color={issue.severity === 'critical' ? 'red' : 'orange'}
                                      style={{ margin: '0 4px 0 0' }}
                                    >{issue.severity}</Tag>
                                    <span style={{ fontWeight: 500 }}>{issue.category}</span>
                                  </div>
                                  <div style={{ marginTop: 2, color: 'rgba(255,255,255,0.85)' }}>{issue.description}</div>
                                  {issue.suggestion && (
                                    <div style={{ marginTop: 2, color: 'rgba(255,255,255,0.65)', fontSize: 12 }}>建议：{issue.suggestion}</div>
                                  )}
                                </div>
                              ))}
                            </div>
                          }
                        >{problemTag}</Tooltip>
                      );
                    }
                    // 提示标签
                    if (tips.length > 0) {
                      const tipTag = (
                        <Tag color="blue" style={{ margin: 0, cursor: 'pointer' }}>
                          {tips.length}个提示
                        </Tag>
                      );
                      elements.push(
                        <Tooltip
                          key="tips"
                          overlayStyle={{ maxWidth: 400 }}
                          title={
                            <div style={{ maxWidth: 360 }}>
                              {tips.map((issue: any, idx: number) => (
                                <div key={idx} style={{ marginBottom: idx < tips.length - 1 ? 8 : 0 }}>
                                  <div><span style={{ fontWeight: 500 }}>{issue.category}</span></div>
                                  <div style={{ marginTop: 2, color: 'rgba(255,255,255,0.85)' }}>{issue.description}</div>
                                  {issue.suggestion && (
                                    <div style={{ marginTop: 2, color: 'rgba(255,255,255,0.65)', fontSize: 12 }}>建议：{issue.suggestion}</div>
                                  )}
                                </div>
                              ))}
                            </div>
                          }
                        >{tipTag}</Tooltip>
                      );
                    }
                    return elements;
                  })()}
                </div>
                <span style={{ color: '#ddd', flexShrink: 0 }}>|</span>
                <span style={{
                  flex: 1, overflow: 'hidden', textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap', color: '#888',
                }}>
                  {analysis.summary}
                </span>
              </div>
            </td>
          </tr>
        </>
      );
    };
    return Comp;
  }, []);

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span>
          {selectedRowKeys.length > 0 ? `已选择 ${selectedRowKeys.length} 项` : ''}
        </span>
        <Space>
          <Button
            type="primary"
            disabled={selectedRowKeys.length === 0}
            onClick={handleBatchAnalyze}
          >
            批量分析
          </Button>
          <Dropdown
            menu={{
              items: exportMenuItems,
              onClick: ({ key }) => handleExport(key as 'filtered' | 'page' | 'selected'),
            }}
          >
            <Button loading={exportLoading}>导出</Button>
          </Dropdown>
        </Space>
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
        onRow={(record) => ({ 'data-job-id': record.mainJob.jobId } as any)}
        components={{ body: { row: AnalysisRow } }}
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
            background: analyzeProgress.status === 'running' ? '#1890ff'
              : analyzeProgress.status === 'cancelled' ? '#faad14'
              : analyzeProgress.failed > 0 ? '#faad14' : '#52c41a',
            color: '#fff', borderRadius: 20, padding: '8px 16px',
            cursor: 'pointer', boxShadow: '0 4px 12px rgba(0,0,0,0.2)',
            display: 'flex', alignItems: 'center', gap: 8, fontSize: 13,
          }}
        >
          {analyzeProgress.status === 'running'
            ? <><LoadingOutlined /> 分析中 {analyzeProgress.current}/{analyzeProgress.total}</>
            : analyzeProgress.status === 'cancelled'
              ? <><ExpandOutlined /> 已取消 {analyzeProgress.success}/{analyzeProgress.total}</>
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
            background: analyzeProgress.status === 'running' ? '#1890ff'
              : analyzeProgress.status === 'cancelled' ? '#faad14'
              : analyzeProgress.failed > 0 ? '#faad14' : '#52c41a',
            color: '#fff',
          }}>
            <span style={{ fontWeight: 500, display: 'flex', alignItems: 'center', gap: 8 }}>
              {analyzeProgress.status === 'running'
                ? <><LoadingOutlined /> 批量分析进行中</>
                : analyzeProgress.status === 'cancelled'
                  ? <><WarningOutlined /> 批量分析已取消</>
                  : analyzeProgress.failed > 0
                    ? <><WarningOutlined /> 批量分析完成（部分失败）</>
                    : <><CheckCircleOutlined /> 批量分析完成</>
              }
            </span>
            <span style={{ display: 'flex', gap: 8 }}>
              {analyzeProgress.status === 'running' && (
                <CloseOutlined onClick={handleCancelBatch} style={{ cursor: 'pointer' }} title="取消" />
              )}
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
              status={analyzeProgress.status === 'running' ? 'active'
                : (analyzeProgress.failed > 0 || analyzeProgress.status === 'cancelled') ? 'exception' : 'success'}
            />
            <div style={{ marginTop: 8, color: '#666', fontSize: 13 }}>
              {analyzeProgress.status === 'running'
                ? <>进度 {analyzeProgress.current}/{analyzeProgress.total}，成功 {analyzeProgress.success}，失败 {analyzeProgress.failed}</>
                : analyzeProgress.status === 'cancelled'
                  ? <>已取消，共处理 {analyzeProgress.current}/{analyzeProgress.total}：{analyzeProgress.success} 个成功{analyzeProgress.failed > 0 ? `，${analyzeProgress.failed} 个失败` : ''}</>
                  : <>共 {analyzeProgress.total} 个作业：{analyzeProgress.success} 个成功{analyzeProgress.failed > 0 ? `，${analyzeProgress.failed} 个失败` : ''}</>
              }
            </div>
            {analyzeProgress.failedItems && analyzeProgress.failedItems.length > 0 && analyzeProgress.status !== 'running' && (
              <div style={{ marginTop: 8, maxHeight: 120, overflowY: 'auto', fontSize: 12 }}>
                <div style={{ color: '#999', marginBottom: 4 }}>失败详情：</div>
                {analyzeProgress.failedItems.map((item, idx) => (
                  <div key={idx} style={{ color: '#f5222d', marginBottom: 2, wordBreak: 'break-all' }}>
                    <span style={{ color: '#666' }}>{item.jobId}：</span>{item.error}
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      )
    )}
    </div>
  );
};

export default JobList;
