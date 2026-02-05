import { create } from 'zustand';
import type { JobStatus, JobType } from '@/types/job';

interface FilterState {
  // 作业筛选
  jobStatus: JobStatus[];
  jobType: JobType[];
  jobFramework: string[];
  jobNodeId: string | null;

  // 节点筛选
  nodeStatus: string | null;

  // 操作方法
  setJobStatus: (status: JobStatus[]) => void;
  setJobType: (type: JobType[]) => void;
  setJobFramework: (framework: string[]) => void;
  setJobNodeId: (nodeId: string | null) => void;
  setNodeStatus: (status: string | null) => void;
  resetJobFilters: () => void;
  resetNodeFilters: () => void;
}

export const useFilterStore = create<FilterState>((set) => ({
  // 初始状态
  jobStatus: [],
  jobType: [],
  jobFramework: [],
  jobNodeId: null,
  nodeStatus: null,

  // 操作方法
  setJobStatus: (status) => set({ jobStatus: status }),
  setJobType: (type) => set({ jobType: type }),
  setJobFramework: (framework) => set({ jobFramework: framework }),
  setJobNodeId: (nodeId) => set({ jobNodeId: nodeId }),
  setNodeStatus: (status) => set({ nodeStatus: status }),

  resetJobFilters: () => set({
    jobStatus: [],
    jobType: [],
    jobFramework: [],
    jobNodeId: null,
  }),

  resetNodeFilters: () => set({
    nodeStatus: null,
  }),
}));
