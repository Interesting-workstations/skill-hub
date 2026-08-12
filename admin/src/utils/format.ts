/** 通用格式化工具。 */

/** 数字千分位 */
export function formatNumber(n: number): string {
  return n.toLocaleString("zh-CN");
}

/** 时间（仅 HH:mm） */
export function formatTime(iso: string): string {
  return iso.slice(11, 16);
}

/** 相对时间（模拟） */
export function timeAgo(iso: string): string {
  return iso;
}
