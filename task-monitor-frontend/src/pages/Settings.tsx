import React, { useEffect, useState } from 'react';
import { Collapse, Form, Input, InputNumber, Switch, Button, message, Spin, Table, Modal, Popconfirm, Space } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { configApi, type LLMConfig } from '@/api/config';
import { authApi, type User } from '@/api/auth';
import { useAuthStore } from '@/stores/useAuthStore';

const Settings: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // 用户管理状态
  const [users, setUsers] = useState<User[]>([]);
  const [usersLoading, setUsersLoading] = useState(false);
  const [addModalOpen, setAddModalOpen] = useState(false);
  const [pwdModalOpen, setPwdModalOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [addForm] = Form.useForm();
  const [pwdForm] = Form.useForm();
  const currentUsername = useAuthStore((s) => s.username);

  useEffect(() => {
    loadConfig();
    loadUsers();
  }, []);

  const loadConfig = async () => {
    try {
      setLoading(true);
      const cfg = await configApi.getLLMConfig();
      form.setFieldsValue(cfg);
    } catch (err: any) {
      message.error('加载配置失败: ' + (err.message || '未知错误'));
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (values: LLMConfig) => {
    try {
      setSaving(true);
      const result = await configApi.updateLLMConfig(values);
      form.setFieldsValue(result);
      message.success('配置已保存');
    } catch (err: any) {
      message.error('保存失败: ' + (err.message || '未知错误'));
    } finally {
      setSaving(false);
    }
  };

  const loadUsers = async () => {
    try {
      setUsersLoading(true);
      const list = await authApi.listUsers();
      setUsers(list);
    } catch (err: any) {
      message.error('加载用户列表失败: ' + (err.message || '未知错误'));
    } finally {
      setUsersLoading(false);
    }
  };

  const handleAddUser = async () => {
    try {
      const values = await addForm.validateFields();
      await authApi.createUser({ username: values.username, password: values.password });
      message.success('用户创建成功');
      setAddModalOpen(false);
      addForm.resetFields();
      loadUsers();
    } catch (err: any) {
      if (err.message) {
        message.error('创建失败: ' + err.message);
      }
    }
  };

  const handleChangePassword = async () => {
    try {
      const values = await pwdForm.validateFields();
      if (!editingUser) return;
      await authApi.changePassword(editingUser.id, { password: values.newPassword });
      message.success('密码修改成功');
      setPwdModalOpen(false);
      pwdForm.resetFields();
      setEditingUser(null);
    } catch (err: any) {
      if (err.message) {
        message.error('修改失败: ' + err.message);
      }
    }
  };

  const handleDeleteUser = async (userId: number) => {
    try {
      await authApi.deleteUser(userId);
      message.success('用户已删除');
      loadUsers();
    } catch (err: any) {
      message.error('删除失败: ' + (err.message || '未知错误'));
    }
  };

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>系统设置</h2>
      <Collapse
        style={{ maxWidth: 600 }}
        items={[
          {
            key: 'llm',
            label: 'LLM 智能分析配置',
            children: (
              <Spin spinning={loading}>
                <Form
                  form={form}
                  layout="vertical"
                  onFinish={handleSave}
                  initialValues={{ enabled: false, timeout: 60 }}
                >
                  <Form.Item name="enabled" label="启用 LLM 分析" valuePropName="checked">
                    <Switch />
                  </Form.Item>
                  <Form.Item name="endpoint" label="接口地址" rules={[{ required: true, message: '请输入接口地址' }]}>
                    <Input placeholder="http://localhost:8000/v1" />
                  </Form.Item>
                  <Form.Item name="api_key" label="API Key">
                    <Input.Password placeholder="输入新的 API Key（留空则不修改）" />
                  </Form.Item>
                  <Form.Item name="model" label="模型名称" rules={[{ required: true, message: '请输入模型名称' }]}>
                    <Input placeholder="qwen2.5" />
                  </Form.Item>
                  <Form.Item name="timeout" label="超时时间（秒）" rules={[{ required: true, message: '请输入超时时间' }]}>
                    <InputNumber min={1} max={300} style={{ width: '100%' }} />
                  </Form.Item>
                  <Form.Item>
                    <Button type="primary" htmlType="submit" loading={saving}>保存配置</Button>
                  </Form.Item>
                </Form>
              </Spin>
            ),
          },
          {
            key: 'users',
            label: '用户管理',
            extra: (
              <Button
                type="primary"
                size="small"
                icon={<PlusOutlined />}
                onClick={(e) => { e.stopPropagation(); setAddModalOpen(true); }}
              >
                添加用户
              </Button>
            ),
            children: (
              <Table
                dataSource={users}
                rowKey="id"
                loading={usersLoading}
                pagination={false}
                columns={[
                  { title: '用户名', dataIndex: 'username', key: 'username' },
                  {
                    title: '创建时间',
                    dataIndex: 'createdAt',
                    key: 'createdAt',
                    render: (v: string) => new Date(v).toLocaleString(),
                  },
                  {
                    title: '操作',
                    key: 'action',
                    render: (_: unknown, record: User) => (
                      <Space>
                        <Button size="small" onClick={() => { setEditingUser(record); setPwdModalOpen(true); }}>
                          修改密码
                        </Button>
                        {record.username !== currentUsername && (
                          <Popconfirm title="确定删除该用户？" onConfirm={() => handleDeleteUser(record.id)}>
                            <Button size="small" danger>删除</Button>
                          </Popconfirm>
                        )}
                      </Space>
                    ),
                  },
                ]}
              />
            ),
          },
        ]}
      />

      <Modal
        title="添加用户"
        open={addModalOpen}
        onOk={handleAddUser}
        onCancel={() => { setAddModalOpen(false); addForm.resetFields(); }}
      >
        <Form form={addForm} layout="vertical">
          <Form.Item
            name="username"
            label="用户名"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input placeholder="请输入用户名" />
          </Form.Item>
          <Form.Item
            name="password"
            label="密码"
            rules={[{ required: true, message: '请输入密码' }, { min: 6, message: '密码至少6位' }]}
          >
            <Input.Password placeholder="请输入密码" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`修改密码 - ${editingUser?.username || ''}`}
        open={pwdModalOpen}
        onOk={handleChangePassword}
        onCancel={() => { setPwdModalOpen(false); pwdForm.resetFields(); setEditingUser(null); }}
      >
        <Form form={pwdForm} layout="vertical">
          <Form.Item
            name="newPassword"
            label="新密码"
            rules={[{ required: true, message: '请输入新密码' }, { min: 6, message: '密码至少6位' }]}
          >
            <Input.Password placeholder="请输入新密码" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Settings;
