export interface Skill {
  id: string;
  name: string;
  author: string;
  authorAvatar?: string;
  description: string;
  tags: string[];
  category: string;
  downloadUrl: string;
  isOfficial?: boolean;
  isFeatured?: boolean;
}

export interface Author {
  name: string;
  avatar: string;
  skillCount: number;
  slug: string;
}

export interface SkillCategory {
  name: string;
  slug: string;
  count: number;
  skills: Skill[];
}

export const authors: Author[] = [
  { name: "Anthropic", avatar: "🅰️", skillCount: 532, slug: "anthropic" },
  { name: "OpenAI", avatar: "🤖", skillCount: 367, slug: "openai" },
  { name: "GitHub", avatar: "🐙", skillCount: 449, slug: "github" },
  { name: "Microsoft", avatar: "🪟", skillCount: 861, slug: "microsoft" },
  { name: "Cloudflare", avatar: "☁️", skillCount: 79, slug: "cloudflare" },
  { name: "Figma", avatar: "🎨", skillCount: 20, slug: "figma" },
  { name: "Vercel", avatar: "▲", skillCount: 282, slug: "vercel" },
  { name: "Google Workspace", avatar: "🇬", skillCount: 99, slug: "google" },
  { name: "Notion", avatar: "📝", skillCount: 23, slug: "notion" },
  { name: "Stripe", avatar: "💳", skillCount: 7, slug: "stripe" },
];

export const featuredSkills: Skill[] = [
  {
    id: "frontend-design",
    name: "Frontend Design",
    author: "Anthropic",
    description: "生成独特且达到生产级水准的前端界面，避免千篇一律的AI美学风格。",
    tags: ["development", "featured", "official"],
    category: "featured",
    downloadUrl: "#",
    isOfficial: true,
    isFeatured: true,
  },
  {
    id: "notebooklm-skill",
    name: "NotebookLM Skill",
    author: "pleaseprompto",
    description: "让Claude Code直接与NotebookLM对话，仅基于你上传的文档提供有据可查的答案。",
    tags: ["productivity", "featured"],
    category: "featured",
    downloadUrl: "#",
    isFeatured: true,
  },
  {
    id: "webapp-testing",
    name: "WebApp Testing",
    author: "Microsoft",
    description: "使用 Playwright 自动化测试 Web 应用程序，支持跨浏览器端到端测试。",
    tags: ["testing", "featured", "official"],
    category: "featured",
    downloadUrl: "#",
    isOfficial: true,
    isFeatured: true,
  },
  {
    id: "pdf-chat",
    name: "PDF Chat",
    author: "OpenAI",
    description: "与 PDF 文档对话，提取关键信息，支持多文档交叉查询和智能摘要。",
    tags: ["document", "featured", "official"],
    category: "featured",
    downloadUrl: "#",
    isOfficial: true,
    isFeatured: true,
  },
];

export const skillCategories: SkillCategory[] = [
  {
    name: "文档技能",
    slug: "document",
    count: 79,
    skills: [
      {
        id: "ai-research-reproduction",
        name: "ai-research-reproduction",
        author: "lllllllama",
        description: "RigorPilot 复现模式编排器，用于README优先的深度学习仓库复现。",
        tags: ["research", "development", "document"],
        category: "document",
        downloadUrl: "#",
      },
      {
        id: "caveman-compress",
        name: "caveman-compress",
        author: "juliusbrussee",
        description: "将自然语言记忆文件压缩为穴居人格式以节省输入令牌，保留所有技术内容。",
        tags: ["development", "document"],
        category: "document",
        downloadUrl: "#",
      },
    ],
  },
  {
    name: "浏览器自动化技能",
    slug: "browser-automation",
    count: 24,
    skills: [
      {
        id: "agent-browser-qu",
        name: "agent-browser",
        author: "qu-skills",
        description: "通过inference.sh为AI代理提供浏览器自动化功能：网页抓取、表单填写、点击、输入等。",
        tags: ["browser-automation", "web-scraping", "testing"],
        category: "browser-automation",
        downloadUrl: "#",
      },
      {
        id: "agent-browser-hcf",
        name: "agent-browser",
        author: "halt-catch-fire",
        description: "为AI代理提供浏览器自动化功能，支持导航、交互、截图、录制视频。",
        tags: ["browser-automation", "web-scraping", "testing"],
        category: "browser-automation",
        downloadUrl: "#",
      },
    ],
  },
  {
    name: "数据库技能",
    slug: "database",
    count: 27,
    skills: [
      {
        id: "azure-storage",
        name: "azure-storage",
        author: "Microsoft",
        description: "统一访问Azure blob存储、文件共享、队列、表格和数据湖服务。",
        tags: ["official", "aws", "database"],
        category: "database",
        downloadUrl: "#",
        isOfficial: true,
      },
      {
        id: "convex",
        name: "convex",
        author: "get-convex",
        description: "将常规Convex请求路由到正确的项目技能。",
        tags: ["development", "database"],
        category: "database",
        downloadUrl: "#",
      },
    ],
  },
  {
    name: "开发技能",
    slug: "development",
    count: 347,
    skills: [
      {
        id: "accessibility",
        name: "accessibility",
        author: "addyosmani",
        description: "根据WCAG 2.2指南审计并改进网页无障碍性，支持屏幕阅读器和键盘导航。",
        tags: ["development", "testing", "code-review"],
        category: "development",
        downloadUrl: "#",
      },
      {
        id: "agentix-ceo",
        name: "agentix-ceo",
        author: "agentix-cloud",
        description: "管理你的团队——创建角色、分配任务、启动工作线程并监控进度。",
        tags: ["project-management", "development", "communication"],
        category: "development",
        downloadUrl: "#",
      },
    ],
  },
  {
    name: "创意技能",
    slug: "creative",
    count: 250,
    skills: [
      {
        id: "ace-step-as",
        name: "ace-step",
        author: "agentspace-so",
        description: "使用ACE Step音乐基础模型生成、修复和扩展音乐，支持多语言歌词。",
        tags: ["creative", "audio", "media"],
        category: "creative",
        downloadUrl: "#",
      },
      {
        id: "ace-step-rc",
        name: "ace-step",
        author: "runcomfy-com",
        description: "通过RunComfy使用ACE Step生成音乐，标签驱动的作曲方式。",
        tags: ["creative", "audio", "api"],
        category: "creative",
        downloadUrl: "#",
      },
    ],
  },
  {
    name: "媒体技能",
    slug: "media",
    count: 117,
    skills: [
      {
        id: "ace-step-media-as",
        name: "ace-step",
        author: "agentspace-so",
        description: "使用ACE Step音乐基础模型生成、修复和扩展音乐，支持多语言歌词。",
        tags: ["creative", "audio", "media"],
        category: "media",
        downloadUrl: "#",
      },
      {
        id: "ace-step-media-da",
        name: "ace-step",
        author: "doany-ai",
        description: "通过RunComfy使用ACE Step生成音乐，标签驱动的作曲方式。",
        tags: ["creative", "media", "audio"],
        category: "media",
        downloadUrl: "#",
      },
    ],
  },
  {
    name: "生产力技能",
    slug: "productivity",
    count: 64,
    skills: [
      {
        id: "meeting-notes",
        name: "meeting-notes",
        author: "Anthropic",
        description: "自动生成会议摘要和行动项，支持多语言转录和智能分类。",
        tags: ["productivity", "official"],
        category: "productivity",
        downloadUrl: "#",
        isOfficial: true,
      },
      {
        id: "task-automation",
        name: "task-automation",
        author: "OpenAI",
        description: "自动化重复性工作任务，支持定时触发和条件执行。",
        tags: ["productivity", "official"],
        category: "productivity",
        downloadUrl: "#",
        isOfficial: true,
      },
    ],
  },
];
