import axios from 'axios';

// 创建axios实例
const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
  paramsSerializer: {
    // 数组参数使用 repeat 格式: status=running&status=failed
    indexes: null,
  },
});

// 请求拦截器
apiClient.interceptors.request.use(
  (config) => {
    // 添加认证token（如果需要）
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器
apiClient.interceptors.response.use(
  (response) => {
    if (response.config.responseType === 'blob' || response.config.responseType === 'arraybuffer') {
      return response.data;
    }
    // API返回格式为 {code, message, data}，解包到data字段
    return response.data.data;
  },
  (error) => {
    // 统一错误处理
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('username');
    }

    const errorMessage = error.response?.data?.message || error.message || '请求失败';
    console.error('API Error:', errorMessage);

    return Promise.reject({
      code: error.response?.status || 500,
      message: errorMessage,
      details: error.response?.data,
    });
  }
);

export default apiClient;
