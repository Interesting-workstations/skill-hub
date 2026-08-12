import { forwardRef, type CSSProperties, type ReactNode } from "react";

interface PageContainerProps {
  children: ReactNode;
  /** 内容最大宽度（px），默认 1280 */
  maxWidth?: number;
  /** 内边距，默认 "40px 24px 60px" */
  padding?: string;
  /** 附加样式（合并进默认样式之后） */
  style?: CSSProperties;
}

/** 统一页面内容容器：水平居中 + 最大宽度 + 统一内边距 */
const PageContainer = forwardRef<HTMLDivElement, PageContainerProps>(
  ({ children, maxWidth = 1280, padding = "40px 24px 60px", style }, ref) => (
    <div
      ref={ref}
      style={{
        maxWidth,
        margin: "0 auto",
        padding,
        ...style,
      }}
    >
      {children}
    </div>
  )
);

export default PageContainer;
