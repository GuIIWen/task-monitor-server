import React from 'react';
import { Card, Statistic } from 'antd';
import type { StatisticProps } from 'antd';

interface StatCardProps extends StatisticProps {
  loading?: boolean;
  onClick?: () => void;
}

const StatCard: React.FC<StatCardProps> = ({ loading, onClick, ...props }) => {
  return (
    <Card
      loading={loading}
      hoverable={!!onClick}
      onClick={onClick}
      style={onClick ? { cursor: 'pointer' } : undefined}
    >
      <Statistic {...props} />
    </Card>
  );
};

export default StatCard;
