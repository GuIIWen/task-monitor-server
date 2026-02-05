import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Card, Descriptions, Button, Space, Tag } from 'antd';
import { ArrowLeftOutlined } from '@ant-design/icons';
import { StatusBadge, LoadingSpinner } from '@/components/Common';
import { useJob } from '@/hooks';
import { formatTimestamp, JOB_TYPE_MAP } from '@/utils';

const JobDetail: React.FC = () => {
  const { jobId } = useParams<{ jobId: string }>();
  const navigate = useNavigate();
  const { data: job, isLoading } = useJob(jobId!);

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (!job) {
    return <Card>作业不存在</Card>;
  }

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
          <Descriptions.Item label="开始时间">
            {formatTimestamp(job.startTime)}
          </Descriptions.Item>
          <Descriptions.Item label="结束时间">
            {formatTimestamp(job.endTime)}
          </Descriptions.Item>
        </Descriptions>
      </Card>
    </Space>
  );
};

export default JobDetail;
