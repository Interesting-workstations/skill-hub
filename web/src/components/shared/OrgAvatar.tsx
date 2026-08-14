import { useState } from "react";
import { API_BASE_URL } from "../../services/api/client";

interface Props {
  /** 图片 URL（相对 /api/v1 或完整 URL）；加载失败时回退到 fallback */
  src?: string;
  /** 回退内容（通常为 emoji） */
  fallback: string;
  size?: number;
  className?: string;
  alt?: string;
}

/** 官方组织 logo：优先显示真实图片（后端代理的 GitHub 组织头像），加载失败回退 emoji。 */
export default function OrgAvatar({ src, fallback, size = 32, className, alt }: Props) {
  const [failed, setFailed] = useState(false);
  // 相对路径（/org-logo/xxx）拼接后端 API 地址
  const resolvedSrc = src && !src.startsWith("http") ? `${API_BASE_URL}${src}` : src;

  if (!resolvedSrc || failed) {
    return (
      <span
        className={className}
        style={{
          display: "inline-flex",
          alignItems: "center",
          justifyContent: "center",
          fontSize: Math.round(size * 0.6),
          width: size,
          height: size,
        }}
        aria-hidden="true"
      >
        {fallback}
      </span>
    );
  }

  return (
    <img
      src={resolvedSrc}
      alt={alt ?? ""}
      className={className}
      width={size}
      height={size}
      style={{ objectFit: "cover", flexShrink: 0 }}
      onError={() => setFailed(true)}
    />
  );
}
