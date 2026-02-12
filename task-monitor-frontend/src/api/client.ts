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

// 防止多个 401 同时触发重复跳转
let isRedirectingToLogin = false;

// 请求拦截器：无 token 时直接拦截（登录接口除外）
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    } else if (!config.url?.includes('/auth/')) {
      return Promise.reject({ __cancelled: true, message: '未登录' });
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
    // 被请求拦截器取消的无 token 请求，静默拒绝
    if (error.__cancelled) {
      return Promise.reject({ code: 401, message: error.message });
    }

    // 401：清凭证 + 跳登录页（只跳一次）
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('username');
      if (!isRedirectingToLogin) {
        isRedirectingToLogin = true;
        const currentPath = window.location.pathname;
        window.location.href = `/login?redirect=${encodeURIComponent(currentPath)}`;
      }
      return Promise.reject({ code: 401, message: '登录已过期' });
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

/** 登录成功后重置跳转标记 */
export function resetAuthRedirectFlag() {
  isRedirectingToLogin = false;
}

export default apiClient;
