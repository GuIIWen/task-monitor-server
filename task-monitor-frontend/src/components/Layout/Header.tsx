import React from 'react';
import { Layout, Typography, Button, Space } from 'antd';
import { DashboardOutlined, LogoutOutlined, UserOutlined, LoginOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/useAuthStore';

const { Header: AntHeader } = Layout;
const { Title } = Typography;

const Header: React.FC = () => {
  const navigate = useNavigate();
  const { username, isAuthenticated, logout } = useAuthStore();

  const handleLogout = () => {
    logout();
    navigate('/login', { replace: true });
  };

  return (
    <AntHeader style={{
      background: '#001529',
      padding: '0 24px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
        <DashboardOutlined style={{ fontSize: '24px', color: '#fff' }} />
        <Title level={3} style={{ color: '#fff', margin: 0 }}>
          Task Monitor
        </Title>
      </div>
      {isAuthenticated ? (
        <Space>
          <UserOutlined style={{ color: '#fff' }} />
          <span style={{ color: '#fff' }}>{username}</span>
          <Button
            type="text"
            icon={<LogoutOutlined />}
            onClick={handleLogout}
            style={{ color: '#fff' }}
          >
            退出
          </Button>
        </Space>
      ) : (
        <Button
          type="text"
          icon={<LoginOutlined />}
          onClick={() => navigate('/login')}
          style={{ color: '#fff' }}
        >
          登录
        </Button>
      )}
    </AntHeader>
  );
};

export default Header;
