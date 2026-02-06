import React from 'react';
import { Row, Col, Space } from 'antd';
import { useNavigate } from 'react-router-dom';
import { StatCard } from '@/components/Cards';
import { useNodeStats, useJobStats } from '@/hooks';
import { LoadingSpinner } from '@/components/Common';

const Dashboard: React.FC = () => {
  const navigate = useNavigate();
  const { data: nodeStats, isLoading: nodeStatsLoading } = useNodeStats();
  const { data: jobStats, isLoading: jobStatsLoading } = useJobStats();

  if (nodeStatsLoading || jobStatsLoading) {
    return <LoadingSpinner />;
  }

  return (
    <Space direction="vertical" size="large" style={{ width: '100%' }}>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="总节点数"
            value={nodeStats?.total || 0}
            loading={nodeStatsLoading}
            onClick={() => navigate('/nodes')}
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="活跃节点"
            value={nodeStats?.active || 0}
            valueStyle={{ color: '#3f8600' }}
            loading={nodeStatsLoading}
            onClick={() => navigate('/nodes?status=active')}
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="运行中作业"
            value={jobStats?.running || 0}
            valueStyle={{ color: '#1890ff' }}
            loading={jobStatsLoading}
            onClick={() => navigate('/jobs?status=running')}
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="总作业数"
            value={jobStats?.total || 0}
            loading={jobStatsLoading}
            onClick={() => navigate('/jobs')}
          />
        </Col>
      </Row>
    </Space>
  );
};

export default Dashboard;
