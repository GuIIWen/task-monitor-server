import React from 'react';
import { Card } from 'antd';
import { Area } from '@ant-design/charts';

interface TrendData {
  time: string;
  value: number;
  category: string;
}

interface TrendChartProps {
  data: TrendData[];
  title?: string;
  loading?: boolean;
}

const TrendChart: React.FC<TrendChartProps> = ({
  data,
  title = '趋势图',
  loading
}) => {
  const config = {
    data,
    xField: 'time',
    yField: 'value',
    seriesField: 'category',
    smooth: true,
    areaStyle: {
      fillOpacity: 0.3,
    },
  };

  return (
    <Card title={title} loading={loading}>
      <Area {...config} />
    </Card>
  );
};

export default TrendChart;
