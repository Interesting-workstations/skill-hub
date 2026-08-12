import type { Author, Skill, SkillCategory } from "./types";

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
    installCommand:
      "npx skills add https://github.com/anthropics/skills --skill frontend-design",
    githubUrl: "https://github.com/anthropics/skills",
    githubStars: "168.1k",
    license: "完整条款见 LICENSE.txt",
    content: [
      {
        heading: "Frontend Design",
        body: [
          "Approach this as the design lead at a small studio known for giving every client a visual identity that could not be mistaken for anyone else's. This client has already rejected proposals that felt templated, and is paying for a distinctive point of view: make deliberate, opinionated choices about palette, typography, and layout that are specific to this brief, and take one real aesthetic risk you can justify.",
        ],
      },
      {
        heading: "Ground it in the subject",
        body: [
          "If the brief does not pin down what the product or subject is, pin it yourself before designing: name one concrete subject, its audience, and the page's single job, and state your choice. The subject's own world, its materials, instruments, artifacts, and vernacular, is where distinctive choices come from.",
        ],
      },
      {
        heading: "Design principles",
        body: [
          "For web designs, the hero is a thesis. Open with the most characteristic thing in the subject's world, in whatever form makes sense for it: a headline, an image, an animation, a live demo, an interactive moment.",
          "Typography carries the personality of the page. Pair the display and body faces deliberately, not the same families you would reach for on any other project, and set a clear type scale with intentional weights, widths, and spacing.",
          "Structure is information. Structural devices, numbering, eyebrows, dividers, labels, should encode something true about the content, not decorate it.",
          "Leverage motion deliberately. Think about where and if animation can serve the subject: a page-load sequence, a scroll-triggered reveal, hover micro-interactions, ambient atmosphere.",
          "Match complexity to the vision. Maximalist directions need elaborate execution; minimal directions need precision in spacing, type, and detail.",
        ],
      },
      {
        heading: "Restraint and self-critique",
        body: [
          "Spend your boldness in one place. Let the signature element be the one memorable thing, keep everything around it quiet and disciplined, and cut any decoration that does not serve the brief.",
          "Build to a quality floor without announcing it: responsive down to mobile, visible keyboard focus, reduced motion respected. Critique your own work as you build.",
        ],
      },
    ],
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
    installCommand: "npx skills add https://github.com/pleaseprompto/skills --skill notebooklm-skill",
    githubUrl: "https://github.com/pleaseprompto/skills",
    githubStars: "2.3k",
    content: [
      {
        heading: "使用方法",
        body: [
          "让 Claude Code 直接与 NotebookLM 对话，仅基于你上传的文档提供有据可查的答案。无需离开你的开发环境，即可获取基于文档的智能分析。",
        ],
      },
      {
        heading: "功能特性",
        body: [
          "自动连接到你的 NotebookLM 实例，读取已上传的文档内容。",
          "基于文档内容提供准确、有据可查的回答，避免幻觉。",
          "支持多文档交叉查询，快速定位关键信息。",
          "保持对话上下文，支持连续追问和深入探讨。",
        ],
      },
    ],
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
    installCommand: "npx skills add https://github.com/microsoft/skills --skill webapp-testing",
    githubUrl: "https://github.com/microsoft/skills",
    githubStars: "861",
    content: [
      {
        heading: "概述",
        body: [
          "使用 Playwright 与本地 Web 应用交互和测试的工具包。支持验证前端功能、调试 UI 行为、捕获浏览器截图以及查看浏览器日志。",
        ],
      },
      {
        heading: "核心能力",
        body: [
          "自动化浏览器操作：点击、输入、导航、截图",
          "端到端测试脚本生成和执行",
          "跨浏览器兼容性测试（Chromium、Firefox、WebKit）",
          "浏览器控制台日志捕获与分析",
          "网络请求拦截与 Mock",
        ],
      },
    ],
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
    installCommand: "npx skills add https://github.com/openai/skills --skill pdf-chat",
    githubUrl: "https://github.com/openai/skills",
    githubStars: "367",
    content: [
      {
        heading: "功能说明",
        body: [
          "与 PDF 文档进行自然语言对话，快速提取关键信息。支持多文档交叉查询，能够智能生成摘要和关键要点。",
        ],
      },
      {
        heading: "使用场景",
        body: [
          "学术论文阅读与摘要生成",
          "合同与法律文档关键条款提取",
          "技术文档快速查询与引用定位",
          "多文档对比分析与差异发现",
        ],
      },
    ],
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
