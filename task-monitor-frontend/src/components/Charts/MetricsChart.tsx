import React from 'react';
import { Card } from 'antd';
import { Line } from '@ant-design/charts';

interface MetricsData {
  timestamp: number;
  value: number;
  metric: string;
}

interface MetricsChartProps {
  data: MetricsData[];
  title?: string;
  loading?: boolean;
}

const MetricsChart: React.FC<MetricsChartProps> = ({
  data,
  title = '指标趋势',
  loading
}) => {
  const config = {
    data,
    xField: 'timestamp',
    yField: 'value',
    seriesField: 'metric',
    smooth: true,
    animation: {
      appear: {
        animation: 'path-in',
        duration: 1000,
      },
    },
    xAxis: {
      type: 'time',
      label: {
        formatter: (v: string) => {
          const date = new Date(Number(v));
          return `${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`;
        },
      },
    },
  };

  return (
    <Card title={title} loading={loading}>
      <Line {...config} />
    </Card>
  );
};

export default MetricsChart;
