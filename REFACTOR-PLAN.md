# skill-hub 架构重构方案

> 分析日期：2026-08-12
> 基线：React 19 + Vite 8 + TypeScript + react-router-dom v7 + GSAP

---

## 1. 当前架构分析

### 1.1 当前目录树

```
skill-hub/
├── index.html
├── package.json
├── vite.config.ts
├── tsconfig*.json
├── public/
│   ├── favicon.svg          ✅ 被 index.html 引用
│   └── icons.svg            ❌ 未使用
└── src/
    ├── main.tsx             ✅ 入口
    ├── App.tsx              ⚠️ Router + 全部路由 + Layout 包裹，职责混在一起
    ├── App.css              ⚠️ 仅 .app/.main 两条规则
    ├── index.css            ✅ 全局 reset
    ├── animations/          ⚠️ GSAP 动画库，大量未使用导出
    │   ├── index.ts         （barrel 导出全部）
    │   ├── utils.ts         ⚠️ 未使用的 gsap 导入（编译错误）
    │   ├── page.ts          ✅ pageEnter/pageLeave/sectionEnter/breadcrumbEnter
    │   ├── card.ts          ✅ cardHover*/cardListEnter
    │   ├── button.ts        ✅ buttonHover*/buttonClick
    │   ├── dialog.ts        ❌ 全部未使用
    │   ├── menu.ts          ❌ 全部未使用
    │   ├── panel.ts         ⚠️ 仅 panelEnterRight 使用
    │   ├── loading.ts       ❌ 全部未使用
    │   └── progress.ts      ❌ 全部未使用
    ├── assets/              ❌ hero.png / react.svg / vite.svg 全部未引用
    ├── components/          ⚠️ 平铺，Layout 混入，无分类
    │   ├── Layout.tsx       （实际是 MainLayout）
    │   ├── Navbar.tsx / Navbar.css
    │   ├── Footer.tsx / Footer.css
    │   ├── Hero.tsx / Hero.css
    │   ├── SkillCard.tsx / SkillCard.css
    │   ├── SkillSection.tsx / SkillSection.css
    │   ├── SponsorBanner.tsx / SponsorBanner.css
    │   ├── OfficialAuthors.tsx / OfficialAuthors.css
    │   └── ScrollToTop.tsx
    ├── data/
    │   └── skills.ts        ⚠️ Types + 数据 + 查询函数混在一个 401 行文件
    └── pages/               ⚠️ 平铺，6 个页面内联样式重复
        ├── HomePage.tsx
        ├── SkillDetailPage.tsx + .css
        ├── AuthorPage.tsx
        ├── CategoryPage.tsx
        ├── CategoriesPage.tsx
        ├── OfficialPage.tsx
        ├── FeaturedPage.tsx
        └── SubmitPage.tsx
```

### 1.2 页面清单

| 路由 | 页面 | 职责 | 数据来源 | 重复代码 |
|---|---|---|---|---|
| `/` | HomePage | 组合 Hero + SponsorBanner + OfficialAuthors + 多个 SkillSection | data/skills | 页面入场动画模式 |
| `/skill/:skillId` | SkillDetailPage | 技能详情 + 侧栏相关技能 | getAllSkills() 查找 | 页面入场 + 面包屑 |
| `/author/:authorSlug` | AuthorPage | 作者主页 | getSkillsByAuthor() | 面包屑 + 容器样式 + 入场动画 |
| `/category/:slug` | CategoryPage | 分类列表 | getSkillsByCategory() | 同上 |
| `/categories` | CategoriesPage | 全部分类 | skillCategories | 同上 |
| `/official` | OfficialPage | 官方技能 | getOfficialSkills() | 同上 |
| `/featured` | FeaturedPage | 精选技能 | featuredSkills | 同上 |
| `/submit` | SubmitPage | 提交表单（本地 state） | 无 | 面包屑 + 容器样式 |
| — | **无 404** | — | — | — |

### 1.3 依赖关系（当前）

```
App.tsx ── BrowserRouter ── Layout ── Navbar / Footer / ScrollToTop
              │
              └── Routes ── 8 个 Page
                             │
               ┌─────────────┼─────────────┐
               ↓             ↓             ↓
           animations/    data/skills   components/SkillCard
           (GSAP)         (types+data+query)
```

### 1.4 关键重复模式（实测）

- **面包屑导航**：同一段 `nav + Link + chevron SVG` 代码在 6 个页面逐字复制（Categories / Category / Author / Official / Featured / Submit）。
- **页面容器样式**：`{ maxWidth: 1280|640, margin: "0 auto", padding: "40px 24px 60px" }` 在 5 个页面重复。
- **入场动画样板**：`pageRef + ctx + useEffect(killAll → pageEnter → sectionEnter → cleanup)` 在 7 个页面几乎逐字重复。
- **颜色硬编码**：`#7c3aed / #111827 / #6b7280 / #e5e7eb / #9ca3af` 散落在 8 个 CSS 和 6 个页面的内联样式中。

---

## 2. 问题清单

### P0 — 构建/运行阻断（必须立即修复）

1. **`tsc -b --force` 报 9 个编译错误**（增量缓存掩盖，`npm run build` 实际失败）：
   - `sectionEnter()` 收到 `NodeListOf<Element>`，签名只接受 `Element[] | string` —— 7 处（Hero, Author, Categories, Category, Featured, Official, SkillDetail）
   - `animations/utils.ts:1` 未使用的 `gsap` 导入
   - `Hero.tsx:4` 未使用的 `featuredSkills` 导入
2. **无 404 路由**：访问未知路径白屏。
3. **无 ErrorBoundary**：任何运行时错误直接整页崩溃。

### P1 — 结构/维护性（核心重构对象）

4. `App.tsx` 把 Router、Layout、全部路由混在一起；无路由模块、无 lazy loading、无 SEO。
5. 面包屑 + 页面容器 + 入场动画三大重复模式散落 6–7 个页面。
6. 无设计令牌（Design Tokens）：颜色/间距/圆角全硬编码。
7. 样式双轨制：组件用 co-located CSS，页面用巨型内联 `style` 对象。
8. `data/skills.ts` 混合 Types + 数据 + 查询函数。
9. 无 SEO：无 per-page title/description/canonical/OG/Twitter/JSON-LD。
10. `Navbar` 的 `menuOpen` 是死状态——移动端菜单按钮无任何行为。
11. `Footer` 用 `<a href>` 而非 `<Link>`，站内跳转整页刷新。
12. `Hero` 内联统计业务逻辑（getAllSkills + filter）。

### P2 — 死代码清理

13. `animations/` 约 75% 导出从未被使用：`pageLeave, breadcrumbEnter, cardListEnter, dialog*, menu*, panelLeave*, panelEnterLeft/Float/Toggle, loading*, progress*`。
14. `data/skills.ts` 的 `getAuthorsWithSkills()` 未使用；`Skill.authorAvatar` 字段未使用。
15. `src/assets/` 三个文件（hero.png/react.svg/vite.svg）全部未被引用。
16. `public/icons.svg` 未被引用。
17. 数据重复：`ace-step` 技能在 creative 和 media 分类下重复出现。
18. `animations/index.ts` 大 barrel 导出未使用项。

### P3 — 打磨（按需）

19. 响应式：页面用 `auto-fill, minmax` 网格，基本可用，但无统一 breakpoint 体系。
20. `Layout` 的 50ms `transitioning` 定时器 hack 可简化。
21. `prefers-reduced-motion` 已在动画 utils 处理 ✅。
22. 无障碍细节：`SkillCard` 下载按钮已有 `title`，可补 `aria-label`。

> 说明：项目**没有后端 API、没有多语言、没有复杂业务域**，因此"services 层、features/、i18n、全局状态管理"等模板结构**不适用**，强行引入属于过度设计，本方案明确不做。

---

## 3. 推荐架构（按实际项目裁剪）

```
src/
├── app/                      # 应用基础设施
│   ├── App.tsx               # 仅组合 Router + Layout + ErrorBoundary
│   └── router/
│       └── routes.tsx        # 统一路由表 + lazy loading + 404
├── layouts/
│   └── MainLayout.tsx        # 由 components/Layout 迁移
├── pages/                    # 页面按功能分目录（每个页面自带 section 和 assets 放得下的规模）
│   ├── home/
│   │   └── HomePage.tsx
│   ├── skill/
│   │   ├── SkillDetailPage.tsx
│   │   └── SkillDetailPage.css
│   ├── author/AuthorPage.tsx
│   ├── category/CategoryPage.tsx
│   ├── categories/CategoriesPage.tsx
│   ├── official/OfficialPage.tsx
│   ├── featured/FeaturedPage.tsx
│   ├── submit/SubmitPage.tsx
│   └── not-found/NotFoundPage.tsx
├── components/
│   ├── layout/               # 跨页面布局组件
│   │   ├── Navbar.tsx / Navbar.css
│   │   ├── Footer.tsx / Footer.css
│   │   └── ScrollToTop.tsx
│   ├── sections/             # 首页/复用 Section
│   │   ├── Hero.tsx / Hero.css
│   │   ├── SkillSection.tsx / SkillSection.css
│   │   ├── SponsorBanner.tsx / SponsorBanner.css
│   │   └── OfficialAuthors.tsx / OfficialAuthors.css
│   ├── skill/                # 技能业务组件
│   │   └── SkillCard.tsx / SkillCard.css
│   └── shared/               # 消除重复的共享组件
│       ├── PageContainer.tsx
│       ├── Breadcrumb.tsx
│       └── PageHeading.tsx
├── hooks/
│   └── usePageAnimation.ts   # 抽取 7 页重复的入场动画样板
├── data/
│   ├── types.ts              # Skill / Author / SkillCategory 类型
│   ├── skills.ts             # 纯数据
│   └── queries.ts            # getAllSkills / getSkillsBy* 查询函数
├── styles/
│   ├── globals.css           # 由 index.css 迁移
│   └── tokens.css            # 颜色/间距/圆角/字号 CSS 变量
├── config/
│   └── site.ts               # 站点名/URL/导航/SEO 默认值
└── assets/                   # 删除未使用的 3 个文件（或后续补图再放）
```

> 保持 `index.html`、`public/`、`vite.config.ts`、`animations/`（清理后）位置不变。

---

## 4. 模块边界与依赖规则

```
app ──→ layouts ──→ pages ──→ components ──→ hooks
  │        │          │           │            │
  │        │          │           └──────┬─────┘
  │        │          ↓                  ↓
  │        └──→   data / config / styles / animations(纯工具，任何人可依赖)
  ↓
router（被 app 依赖，依赖 pages）
```

规则：

1. **依赖方向单向向下**：`app → layouts → pages → components → hooks → data`。
2. `components/` 禁止 import `pages/`（公共组件不依赖具体页面）。
3. `pages/` 可以 import `components/`、`hooks/`、`data/`、`animations/`。
4. `components/` 可以 import `hooks/`、`data/`、`animations/`。
5. `data/`、`animations/`、`config/`、`styles/` 为叶子模块，不依赖任何上层。
6. UI 组件（shared/）不读取业务数据、不发起任何请求。
7. 查询逻辑统一走 `data/queries.ts`，组件不直接操作原始数组。
8. 禁止循环依赖：`pages → components` 时，组件不得反向 import 页面。

---

## 5. 重构路线（渐进式，每阶段验证）

| 阶段 | 范围 | 目标 | 验证 |
|---|---|---|---|
| **S1** | P0 构建修复 | 修 9 个编译错误 + 删除未使用导入；新增 404 页 + ErrorBoundary + catch-all 路由 | `tsc -b --force`、`npm run build` |
| **S2** | 死代码清理 | 删除未用动画导出/文件、`getAuthorsWithSkills`、`authorAvatar`、`assets/`、`public/icons.svg` | build + 手工回归 |
| **S3** | 共享组件 | 新增 `PageContainer` + `Breadcrumb`，替换 6 个页面的重复结构 | 页面视觉不变（截图对比） |
| **S4** | 布局迁移 | `components/Layout` → `layouts/MainLayout`；components 按 layout/sections/skill 分类 | build + 路由回归 |
| **S5** | 页面分目录 | pages/ 平铺 → 按功能目录；新增 404 页 | 全部路由点击回归 |
| **S6** | 数据拆分 | `data/skills.ts` → `types.ts` + `skills.ts` + `queries.ts` | tsc + 功能一致 |
| **S7** | 动画 Hook | 抽取 `usePageAnimation`，替换 7 页样板；修复 Hero 业务逻辑 | 动画行为一致 |
| **S8** | 路由与性能 | 路由表集中 + lazy loading + 每页 title/description（SEO 轻量方案） | build 产物对比、SEO 检查 |
| **S9** | 设计令牌 | `styles/tokens.css`，替换高频硬编码颜色 | 视觉一致 |
| **S10** | 收尾 | Navbar 死状态、Footer `<Link>`、README 更新 | 全量回归 |

**每阶段保证**：功能一致 / 页面一致 / 路由一致 / 交互一致 / SEO 不下降 / 性能不下降。任何行为变更先说明原因。

---

## 6. 明确不做的（避免过度设计）

- ❌ 不建 `services/api/` 层：项目无后端，数据全部本地静态。
- ❌ 不建 `features/`：只有一个业务域（skills），拆分无收益。
- ❌ 不引入状态管理库：无跨页面共享可变状态。
- ❌ 不引入 i18n 框架：仅中文，导航栏语言按钮为死功能。
- ❌ 不引入 CSS-in-JS / Tailwind：保持现有 co-located CSS 风格。
- ❌ 不创建测试/Storybook：项目无测试基础设施，本方案不引入。

---

## 7. 重构完成记录（2026-08-12）

| 阶段 | 内容 | 状态 |
|---|---|---|
| S1 | 修复 9 个编译错误（`sectionEnter` 签名 + 未使用导入）；新增 404 页 + ErrorBoundary + catch-all 路由 | ✅ |
| S2 | 清理死代码：4 个动画文件 / 25 个未用导出 / `getAuthorsWithSkills` / `authorAvatar` / `assets/` / `public/icons.svg` | ✅ |
| S3 | 新建 `PageContainer` + `Breadcrumb`，消除 6 个页面重复的面包屑与容器样式 | ✅ |
| S4 | `components/Layout` → `layouts/MainLayout`；components 按 layout/sections/skill/shared 分类 | ✅ |
| S5 | pages 按功能分目录（home/skill/author/category/categories/official/featured/submit/not-found） | ✅ |
| S6 | `data/skills.ts` 拆分为 `types.ts` + `skills.ts` + `queries.ts` | ✅ |
| S7 | 抽取 `usePageAnimation`（9 页样板收敛，oxlint 10 警告 → 0）；Hero 使用查询函数 | ✅ |
| S8 | 集中路由表 `app/router/routes.tsx` + lazy 代码分割 + `usePageMeta` SEO | ✅ |
| S9 | 设计令牌 `styles/tokens.css`，全部 CSS 与内联样式颜色迁移（残留 hex = 0） | ✅ |
| S10 | Navbar 移动菜单可用化（消除死状态）；Footer 站内链接改 `<Link>`；README 重写 | ✅ |

**最终验证**：`tsc -b --force` ✅ · `oxlint` 0 警告 ✅ · `npm run build` ✅（主包 306 kB，页面独立分包，较重构前 340 kB 下降约 10%）· 全路由浏览器回归 ✅

**最终目录**：

```
src/
├── app/
│   ├── App.tsx
│   ├── ErrorBoundary.tsx
│   └── router/routes.tsx
├── layouts/MainLayout.tsx
├── pages/{home,skill,author,category,categories,official,featured,submit,not-found}/
├── components/{layout,sections,skill,shared}/
├── hooks/{usePageAnimation,usePageMeta}.ts
├── data/{types,skills,queries}.ts
├── styles/{tokens,globals}.css
├── config/site.ts
└── animations/
```
