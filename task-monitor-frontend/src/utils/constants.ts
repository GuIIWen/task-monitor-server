/**
 * 作业类型映射
 */
export const JOB_TYPE_MAP = {
  training: '训练',
  inference: '推理',
  testing: '测试',
  unknown: '未知',
} as const;

/**
 * 作业状态映射
 */
export const JOB_STATUS_MAP = {
  running: '运行中',
  completed: '已完成',
  failed: '失败',
  stopped: '已停止',
  lost: '丢失',
} as const;

/**
 * 节点状态映射
 */
export const NODE_STATUS_MAP = {
  active: '活跃',
  inactive: '离线',
  error: '错误',
} as const;

/**
 * 框架映射
 */
export const FRAMEWORK_MAP = {
  pytorch: 'PyTorch',
  tensorflow: 'TensorFlow',
  mindspore: 'MindSpore',
  unknown: '未知',
} as const;

/**
 * 默认分页配置
 */
export const DEFAULT_PAGE_SIZE = 20;
export const PAGE_SIZE_OPTIONS = [10, 20, 50, 100];

/**
 * 刷新间隔（毫秒）
 */
export const REFRESH_INTERVAL = 30000; // 30秒
