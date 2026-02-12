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

function extractBearerToken(authorization: unknown): string | null {
  if (typeof authorization !== 'string') {
    return null;
  }
  const prefix = 'Bearer ';
  if (!authorization.startsWith(prefix)) {
    return null;
  }
  return authorization.slice(prefix.length);
}

function getRequestToken(error: any): string | null {
  const headers = error?.config?.headers;
  const authorization =
    headers?.Authorization ??
    headers?.authorization ??
    (typeof headers?.get === 'function' ? headers.get('Authorization') : undefined);
  return extractBearerToken(authorization);
}

// 请求拦截器：无 token 时直接拦截（登录接口除外）
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    (config as any).__authToken = token;

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

    // 登录接口返回 401 属于账号密码错误，不触发过期跳转
    if (error.response?.status === 401 && error.config?.url?.includes('/auth/login')) {
      const errorMessage = error.response?.data?.message || error.message || '登录失败';
      return Promise.reject({ code: 401, message: errorMessage, details: error.response?.data });
    }

    // 401：清凭证 + 跳登录页（只跳一次）
    if (error.response?.status === 401) {
      const latestToken = localStorage.getItem('token');
      const requestToken = error?.config?.__authToken ?? getRequestToken(error);

      // 旧 token 的延迟 401，不应覆盖新登录态
      if (latestToken && latestToken !== requestToken) {
        return Promise.reject({ code: 401, message: '请求已过期，请重试' });
      }

      localStorage.removeItem('token');
      localStorage.removeItem('username');

      const isOnLoginPage = window.location.pathname.startsWith('/login');
      if (!isRedirectingToLogin && !isOnLoginPage) {
        isRedirectingToLogin = true;
        const currentPath = `${window.location.pathname}${window.location.search}`;
        window.location.replace(`/login?redirect=${encodeURIComponent(currentPath || '/')}`);
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
