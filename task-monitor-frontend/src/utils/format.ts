import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import 'dayjs/locale/zh-cn';

dayjs.extend(relativeTime);
dayjs.locale('zh-cn');

/**
 * 格式化时间戳为可读字符串
 */
export const formatTimestamp = (timestamp: number | null | undefined, format = 'YYYY-MM-DD HH:mm:ss'): string => {
  if (!timestamp) return '-';
  return dayjs(timestamp).format(format);
};

/**
 * 格式化为相对时间
 */
export const formatRelativeTime = (timestamp: number | null | undefined): string => {
  if (!timestamp) return '-';
  return dayjs(timestamp).fromNow();
};

/**
 * 格式化ISO时间字符串
 */
export const formatISOTime = (isoString: string | null | undefined, format = 'YYYY-MM-DD HH:mm:ss'): string => {
  if (!isoString) return '-';
  return dayjs(isoString).format(format);
};

/**
 * 格式化数字，添加千分位分隔符
 */
export const formatNumber = (num: number | null | undefined): string => {
  if (num === null || num === undefined) return '-';
  return num.toLocaleString('zh-CN');
};

/**
 * 格式化百分比
 */
export const formatPercent = (value: number | null | undefined, decimals = 2): string => {
  if (value === null || value === undefined) return '-';
  return `${(value * 100).toFixed(decimals)}%`;
};

/**
 * 格式化字节大小
 */
export const formatBytes = (bytes: number | null | undefined, decimals = 2): string => {
  if (bytes === null || bytes === undefined) return '-';
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(decimals))} ${sizes[i]}`;
};
