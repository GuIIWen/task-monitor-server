import React from 'react';
import { Card, Space, Typography, Tag } from 'antd';
import { AppstoreOutlined } from '@ant-design/icons';
import { StatusBadge } from '@/components/Common';
import { formatTimestamp, JOB_TYPE_MAP } from '@/utils';
import type { Job } from '@/types/job';

const { Text } = Typography;

interface JobCardProps {
  job: Job;
  onClick?: () => void;
}

const JobCard: React.FC<JobCardProps> = ({ job, onClick }) => {
  return (
    <Card
      hoverable
      onClick={onClick}
      style={{ cursor: onClick ? 'pointer' : 'default' }}
    >
      <Space direction="vertical" style={{ width: '100%' }}>
        <Space>
          <AppstoreOutlined style={{ fontSize: '20px' }} />
          <Text strong>{job.jobName || job.jobId}</Text>
          <StatusBadge status={job.status} type="job" />
        </Space>
        <Space direction="vertical" size="small">
          <Space>
            {job.jobType && (
              <Tag>{JOB_TYPE_MAP[job.jobType] || job.jobType}</Tag>
            )}
            {job.framework && <Tag color="blue">{job.framework}</Tag>}
          </Space>
          <Text type="secondary">节点: {job.nodeId || '-'}</Text>
          <Text type="secondary">
            开始时间: {formatTimestamp(job.startTime)}
          </Text>
        </Space>
      </Space>
    </Card>
  );
};

export default JobCard;
