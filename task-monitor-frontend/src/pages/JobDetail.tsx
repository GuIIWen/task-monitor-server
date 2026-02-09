import React, { useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Collapse, Descriptions, Button, Space, Tag, Modal, Table, Progress, Typography, Spin, Alert, message } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { ArrowLeftOutlined, CodeOutlined, DownOutlined, RightOutlined, RobotOutlined } from '@ant-design/icons';
import { StatusBadge, LoadingSpinner } from '@/components/Common';
import { useJob, useJobCode, useJobParameters } from '@/hooks';
import { jobApi } from '@/api';
import { formatTimestamp, JOB_TYPE_MAP } from '@/utils';
import type { Job, NPUCardInfo, NPUMetricInfo, JobDetailResponse, JobAnalysis } from '@/types/job';

// 扁平化的 Chip 行，用于 NPU 卡表格展示
interface NPUChipRow {
  key: string;
  npuId: number;
  chipLabel: string;
  memoryUsageMb: number;
  metric: NPUMetricInfo;
}

function flattenNPUCards(cards: NPUCardInfo[]): NPUChipRow[] {
  const rows: NPUChipRow[] = [];
  for (const card of cards) {
    if (!card.metrics || card.metrics.length === 0) {
      rows.push({
        key: `npu-${card.npuId}`,
        npuId: card.npuId,
        chipLabel: `NPU ${card.npuId}`,
        memoryUsageMb: card.memoryUsageMb,
        metric: {} as NPUMetricInfo,
      });
      continue;
    }
    for (let i = 0; i < card.metrics.length; i++) {
      const m = card.metrics[i];
      const label = card.metrics.length > 1
        ? `NPU ${card.npuId} - Chip${i}`
        : `NPU ${card.npuId}`;
      rows.push({
        key: `npu-${card.npuId}-${m.busId ?? i}`,
        npuId: card.npuId,
        chipLabel: label,
        memoryUsageMb: card.memoryUsageMb,
        metric: m,
      });
    }
  }
  return rows;
}

const JobDetail: React.FC = () => {
  const { jobId } = useParams<{ jobId: string }>();
  const navigate = useNavigate();
  const { data: detail, isLoading } = useJob(jobId!);
  const { data: codeList } = useJobCode(jobId!);
  const { data: paramList } = useJobParameters(jobId!);
  const [codeModalOpen, setCodeModalOpen] = useState(false);
  const [envModalOpen, setEnvModalOpen] = useState(false);
  const [expandedRowKeys, setExpandedRowKeys] = useState<string[]>([]);
  const [childDetailMap, setChildDetailMap] = useState<Record<string, JobDetailResponse>>({});
  const [childLoadingMap, setChildLoadingMap] = useState<Record<string, boolean>>({});
  const [analysisData, setAnalysisData] = useState<JobAnalysis | null>(null);
  const [analysisLoading, setAnalysisLoading] = useState(false);

  const handleExpand = useCallback(async (expanded: boolean, record: Job) => {
    const key = record.jobId;
    if (!expanded) {
      setExpandedRowKeys(prev => prev.filter(k => k !== key));
      return;
    }
    setExpandedRowKeys(prev => [...prev, key]);
    if (childDetailMap[key]) return;
    setChildLoadingMap(prev => ({ ...prev, [key]: true }));
    try {
      const data = await jobApi.getJobById(key);
      setChildDetailMap(prev => ({ ...prev, [key]: data }));
    } catch {
      // 加载失败静默处理
    } finally {
      setChildLoadingMap(prev => ({ ...prev, [key]: false }));
    }
  }, [childDetailMap]);

  const handleAnalyze = useCallback(async () => {
    if (!jobId) return;
    setAnalysisLoading(true);
    try {
      const data = await jobApi.analyzeJob(jobId);
      setAnalysisData(data);
    } catch (e: any) {
      message.error(e?.message || 'AI 分析失败');
    } finally {
      setAnalysisLoading(false);
    }
  }, [jobId]);

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (!detail || !detail.job) {
    return <Card>作业不存在</Card>;
  }

  const job = detail.job;
  const npuCards = detail.npuCards || [];
  const relatedJobs = detail.relatedJobs || [];

  const npuChipRows = flattenNPUCards(npuCards);

  const npuChipColumns: ColumnsType<NPUChipRow> = [
    {
      title: '卡号 / Chip',
      dataIndex: 'chipLabel',
      width: 140,
    },
    {
      title: '进程显存 (MB)',
      dataIndex: 'memoryUsageMb',
      width: 120,
      render: (v: number) => v?.toFixed(1) ?? '-',
    },
    {
      title: '健康状态',
      key: 'health',
      width: 100,
      render: (_, record) => {
        const h = record.metric?.health;
        if (!h) return '-';
        return <Tag color={h === 'OK' ? 'green' : 'red'}>{h}</Tag>;
      },
    },
    {
      title: '功率 (W)',
      key: 'powerW',
      width: 100,
      render: (_, record) => record.metric?.powerW?.toFixed(1) ?? '-',
    },
    {
      title: '温度 (°C)',
      key: 'tempC',
      width: 100,
      render: (_, record) => record.metric?.tempC?.toFixed(1) ?? '-',
    },
    {
      title: 'AICore 使用率',
      key: 'aicore',
      width: 130,
      render: (_, record) => {
        const v = record.metric?.aicoreUsagePercent;
        if (v == null) return '-';
        return <Progress percent={Number(v.toFixed(1))} size="small" />;
      },
    },
    {
      title: 'HBM',
      key: 'hbm',
      width: 160,
      render: (_, record) => {
        const used = record.metric?.hbmUsageMb;
        const total = record.metric?.hbmTotalMb;
        if (used == null || total == null) return '-';
        const pct = total > 0 ? Number(((used / total) * 100).toFixed(1)) : 0;
        return (
          <Space direction="vertical" size={0} style={{ width: '100%' }}>
            <Progress percent={pct} size="small" />
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              {used.toFixed(0)} / {total.toFixed(0)} MB
            </Typography.Text>
          </Space>
        );
      },
    },
  ];

  // 子进程 NPU 卡表：使用扁平化 chip 行
  const childNpuColumns: ColumnsType<NPUChipRow> = [
    { title: '卡号 / Chip', dataIndex: 'chipLabel', width: 140 },
    {
      title: '健康状态',
      key: 'health',
      width: 100,
      render: (_, record) => {
        const h = record.metric?.health;
        if (!h) return '-';
        return <Tag color={h === 'OK' ? 'green' : 'red'}>{h}</Tag>;
      },
    },
    {
      title: '进程 HBM 占用',
      key: 'processHbm',
      width: 180,
      render: (_, record) => {
        const used = record.memoryUsageMb;
        const total = record.metric?.hbmTotalMb;
        if (used == null) return '-';
        if (total == null || total <= 0) return `${used.toFixed(0)} MB`;
        const pct = Number(((used / total) * 100).toFixed(1));
        return (
          <Space direction="vertical" size={0} style={{ width: '100%' }}>
            <Progress percent={pct} size="small" />
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              {used.toFixed(0)} / {total.toFixed(0)} MB
            </Typography.Text>
          </Space>
        );
      },
    },
  ];

  const relatedJobColumns: ColumnsType<Job> = [
    {
      title: 'PID',
      dataIndex: 'pid',
      width: 150,
      align: 'center',
    },
    {
      title: '进程名称',
      dataIndex: 'processName',
      ellipsis: true,
      align: 'center',
      render: (v: string | null) => v || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 150,
      align: 'center',
      render: (status: string | null) => <StatusBadge status={status as any} type="job" />,
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      width: 180,
      align: 'center',
      render: (v: number | null) => formatTimestamp(v),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      align: 'center',
      render: (_, record) => {
        const expanded = expandedRowKeys.includes(record.jobId);
        return (
          <Button
            type="link"
            size="small"
            icon={expanded ? <DownOutlined /> : <RightOutlined />}
            onClick={() => handleExpand(!expanded, record)}
          >
            {expanded ? '收起' : '详情'}
          </Button>
        );
      },
    },
  ];

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Button
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/jobs')}
      >
        返回列表
      </Button>

      <Card title="作业详情">
        <Descriptions bordered column={2}>
          <Descriptions.Item label="作业ID">{job.jobId}</Descriptions.Item>
          <Descriptions.Item label="作业名称">{job.jobName || '-'}</Descriptions.Item>
          <Descriptions.Item label="类型">
            {job.jobType ? <Tag>{JOB_TYPE_MAP[job.jobType] || job.jobType}</Tag> : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="框架">
            {job.framework ? <Tag color="blue">{job.framework}</Tag> : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="状态">
            <StatusBadge status={job.status} type="job" />
          </Descriptions.Item>
          <Descriptions.Item label="节点ID">{job.nodeId || '-'}</Descriptions.Item>
          <Descriptions.Item label="进程ID">{job.pid || '-'}</Descriptions.Item>
          <Descriptions.Item label="父进程ID">{job.ppid || '-'}</Descriptions.Item>
          <Descriptions.Item label="进程组ID">{job.pgid || '-'}</Descriptions.Item>
          <Descriptions.Item label="进程名称">{job.processName || '-'}</Descriptions.Item>
          <Descriptions.Item label="工作目录" span={2}>
            <code style={{ wordBreak: 'break-all' }}>{job.cwd || '-'}</code>
          </Descriptions.Item>
          <Descriptions.Item label="命令行" span={2}>
            <code style={{ wordBreak: 'break-all' }}>{job.commandLine || '-'}</code>
          </Descriptions.Item>
          <Descriptions.Item label="启动脚本" span={2}>
            {codeList && codeList.length > 0 ? (
              <Space>
                <code>{codeList[0].scriptPath || '-'}</code>
                <Button
                  type="link"
                  icon={<CodeOutlined />}
                  onClick={() => setCodeModalOpen(true)}
                >
                  查看代码
                </Button>
              </Space>
            ) : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="环境变量" span={2}>
            {(() => {
              const envStr = paramList?.[0]?.envVars;
              if (!envStr) return '-';
              try {
                const envObj = JSON.parse(envStr) as Record<string, string>;
                const entries = Object.entries(envObj);
                if (entries.length === 0) return '-';
                const preview = entries.slice(0, 3).map(([k, v]) => `${k}=${v}`).join('; ');
                const suffix = entries.length > 3 ? ` ... 共${entries.length}项` : '';
                return (
                  <Space>
                    <code style={{
                      fontSize: 12,
                      maxWidth: 600,
                      overflow: 'hidden',
                      textOverflow: 'ellipsis',
                      whiteSpace: 'nowrap',
                      display: 'inline-block',
                      verticalAlign: 'middle',
                    }}>{preview}{suffix}</code>
                    <Button type="link" size="small" onClick={() => setEnvModalOpen(true)}>
                      查看全部
                    </Button>
                  </Space>
                );
              } catch {
                return <code style={{ fontSize: 12 }}>{envStr.slice(0, 100)}</code>;
              }
            })()}
          </Descriptions.Item>
          <Descriptions.Item label="开始时间">
            {formatTimestamp(job.startTime)}
          </Descriptions.Item>
          <Descriptions.Item label="结束时间">
            {formatTimestamp(job.endTime)}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {npuCards.length === 0 ? (
        <Card title="NPU 卡信息">
          <Typography.Text type="secondary">该进程未占用 NPU</Typography.Text>
        </Card>
      ) : (
        <Collapse
          defaultActiveKey={npuChipRows.length <= 4 ? ['npu'] : []}
          items={[{
            key: 'npu',
            label: `NPU 卡信息 (${npuCards.length} 张${npuChipRows.length > npuCards.length ? `，${npuChipRows.length} 个 Chip` : ''})`,
            children: (
              <Table<NPUChipRow>
                dataSource={npuChipRows}
                rowKey="key"
                pagination={false}
                size="small"
                columns={npuChipColumns}
              />
            ),
          }]}
        />
      )}

      {relatedJobs.length > 0 && (
        <Card title={`关联 NPU 进程 (${relatedJobs.length})`}>
          <Table<Job>
            dataSource={relatedJobs}
            rowKey="jobId"
            pagination={false}
            size="small"
            bordered
            columns={relatedJobColumns}
            expandable={{
              expandedRowKeys,
              showExpandColumn: false,
              expandedRowRender: (record) => {
                const childDetail = childDetailMap[record.jobId];
                const loading = childLoadingMap[record.jobId];
                if (loading) {
                  return <Spin size="small" style={{ padding: 16 }} />;
                }
                if (!childDetail) {
                  return <Typography.Text type="secondary">加载失败</Typography.Text>;
                }
                const childCards = childDetail.npuCards || [];
                const childChipRows = flattenNPUCards(childCards);
                return (
                  <div style={{ padding: '8px 0' }}>
                    <Descriptions bordered size="small" column={2}>
                      <Descriptions.Item label="PID">{childDetail.job.pid ?? '-'}</Descriptions.Item>
                      <Descriptions.Item label="进程名称">{childDetail.job.processName ?? '-'}</Descriptions.Item>
                    </Descriptions>
                    {childChipRows.length > 0 && (
                      <Table<NPUChipRow>
                        dataSource={childChipRows}
                        rowKey="key"
                        pagination={false}
                        size="small"
                        columns={childNpuColumns}
                        style={{ marginTop: 8 }}
                      />
                    )}
                  </div>
                );
              },
            }}
          />
        </Card>
      )}

      <Card
        title={<Space><RobotOutlined />AI 智能分析</Space>}
        extra={
          <Button
            type="primary"
            icon={<RobotOutlined />}
            loading={analysisLoading}
            onClick={handleAnalyze}
          >
            {analysisData ? '重新分析' : '开始分析'}
          </Button>
        }
      >
        {analysisLoading && (
          <div style={{ textAlign: 'center', padding: 40 }}>
            <Spin size="large" />
            <div style={{ marginTop: 16 }}>
              <Typography.Text type="secondary">正在分析作业信息，请稍候...</Typography.Text>
            </div>
          </div>
        )}
        {!analysisLoading && !analysisData && (
          <Typography.Text type="secondary">
            点击"开始分析"按钮，AI 将综合分析作业的基本信息、NPU 资源、脚本代码、参数配置和环境变量。
          </Typography.Text>
        )}
        {!analysisLoading && analysisData && (
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <div>
              <Typography.Title level={5} style={{ marginBottom: 8 }}>作业概要</Typography.Title>
              <Typography.Paragraph>{analysisData.summary}</Typography.Paragraph>
            </div>

            <div>
              <Typography.Title level={5} style={{ marginBottom: 8 }}>作业类型</Typography.Title>
              <Space wrap>
                <Tag color="blue">
                  {analysisData.taskType.category === 'training' ? '训练' : analysisData.taskType.category === 'inference' ? '推理' : '未知'}
                </Tag>
                {analysisData.taskType.subCategory && (
                  <Tag color="cyan">{analysisData.taskType.subCategory}</Tag>
                )}
                {analysisData.taskType.inferenceFramework && (
                  <Tag color="purple">{analysisData.taskType.inferenceFramework}</Tag>
                )}
              </Space>
              {analysisData.taskType.evidence && (
                <Typography.Paragraph type="secondary" style={{ marginTop: 8, marginBottom: 0 }}>
                  判断依据：{analysisData.taskType.evidence}
                </Typography.Paragraph>
              )}
            </div>

            {analysisData.modelInfo && (
              <div>
                <Typography.Title level={5} style={{ marginBottom: 8 }}>模型信息</Typography.Title>
                <Descriptions bordered size="small" column={2}>
                  <Descriptions.Item label="模型名称">{analysisData.modelInfo.modelName || '-'}</Descriptions.Item>
                  <Descriptions.Item label="模型大小">{analysisData.modelInfo.modelSize || '-'}</Descriptions.Item>
                  <Descriptions.Item label="精度">{analysisData.modelInfo.precision || '-'}</Descriptions.Item>
                  <Descriptions.Item label="并行策略">{analysisData.modelInfo.parallelStrategy || '-'}</Descriptions.Item>
                </Descriptions>
              </div>
            )}

            {analysisData.runtimeAnalysis && (
              <div>
                <Typography.Title level={5} style={{ marginBottom: 8 }}>运行时长分析</Typography.Title>
                <Space direction="vertical" size="small" style={{ width: '100%' }}>
                  <Space>
                    <Typography.Text>运行时长：</Typography.Text>
                    <Typography.Text strong>{analysisData.runtimeAnalysis.duration}</Typography.Text>
                    <Tag color={
                      analysisData.runtimeAnalysis.status === 'normal' ? 'green' :
                      analysisData.runtimeAnalysis.status === 'completed' ? 'blue' :
                      analysisData.runtimeAnalysis.status === 'just-started' ? 'cyan' :
                      'orange'
                    }>
                      {analysisData.runtimeAnalysis.status === 'normal' ? '正常' :
                       analysisData.runtimeAnalysis.status === 'completed' ? '已完成' :
                       analysisData.runtimeAnalysis.status === 'just-started' ? '刚启动' : '长时间运行'}
                    </Tag>
                  </Space>
                  <Typography.Text type="secondary">{analysisData.runtimeAnalysis.description}</Typography.Text>
                </Space>
              </div>
            )}

            {analysisData.parameterCheck && analysisData.parameterCheck.items.length > 0 && (
              <div>
                <Typography.Title level={5} style={{ marginBottom: 8 }}>
                  参数检查
                  <Tag
                    color={
                      analysisData.parameterCheck.status === 'normal' ? 'green' :
                      analysisData.parameterCheck.status === 'warning' ? 'orange' : 'red'
                    }
                    style={{ marginLeft: 8 }}
                  >
                    {analysisData.parameterCheck.status === 'normal' ? '正常' :
                     analysisData.parameterCheck.status === 'warning' ? '警告' : '异常'}
                  </Tag>
                </Typography.Title>
                <Table
                  dataSource={analysisData.parameterCheck.items}
                  rowKey={(_, idx) => String(idx)}
                  pagination={false}
                  size="small"
                  bordered
                  columns={[
                    { title: '参数', dataIndex: 'parameter', width: 200 },
                    { title: '当前值', dataIndex: 'value', width: 150 },
                    {
                      title: '评估',
                      dataIndex: 'assessment',
                      width: 100,
                      render: (v: string) => (
                        <Tag color={v === 'normal' ? 'green' : v === 'warning' ? 'orange' : 'red'}>
                          {v === 'normal' ? '正常' : v === 'warning' ? '警告' : '异常'}
                        </Tag>
                      ),
                    },
                    { title: '说明', dataIndex: 'reason' },
                  ]}
                />
              </div>
            )}

            <div>
              <Typography.Title level={5} style={{ marginBottom: 8 }}>资源评估</Typography.Title>
              <Space direction="vertical" size="small" style={{ width: '100%' }}>
                <Space>
                  <Typography.Text>NPU 利用率：</Typography.Text>
                  <Tag color={
                    analysisData.resourceAssessment.npuUtilization === 'high' ? 'green' :
                    analysisData.resourceAssessment.npuUtilization === 'medium' ? 'orange' :
                    analysisData.resourceAssessment.npuUtilization === 'low' ? 'red' : 'default'
                  }>
                    {analysisData.resourceAssessment.npuUtilization}
                  </Tag>
                  <Typography.Text>HBM 利用率：</Typography.Text>
                  <Tag color={
                    analysisData.resourceAssessment.hbmUtilization === 'high' ? 'green' :
                    analysisData.resourceAssessment.hbmUtilization === 'medium' ? 'orange' : 'red'
                  }>
                    {analysisData.resourceAssessment.hbmUtilization}
                  </Tag>
                </Space>
                <Typography.Text type="secondary">{analysisData.resourceAssessment.description}</Typography.Text>
              </Space>
            </div>

            {analysisData.issues.length > 0 && (
              <div>
                <Typography.Title level={5} style={{ marginBottom: 8 }}>问题诊断</Typography.Title>
                <Space direction="vertical" size="small" style={{ width: '100%' }}>
                  {analysisData.issues.map((issue, idx) => (
                    <Alert
                      key={idx}
                      type={issue.severity === 'critical' ? 'error' : issue.severity === 'warning' ? 'warning' : 'info'}
                      showIcon
                      message={<span><Tag>{issue.category}</Tag>{issue.description}</span>}
                      description={<span>建议：{issue.suggestion}</span>}
                    />
                  ))}
                </Space>
              </div>
            )}
          </Space>
        )}
      </Card>

      {codeList && codeList.length > 0 && (
        <Modal
          title={`启动脚本 - ${codeList[0].scriptPath || ''}`}
          open={codeModalOpen}
          onCancel={() => setCodeModalOpen(false)}
          footer={null}
          width={900}
        >
          <pre style={{
            background: '#f5f5f5',
            padding: 16,
            borderRadius: 4,
            maxHeight: 600,
            overflow: 'auto',
            fontSize: 13,
            lineHeight: 1.6,
          }}>
            {codeList[0].scriptContent || '暂无内容'}
          </pre>
          {codeList[0].shScriptContent && (
            <>
              <h4 style={{ marginTop: 16 }}>Shell 启动脚本 - {codeList[0].shScriptPath}</h4>
              <pre style={{
                background: '#f5f5f5',
                padding: 16,
                borderRadius: 4,
                maxHeight: 300,
                overflow: 'auto',
                fontSize: 13,
                lineHeight: 1.6,
              }}>
                {codeList[0].shScriptContent}
              </pre>
            </>
          )}
        </Modal>
      )}

      <Modal
        title="环境变量"
        open={envModalOpen}
        onCancel={() => setEnvModalOpen(false)}
        footer={null}
        width={900}
      >
        {(() => {
          const envStr = paramList?.[0]?.envVars;
          if (!envStr) return <Typography.Text type="secondary">暂无环境变量</Typography.Text>;
          try {
            const envObj = JSON.parse(envStr) as Record<string, string>;
            const entries = Object.entries(envObj);
            if (entries.length === 0) return <Typography.Text type="secondary">暂无环境变量</Typography.Text>;
            return (
              <pre style={{
                background: '#f5f5f5',
                padding: 16,
                borderRadius: 4,
                maxHeight: 600,
                overflow: 'auto',
                fontSize: 13,
                lineHeight: 1.8,
              }}>
                {entries.map(([k, v]) => `${k}=${v}`).join('\n')}
              </pre>
            );
          } catch {
            return <pre style={{ background: '#f5f5f5', padding: 16, maxHeight: 600, overflow: 'auto' }}>{envStr}</pre>;
          }
        })()}
      </Modal>
    </Space>
  );
};

export default JobDetail;
