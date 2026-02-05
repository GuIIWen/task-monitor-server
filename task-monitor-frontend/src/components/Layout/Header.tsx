import React from 'react';
import { Layout, Typography } from 'antd';
import { DashboardOutlined } from '@ant-design/icons';

const { Header: AntHeader } = Layout;
const { Title } = Typography;

const Header: React.FC = () => {
  return (
    <AntHeader style={{
      background: '#001529',
      padding: '0 24px',
      display: 'flex',
      alignItems: 'center',
      gap: '12px'
    }}>
      <DashboardOutlined style={{ fontSize: '24px', color: '#fff' }} />
      <Title level={3} style={{ color: '#fff', margin: 0 }}>
        Task Monitor
      </Title>
    </AntHeader>
  );
};

export default Header;
