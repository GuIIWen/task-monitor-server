import React, { useState, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Space, Tag, Modal, Table, Progress, Typography, Spin } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { ArrowLeftOutlined, CodeOutlined, DownOutlined, RightOutlined } from '@ant-design/icons';
import { StatusBadge, LoadingSpinner } from '@/components/Common';
import { useJob, useJobCode, useJobParameters } from '@/hooks';
import { jobApi } from '@/api';
import { formatTimestamp, JOB_TYPE_MAP } from '@/utils';
import type { Job, NPUCardInfo, JobDetailResponse } from '@/types/job';

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

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (!detail || !detail.job) {
    return <Card>作业不存在</Card>;
  }

  const job = detail.job;
  const npuCards = detail.npuCards || [];
  const relatedJobs = detail.relatedJobs || [];

  const npuCardColumns: ColumnsType<NPUCardInfo> = [
    {
      title: '卡号',
      dataIndex: 'npuId',
      width: 80,
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

  // 子进程 NPU 卡表：HBM 列显示该进程自己的显存占用，而非整卡 HBM
  const childNpuColumns: ColumnsType<NPUCardInfo> = [
    { title: '卡号', dataIndex: 'npuId', width: 80 },
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

  const relatedJobColumns: ColumnsType<Job> = [
    {
      title: 'PID',
      dataIndex: 'pid',
      width: 100,
    },
    {
      title: '进程名称',
      dataIndex: 'processName',
      ellipsis: true,
      render: (v: string | null) => v || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (status: string | null) => <StatusBadge status={status as any} type="job" />,
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      width: 180,
      render: (v: number | null) => formatTimestamp(v),
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
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
          <Descriptions.Item label="工作目录" span={2}>{job.cwd || '-'}</Descriptions.Item>
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
                    <Typography.Text style={{ fontSize: 12 }} ellipsis>{preview}{suffix}</Typography.Text>
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

      <Card title={`NPU 卡信息${npuCards.length > 0 ? ` (${npuCards.length} 张)` : ''}`}>
        {npuCards.length === 0 ? (
          <Typography.Text type="secondary">该进程未占用 NPU</Typography.Text>
        ) : (
          <Table<NPUCardInfo>
            dataSource={npuCards}
            rowKey="npuId"
            pagination={false}
            size="small"
            columns={npuCardColumns}
          />
        )}
      </Card>

      {relatedJobs.length > 0 && (
        <Card title={`关联 NPU 进程 (${relatedJobs.length})`}>
          <Table<Job>
            dataSource={relatedJobs}
            rowKey="jobId"
            pagination={false}
            size="small"
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
                return (
                  <div style={{ padding: '8px 0' }}>
                    <Descriptions bordered size="small" column={2}>
                      <Descriptions.Item label="PID">{childDetail.job.pid ?? '-'}</Descriptions.Item>
                      <Descriptions.Item label="进程名称">{childDetail.job.processName ?? '-'}</Descriptions.Item>
                      <Descriptions.Item label="命令行" span={2}>
                        <code style={{ wordBreak: 'break-all', fontSize: 12 }}>
                          {childDetail.job.commandLine ?? '-'}
                        </code>
                      </Descriptions.Item>
                    </Descriptions>
                    {childCards.length > 0 && (
                      <Table<NPUCardInfo>
                        dataSource={childCards}
                        rowKey={(r) => `${r.npuId}-${r.metric?.busId ?? ''}`}
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
