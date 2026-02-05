import React from 'react';
import { Card, Statistic } from 'antd';
import type { StatisticProps } from 'antd';

interface StatCardProps extends StatisticProps {
  loading?: boolean;
}

const StatCard: React.FC<StatCardProps> = ({ loading, ...props }) => {
  return (
    <Card loading={loading}>
      <Statistic {...props} />
    </Card>
  );
};

export default StatCard;
