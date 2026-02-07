import apiClient from './client';

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  username: string;
}

export interface User {
  id: number;
  username: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateUserRequest {
  username: string;
  password: string;
}

export interface ChangePasswordRequest {
  password: string;
}

export const authApi = {
  login: async (data: LoginRequest): Promise<LoginResponse> => {
    return apiClient.post('/auth/login', data);
  },
  getCurrentUser: async (): Promise<User> => {
    return apiClient.get('/auth/me');
  },
  listUsers: async (): Promise<User[]> => {
    return apiClient.get('/users');
  },
  createUser: async (data: CreateUserRequest): Promise<User> => {
    return apiClient.post('/users', data);
  },
  changePassword: async (id: number, data: ChangePasswordRequest): Promise<void> => {
    return apiClient.put(`/users/${id}/password`, data);
  },
  deleteUser: async (id: number): Promise<void> => {
    return apiClient.delete(`/users/${id}`);
  },
};
