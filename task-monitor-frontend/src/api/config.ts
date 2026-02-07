import apiClient from './client';

export interface LLMConfig {
  enabled: boolean;
  endpoint: string;
  api_key: string;
  model: string;
  timeout: number;
}

export const configApi = {
  getLLMConfig: async (): Promise<LLMConfig> => {
    return apiClient.get('/config/llm');
  },

  updateLLMConfig: async (data: Partial<LLMConfig>): Promise<LLMConfig> => {
    return apiClient.put('/config/llm', data);
  },
};
