/** 通用表格：工具栏 + 表头 + 分页，最核心的后台组件。 */

import type { ReactNode } from "react";

export interface Column<T> {
  key: string;
  title: string;
  width?: string;
  render?: (row: T) => ReactNode;
}

interface Props<T> {
  columns: Column<T>[];
  data: T[];
  rowKey: (row: T) => string;
  toolbar?: ReactNode;
  page?: number;
  pageSize?: number;
  total?: number;
  onPageChange?: (page: number) => void;
  loading?: boolean;
  emptyText?: string;
}

export default function AppTable<T>({
  columns,
  data,
  rowKey,
  toolbar,
  page = 1,
  pageSize = 10,
  total,
  onPageChange,
  loading,
  emptyText = "暂无数据",
}: Props<T>) {
  const totalCount = total ?? data.length;
  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));

  return (
    <div className="table-wrap">
      {toolbar && <div className="table-toolbar">{toolbar}</div>}
      <table className="table">
        <thead>
          <tr>
            {columns.map((c) => (
              <th key={c.key} style={c.width ? { width: c.width } : undefined}>
                {c.title}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.length === 0 ? (
            <tr>
              <td colSpan={columns.length}>
                <div className="empty">{loading ? <span className="spin" /> : null}{emptyText}</div>
              </td>
            </tr>
          ) : (
            data.map((row) => (
              <tr key={rowKey(row)}>
                {columns.map((c) => (
                  <td key={c.key}>{c.render ? c.render(row) : String((row as Record<string, unknown>)[c.key] ?? "")}</td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
      {totalPages > 1 && (
        <div className="table-footer">
          <span>共 {totalCount} 条</span>
          <div className="pagination">
            <button disabled={page <= 1} onClick={() => onPageChange?.(page - 1)}>‹</button>
            {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
              <button key={p} className={p === page ? "active" : ""} onClick={() => onPageChange?.(p)}>
                {p}
              </button>
            ))}
            <button disabled={page >= totalPages} onClick={() => onPageChange?.(page + 1)}>›</button>
          </div>
        </div>
      )}
    </div>
  );
}
