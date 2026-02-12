import React, { useEffect, useMemo, useState } from "react";
import {
  Collapse,
  Form,
  Input,
  InputNumber,
  Switch,
  Button,
  message,
  Spin,
  Table,
  Modal,
  Popconfirm,
  Space,
  Select,
  Card,
  Typography,
} from "antd";
import { PlusOutlined, DeleteOutlined } from "@ant-design/icons";
import { configApi, type LLMConfig, type LLMModelConfig } from "@/api/config";
import { authApi, type User } from "@/api/auth";
import { useAuthStore } from "@/stores/useAuthStore";

function buildDefaultModel(seed: number): LLMModelConfig {
  return {
    id: `model-${seed}`,
    name: "",
    endpoint: "",
    api_key: "",
    model: "",
    timeout: 60,
    enabled: true,
  };
}

function normalizeLLMConfig(cfg: LLMConfig): LLMConfig {
  let models = cfg.models || [];
  if (!models.length && (cfg.endpoint || cfg.model || cfg.api_key)) {
    models = [
      {
        id: cfg.default_model_id || "default",
        name: "默认模型",
        endpoint: cfg.endpoint || "",
        api_key: cfg.api_key || "",
        model: cfg.model || "",
        timeout: cfg.timeout || 60,
        enabled: true,
      },
    ];
  }

  if (!models.length) {
    models = [buildDefaultModel(1)];
  }

  const defaultModel =
    models.find((m) => m.id === cfg.default_model_id) ||
    models.find((m) => m.enabled) ||
    models[0];

  return {
    enabled: cfg.enabled,
    batch_concurrency: cfg.batch_concurrency || 5,
    default_model_id: defaultModel?.id,
    models,
    endpoint: cfg.endpoint,
    api_key: cfg.api_key,
    model: cfg.model,
    timeout: cfg.timeout,
  };
}

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

  const watchedModels = Form.useWatch("models", form) as
    | LLMModelConfig[]
    | undefined;
  const modelOptions = useMemo(() => {
    return (watchedModels || []).map((m) => ({
      label: `${m.name || m.id}${m.enabled ? "" : "（已禁用）"}`,
      value: m.id,
    }));
  }, [watchedModels]);

  useEffect(() => {
    loadConfig();
    loadUsers();
  }, []);

  const loadConfig = async () => {
    try {
      setLoading(true);
      const cfg = await configApi.getLLMConfig();
      form.setFieldsValue(normalizeLLMConfig(cfg));
    } catch (err: any) {
      message.error("加载配置失败: " + (err.message || "未知错误"));
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async (values: LLMConfig) => {
    const normalized = normalizeLLMConfig(values);
    const models = normalized.models || [];

    if (!models.length) {
      message.error("请至少配置一个模型");
      return;
    }

    const idSet = new Set<string>();
    for (const model of models) {
      const id = (model.id || "").trim();
      if (!id) {
        message.error("模型ID不能为空");
        return;
      }
      if (idSet.has(id)) {
        message.error(`存在重复模型ID: ${id}`);
        return;
      }
      idSet.add(id);
    }

    const defaultID = (normalized.default_model_id || "").trim();
    const defaultModel = models.find((m) => m.id === defaultID);
    if (!defaultModel) {
      message.error("请选择默认模型");
      return;
    }
    if (!defaultModel.enabled) {
      message.error("默认模型必须为启用状态");
      return;
    }

    try {
      setSaving(true);
      const result = await configApi.updateLLMConfig(normalized);
      form.setFieldsValue(normalizeLLMConfig(result));
      message.success("配置已保存");
    } catch (err: any) {
      message.error("保存失败: " + (err.message || "未知错误"));
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
      message.error("加载用户列表失败: " + (err.message || "未知错误"));
    } finally {
      setUsersLoading(false);
    }
  };

  const handleAddUser = async () => {
    try {
      const values = await addForm.validateFields();
      await authApi.createUser({ username: values.username, password: values.password });
      message.success("用户创建成功");
      setAddModalOpen(false);
      addForm.resetFields();
      loadUsers();
    } catch (err: any) {
      if (err.message) {
        message.error("创建失败: " + err.message);
      }
    }
  };

  const handleChangePassword = async () => {
    try {
      const values = await pwdForm.validateFields();
      if (!editingUser) return;
      await authApi.changePassword(editingUser.id, { password: values.newPassword });
      message.success("密码修改成功");
      setPwdModalOpen(false);
      pwdForm.resetFields();
      setEditingUser(null);
    } catch (err: any) {
      if (err.message) {
        message.error("修改失败: " + err.message);
      }
    }
  };

  const handleDeleteUser = async (userId: number) => {
    try {
      await authApi.deleteUser(userId);
      message.success("用户已删除");
      loadUsers();
    } catch (err: any) {
      message.error("删除失败: " + (err.message || "未知错误"));
    }
  };

  return (
    <div>
      <h2 style={{ marginBottom: 24 }}>系统设置</h2>
      <Collapse
        style={{ maxWidth: 960 }}
        items={[
          {
            key: "llm",
            label: "LLM 智能分析配置",
            children: (
              <Spin spinning={loading}>
                <Form
                  form={form}
                  layout="vertical"
                  onFinish={handleSave}
                  initialValues={{
                    enabled: false,
                    batch_concurrency: 5,
                    models: [buildDefaultModel(1)],
                    default_model_id: "model-1",
                  }}
                >
                  <Space size={16} style={{ width: "100%", marginBottom: 8 }} wrap>
                    <Form.Item
                      name="enabled"
                      label="启用 LLM 分析"
                      valuePropName="checked"
                      style={{ marginBottom: 8 }}
                    >
                      <Switch />
                    </Form.Item>
                    <Form.Item
                      name="batch_concurrency"
                      label="批量分析并发数"
                      rules={[{ required: true, message: "请输入并发数" }]}
                      style={{ minWidth: 220, marginBottom: 8 }}
                    >
                      <InputNumber min={1} max={20} style={{ width: "100%" }} />
                    </Form.Item>
                    <Form.Item
                      name="default_model_id"
                      label="默认模型"
                      rules={[{ required: true, message: "请选择默认模型" }]}
                      style={{ minWidth: 300, marginBottom: 8 }}
                    >
                      <Select options={modelOptions} placeholder="请选择默认模型" />
                    </Form.Item>
                  </Space>

                  <Typography.Text
                    type="secondary"
                    style={{ display: "block", marginBottom: 12 }}
                  >
                    支持维护多个模型；任务详情页可按需选择其中任意启用模型做分析。
                  </Typography.Text>

                  <Form.List name="models">
                    {(fields, { add, remove }) => (
                      <Space direction="vertical" style={{ width: "100%" }} size={12}>
                        {fields.map((field, index) => (
                          <Card
                            key={field.key}
                            size="small"
                            title={`模型 ${index + 1}`}
                            extra={
                              fields.length > 1 ? (
                                <Button
                                  danger
                                  size="small"
                                  icon={<DeleteOutlined />}
                                  onClick={() => remove(field.name)}
                                >
                                  删除
                                </Button>
                              ) : null
                            }
                          >
                            <Space direction="vertical" style={{ width: "100%" }} size={8}>
                              <Space style={{ width: "100%" }} size={12} wrap>
                                <Form.Item
                                  {...field}
                                  name={[field.name, "id"]}
                                  label="模型ID"
                                  rules={[{ required: true, message: "请输入模型ID" }]}
                                  style={{ minWidth: 220, marginBottom: 8 }}
                                >
                                  <Input placeholder="如 qwen-max" />
                                </Form.Item>
                                <Form.Item
                                  {...field}
                                  name={[field.name, "name"]}
                                  label="显示名称"
                                  rules={[{ required: true, message: "请输入模型名称" }]}
                                  style={{ minWidth: 220, marginBottom: 8 }}
                                >
                                  <Input placeholder="如 Qwen Max" />
                                </Form.Item>
                                <Form.Item
                                  {...field}
                                  name={[field.name, "enabled"]}
                                  label="启用"
                                  valuePropName="checked"
                                  style={{ marginBottom: 8 }}
                                >
                                  <Switch />
                                </Form.Item>
                              </Space>

                              <Space style={{ width: "100%" }} size={12} wrap>
                                <Form.Item
                                  {...field}
                                  name={[field.name, "endpoint"]}
                                  label="接口地址"
                                  rules={[{ required: true, message: "请输入接口地址" }]}
                                  style={{ minWidth: 360, flex: 1, marginBottom: 8 }}
                                >
                                  <Input placeholder="http://localhost:8000/v1" />
                                </Form.Item>
                                <Form.Item
                                  {...field}
                                  name={[field.name, "model"]}
                                  label="模型名"
                                  rules={[{ required: true, message: "请输入模型名" }]}
                                  style={{ minWidth: 220, marginBottom: 8 }}
                                >
                                  <Input placeholder="qwen2.5" />
                                </Form.Item>
                              </Space>

                              <Space style={{ width: "100%" }} size={12} wrap>
                                <Form.Item
                                  {...field}
                                  name={[field.name, "api_key"]}
                                  label="API Key"
                                  style={{ minWidth: 360, flex: 1, marginBottom: 8 }}
                                >
                                  <Input.Password placeholder="留空或保留掩码表示不修改" />
                                </Form.Item>
                                <Form.Item
                                  {...field}
                                  name={[field.name, "timeout"]}
                                  label="超时（秒）"
                                  rules={[{ required: true, message: "请输入超时" }]}
                                  style={{ minWidth: 160, marginBottom: 8 }}
                                >
                                  <InputNumber min={1} style={{ width: "100%" }} />
                                </Form.Item>
                              </Space>
                            </Space>
                          </Card>
                        ))}

                        <Button
                          type="dashed"
                          icon={<PlusOutlined />}
                          onClick={() => add(buildDefaultModel(Date.now()))}
                          style={{ width: "100%" }}
                        >
                          新增模型
                        </Button>
                      </Space>
                    )}
                  </Form.List>

                  <Form.Item style={{ marginTop: 16 }}>
                    <Button type="primary" htmlType="submit" loading={saving}>
                      保存配置
                    </Button>
                  </Form.Item>
                </Form>
              </Spin>
            ),
          },
          {
            key: "users",
            label: "用户管理",
            extra: (
              <Button
                type="primary"
                size="small"
                icon={<PlusOutlined />}
                onClick={(e) => {
                  e.stopPropagation();
                  setAddModalOpen(true);
                }}
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
                  { title: "用户名", dataIndex: "username", key: "username" },
                  {
                    title: "创建时间",
                    dataIndex: "createdAt",
                    key: "createdAt",
                    render: (v: string) => new Date(v).toLocaleString(),
                  },
                  {
                    title: "操作",
                    key: "action",
                    render: (_: unknown, record: User) => (
                      <Space>
                        <Button
                          size="small"
                          onClick={() => {
                            setEditingUser(record);
                            setPwdModalOpen(true);
                          }}
                        >
                          修改密码
                        </Button>
                        {record.username !== currentUsername && (
                          <Popconfirm
                            title="确定删除该用户？"
                            onConfirm={() => handleDeleteUser(record.id)}
                          >
                            <Button size="small" danger>
                              删除
                            </Button>
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
        onCancel={() => {
          setAddModalOpen(false);
          addForm.resetFields();
        }}
      >
        <Form form={addForm} layout="vertical">
          <Form.Item
            name="username"
            label="用户名"
            rules={[{ required: true, message: "请输入用户名" }]}
          >
            <Input placeholder="请输入用户名" />
          </Form.Item>
          <Form.Item
            name="password"
            label="密码"
            rules={[
              { required: true, message: "请输入密码" },
              { min: 6, message: "密码至少6位" },
            ]}
          >
            <Input.Password placeholder="请输入密码" />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`修改密码 - ${editingUser?.username || ""}`}
        open={pwdModalOpen}
        onOk={handleChangePassword}
        onCancel={() => {
          setPwdModalOpen(false);
          pwdForm.resetFields();
          setEditingUser(null);
        }}
      >
        <Form form={pwdForm} layout="vertical">
          <Form.Item
            name="newPassword"
            label="新密码"
            rules={[
              { required: true, message: "请输入新密码" },
              { min: 6, message: "密码至少6位" },
            ]}
          >
            <Input.Password placeholder="请输入新密码" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Settings;
