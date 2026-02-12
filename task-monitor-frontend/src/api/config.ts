import apiClient from './client';

export interface LLMModelConfig {
  id: string;
  name: string;
  endpoint: string;
  api_key: string;
  model: string;
  timeout: number;
  enabled: boolean;
}

export interface LLMConfig {
  enabled: boolean;
  endpoint?: string;
  api_key?: string;
  model?: string;
  timeout?: number;
  batch_concurrency?: number;
  default_model_id?: string;
  models?: LLMModelConfig[];
}

export const configApi = {
  getLLMConfig: async (): Promise<LLMConfig> => {
    return apiClient.get('/config/llm');
  },

  updateLLMConfig: async (data: Partial<LLMConfig>): Promise<LLMConfig> => {
    return apiClient.put('/config/llm', data);
  },
};
