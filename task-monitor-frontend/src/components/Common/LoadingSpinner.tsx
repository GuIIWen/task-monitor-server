import React from 'react';
import { Spin } from 'antd';

interface LoadingSpinnerProps {
  tip?: string;
  size?: 'small' | 'default' | 'large';
}

const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  tip = '加载中...',
  size = 'large'
}) => {
  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '200px'
    }}>
      <Spin size={size} tip={tip} />
    </div>
  );
};

export default LoadingSpinner;
