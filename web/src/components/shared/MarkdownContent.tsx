import { memo, type ReactElement, type ReactNode } from "react";
import ReactMarkdown, { type Components } from "react-markdown";
import remarkGfm from "remark-gfm";
import hljs from "highlight.js/lib/common";
import dockerfile from "highlight.js/lib/languages/dockerfile";
import "./markdown.css";

// common 集不含 dockerfile，单独注册
hljs.registerLanguage("dockerfile", dockerfile);

interface MarkdownContentProps {
  /** 待渲染的 markdown 文本 */
  content: string;
  /** 追加到容器的额外类名（容器始终带 md-body 基础类） */
  className?: string;
}

// 常见别名 → highlight.js 标准语言名
const LANG_ALIASES: Record<string, string> = {
  js: "javascript",
  jsx: "javascript",
  ts: "typescript",
  tsx: "typescript",
  py: "python",
  sh: "bash",
  shell: "bash",
  zsh: "bash",
  yml: "yaml",
  md: "markdown",
  text: "plaintext",
  plain: "plaintext",
  "c++": "cpp",
  golang: "go",
};

/** 归一化语言名；无法识别时原样返回（会按纯文本展示） */
function normalizeLang(lang: string): string {
  const l = lang.toLowerCase();
  return LANG_ALIASES[l] ?? l;
}

/** 高亮代码并返回安全 HTML；未知语言仅做转义（不产生 token 着色） */
function highlightCode(raw: string, lang: string): string {
  const language = normalizeLang(lang);
  if (hljs.getLanguage(language)) {
    return hljs.highlight(raw, { language, ignoreIllegals: true }).value;
  }
  return hljs.highlight(raw, { language: "plaintext", ignoreIllegals: true }).value;
}

// 自定义各元素渲染，统一挂上样式类
const components: Components = {
  // 块级代码（含无语言的 ``` 代码块）统一由 pre 组件渲染，保证始终有 <pre>
  // 容器、保留换行并应用块级样式；这里从 children（code 元素）提取语言与文本。
  pre({ children }) {
    const codeEl = children as
      | ReactElement<{ className?: string; children?: ReactNode }>
      | undefined;
    const className = codeEl?.props?.className ?? "";
    const lang = /language-(\w+)/.exec(className)?.[1];
    const raw = String(codeEl?.props?.children ?? "").replace(/\n$/, "");
    return (
      <pre
        className="md-pre hljs"
        data-language={lang}
        dangerouslySetInnerHTML={{ __html: highlightCode(raw, lang ?? "plaintext") }}
      />
    );
  },
  code({ className, children }) {
    const lang = /language-(\w+)/.exec(className ?? "")?.[1];
    if (lang) {
      // 块级代码：保留 language-xxx className，供 pre 组件识别语言
      return <code className={className}>{children}</code>;
    }
    // 行内代码
    return <code className="md-code-inline">{children}</code>;
  },
  // 表格套一层横向滚动容器（markdown 表格列多时防溢出）
  table({ children }) {
    return (
      <div className="md-table-wrap">
        <table className="md-table">{children}</table>
      </div>
    );
  },
  a({ href, children }) {
    const external = /^https?:\/\//.test(href ?? "");
    return (
      <a
        className="md-link"
        href={href}
        target={external ? "_blank" : undefined}
        rel={external ? "noopener noreferrer" : undefined}
      >
        {children}
      </a>
    );
  },
  img({ src, alt }) {
    return <img className="md-image" src={src} alt={alt ?? ""} loading="lazy" />;
  },
};

/**
 * MarkdownContent 将 markdown 文本渲染为带完整样式的富文本。
 * - 支持 GFM 扩展（表格、任务列表、删除线、自动链接）
 * - 默认不渲染原始 HTML（XSS 安全）
 * - 代码块自动显示语言标签；外部链接自动新窗口打开
 */
export default memo(function MarkdownContent({ content, className }: MarkdownContentProps) {
  return (
    <div className={className ? `md-body ${className}` : "md-body"}>
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={components}>
        {content}
      </ReactMarkdown>
    </div>
  );
});
