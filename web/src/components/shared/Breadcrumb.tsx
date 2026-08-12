import { Fragment, type ReactNode } from "react";
import { Link } from "react-router-dom";

export interface BreadcrumbItem {
  label: string;
  /** 有 to 时渲染为链接，无 to 时渲染为当前页文本 */
  to?: string;
}

interface BreadcrumbProps {
  items: BreadcrumbItem[];
}

const chevron: ReactNode = (
  <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
    <path d="M5 3l4 4-4 4" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
  </svg>
);

/** 统一面包屑导航 */
export default function Breadcrumb({ items }: BreadcrumbProps) {
  return (
    <nav
      aria-label="breadcrumb"
      style={{
        display: "flex",
        alignItems: "center",
        gap: 8,
        fontSize: 13,
        color: "var(--color-text-muted)",
        marginBottom: 28,
      }}
    >
      {items.map((item, index) => (
        <Fragment key={index}>
          {index > 0 && chevron}
          {item.to ? (
            <Link to={item.to} style={{ color: "var(--color-text-secondary)", textDecoration: "none" }}>
              {item.label}
            </Link>
          ) : (
            <span style={{ color: "var(--color-text)", fontWeight: 500 }}>{item.label}</span>
          )}
        </Fragment>
      ))}
    </nav>
  );
}
