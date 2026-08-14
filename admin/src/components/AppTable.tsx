/** 通用表格：工具栏 + 表头 + 分页 + 可选行选择，最核心的后台组件。 */

import { useEffect, useRef, type ReactNode } from "react";

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
  /** 启用行选择（批量操作），配合 selectedKeys / onSelectionChange */
  selectable?: boolean;
  selectedKeys?: string[];
  onSelectionChange?: (keys: string[]) => void;
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
  selectable = false,
  selectedKeys = [],
  onSelectionChange,
}: Props<T>) {
  const totalCount = total ?? data.length;
  const totalPages = Math.max(1, Math.ceil(totalCount / pageSize));
  const pageData = data.slice((page - 1) * pageSize, page * pageSize);
  const selectedSet = new Set(selectedKeys);
  const allChecked = pageData.length > 0 && pageData.every((d) => selectedSet.has(rowKey(d)));
  const someChecked = pageData.some((d) => selectedSet.has(rowKey(d)));
  const allRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (allRef.current) {
      allRef.current.indeterminate = someChecked && !allChecked;
    }
  }, [someChecked, allChecked]);

  const toggleAll = () => {
    const set = new Set(selectedKeys);
    if (allChecked) {
      pageData.forEach((d) => set.delete(rowKey(d)));
    } else {
      pageData.forEach((d) => set.add(rowKey(d)));
    }
    onSelectionChange?.(Array.from(set));
  };

  const toggle = (key: string) => {
    const set = new Set(selectedKeys);
    if (set.has(key)) set.delete(key);
    else set.add(key);
    onSelectionChange?.(Array.from(set));
  };

  return (
    <div className="table-wrap">
      {toolbar && <div className="table-toolbar">{toolbar}</div>}
      <table className="table">
        <thead>
          <tr>
            {selectable && (
              <th style={{ width: 40 }}>
                <input
                  ref={allRef}
                  type="checkbox"
                  checked={allChecked}
                  onChange={toggleAll}
                  aria-label="全选"
                />
              </th>
            )}
            {columns.map((c) => (
              <th key={c.key} style={c.width ? { width: c.width } : undefined}>
                {c.title}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {pageData.length === 0 ? (
            <tr>
              <td colSpan={columns.length + (selectable ? 1 : 0)}>
                <div className="empty">{loading ? <span className="spin" /> : null}{emptyText}</div>
              </td>
            </tr>
          ) : (
            pageData.map((row) => (
              <tr key={rowKey(row)}>
                {selectable && (
                  <td>
                    <input
                      type="checkbox"
                      checked={selectedSet.has(rowKey(row))}
                      onChange={() => toggle(rowKey(row))}
                      aria-label="选择"
                    />
                  </td>
                )}
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
