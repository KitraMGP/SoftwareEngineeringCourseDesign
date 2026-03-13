# 前端设计协作文档（准备版）

更新时间：2026-03-13

## 1. 文档目的

本文档用于在前端正式编码前，先固定以下内容：

- 项目目标与前端范围
- 现有 UI 原型可提取的信息
- 当前后端能力、接口契约现状与联调风险
- 前端项目架构和开发规范的候选方案
- 仍需由用户拍板的关键工程决策

说明：

- 本文档是前端讨论用工作文档，不替代 [detailed-design.md](/home/kitra/Projects/SoftwareEngineeringCourseDesign/docs/detailed-design.md)。
- 后端主设计仍以 [detailed-design.md](/home/kitra/Projects/SoftwareEngineeringCourseDesign/docs/detailed-design.md) 为准。
- 本文档当前状态是“准备版”，在完成本轮协商后再整理成正式前端详细设计文档。

## 2. 已阅读资料

- [detailed-design.md](/home/kitra/Projects/SoftwareEngineeringCourseDesign/docs/detailed-design.md)
- [implementation-kickoff.md](/home/kitra/Projects/SoftwareEngineeringCourseDesign/docs/implementation-kickoff.md)
- [openapi-v1-draft.yaml](/home/kitra/Projects/SoftwareEngineeringCourseDesign/docs/openapi-v1-draft.yaml)
- [chat-record.md](/home/kitra/Projects/SoftwareEngineeringCourseDesign/docs/chat-record.md)
- [README.md](/home/kitra/Projects/SoftwareEngineeringCourseDesign/backend/README.md)
- [backend-notes.md](/home/kitra/Projects/SoftwareEngineeringCourseDesign/tmp/backend-notes.md)
- 原型图：
  - [login.png](/home/kitra/Projects/SoftwareEngineeringCourseDesign/frontend/UI原型/login.png)
  - [chat.png](/home/kitra/Projects/SoftwareEngineeringCourseDesign/frontend/UI原型/chat.png)

## 3. 项目目标与前端范围背景

从主设计文档可以确认，本项目 V1 的产品目标是“基于大模型的私有知识库问答系统”，核心角色有两类：

- 普通用户：注册、登录、管理自己的知识库与文档、创建会话并进行问答
- 管理员：管理用户、模型配置、系统参数、配额、任务与审计

已确认的系统边界：

- 后端形态是“模块化单体 API + 独立 worker”
- 问答链路采用标准 RAG
- 会话最多绑定一个知识库
- 登录策略是单设备登录
- 聊天使用 SSE 流式返回
- 前端首期引用展示只要求显示文档名
- 文档上传支持 `PDF / TXT / Markdown / DOCX`
- 管理员不能查看用户文档正文和聊天正文

对前端的直接影响：

- 用户端和管理端的权限边界必须在路由和页面能力上明确区分
- 文档上传、重建索引、资源删除都存在异步状态，需要前端有状态轮询或任务可视化方案
- 聊天界面不能只做静态消息列表，必须预留 SSE、停止生成、重生成、引用来源、回退提示等能力

## 4. 现有 UI 原型观察

当前仓库只提供了两个原型页面：登录页和聊天页，没有提供知识库管理、文档管理、个人资料、管理员界面的原型图。

### 4.1 登录页原型

从 [login.png](/home/kitra/Projects/SoftwareEngineeringCourseDesign/frontend/UI原型/login.png) 可以确认：

- 视觉风格偏插画背景 + 半透明浮层，不是传统后台表单页
- 页面中心只有两个输入项：账号、密码
- 登录入口为单按钮
- 右侧有“注册”入口
- 右上角存在关闭按钮样式

目前仍不明确的地方：

- 注册是独立页、弹窗还是同页切换
- 是否需要“记住我”
- 是否展示登录错误提示、冻结提示、验证码或找回密码入口
- 关闭按钮对应的真实行为是什么

### 4.2 聊天页原型

从 [chat.png](/home/kitra/Projects/SoftwareEngineeringCourseDesign/frontend/UI原型/chat.png) 可以确认：

- 页面采用左侧窄边栏 + 右侧主聊天区域布局
- 左侧目前仅出现一个“聊天”导航项，没有体现会话列表和知识库列表
- 主区顶部是助手欢迎语
- 中部展示用户和助手对话气泡
- 底部是固定输入框，右下角有发送按钮
- 助手消息旁边出现了一个重新生成图标
- 输入框上方或附近出现了一个额外图标，可能表示清空、撤销或重新生成

目前仍不明确的地方：

- 左侧是否只是一级导航，还是后续要展开历史会话列表
- 知识库选择控件放在何处
- 引用来源、未命中知识库提示、流式输出状态放在何处
- 是否支持多标签页式工作区
- 会话重命名、删除、新建入口放在何处

### 4.3 原型覆盖缺口

当前原型没有覆盖但 V1 实际需要的前端界面：

- 注册页
- 知识库列表页
- 知识库详情页
- 文档上传与文档状态页
- 个人资料页
- 修改密码页
- 管理端相关页面
- 空状态、错误状态、加载状态
- 移动端或窄屏适配

结论：

- 现有原型只能帮助确定“视觉方向”和“两个核心入口页面的基础布局”。
- 还不足以直接支持完整前端开发，需要补充信息架构和页面流设计。

## 5. 当前后端真实能力与联调现状

前端设计不能只看理想设计文档，还要对齐当前后端已实现情况。

### 5.1 当前已可联调的后端能力

- 认证：
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `POST /api/v1/auth/refresh`
  - `POST /api/v1/auth/logout`
- 当前用户：
  - `GET /api/v1/users/me`
  - `PUT /api/v1/users/me`
  - `PUT /api/v1/users/me/password`
- 会话：
  - `GET /api/v1/sessions`
  - `POST /api/v1/sessions`
  - `GET /api/v1/sessions/{sessionId}`
  - `DELETE /api/v1/sessions/{sessionId}`
- 知识库：
  - `GET /api/v1/knowledge-bases`
  - `POST /api/v1/knowledge-bases`
  - `GET /api/v1/knowledge-bases/{kbId}`
  - `PUT /api/v1/knowledge-bases/{kbId}`
  - `DELETE /api/v1/knowledge-bases/{kbId}`
- 文档：
  - `GET /api/v1/knowledge-bases/{kbId}/documents`
  - `POST /api/v1/knowledge-bases/{kbId}/documents`
  - `GET /api/v1/knowledge-bases/{kbId}/documents/{docId}`
  - `DELETE /api/v1/knowledge-bases/{kbId}/documents/{docId}`
- 任务触发：
  - `POST /api/v1/knowledge-bases/{kbId}/reindex`

### 5.2 当前尚未可联调的能力

- 聊天主链路尚未实现：
  - `POST /api/v1/sessions/{sessionId}/messages` 当前为占位
  - `POST /api/v1/sessions/{sessionId}/messages/{messageId}/regenerate` 当前返回 `501`
  - `POST /api/v1/sessions/{sessionId}/stream/stop` 当前返回 `501`
- 管理端接口整体仍为占位，当前全部返回 `501`
- `pdf` 文档虽然可以上传，但 worker 目前会把 ingest 任务标记为 `failed`
- embedding 与检索仍是占位实现，聊天相关真实 RAG 结果还不能验证

### 5.3 对前端范围的直接影响

如果现在进入前端开发，推荐按阶段拆分：

1. 用户端基础骨架
   - 登录、注册、鉴权、路由守卫、全局布局
2. 知识库与文档管理
   - 知识库 CRUD
   - 文档上传、列表、状态展示
3. 会话管理与聊天布局
   - 新建会话、历史会话列表、查看已有消息
   - 聊天发送区先做占位或 mock
4. SSE 问答完成后补齐真实聊天链路
5. 管理端等后端可用后再推进

## 6. 接口契约现状与前端风险

OpenAPI 草案和当前后端代码存在差异，前端若直接按 OpenAPI 生成客户端，当前会有一定联调风险。

### 6.1 已确认的差异

- 列表响应字段不一致
  - OpenAPI 倾向返回 `list`
  - 当前后端实际返回 `items`
- 创建知识库返回字段不一致
  - OpenAPI：`data.id`
  - 当前后端：`data.knowledge_base_id`
- 更新知识库响应不一致
  - OpenAPI：简单成功响应
  - 当前后端：直接返回更新后的知识库对象
- 会话详情响应不一致
  - OpenAPI：建议包含知识库摘要
  - 当前后端：返回 `session + messages`，其中 `session` 带 `knowledge_base_name`
- 更新个人资料契约不一致
  - OpenAPI 中 `UpdateProfileRequest` 包含 `email`
  - 当前后端实际只接受 `nickname`、`avatar_url`
- 错误码表尚未完全对齐
  - 主设计文档中的错误码体系更细
  - 当前后端实现的 `httpx` 错误码更少，且部分数值不同

### 6.2 前端设计建议

在接口契约未统一前，前端不建议直接把 OpenAPI 生成结果作为唯一事实来源。

建议候选方案：

- 方案 A：先手写类型化 API Client，对当前后端真实响应做一层适配
- 方案 B：先修订 OpenAPI，使其与当前后端一致，再做代码生成
- 方案 C：生成客户端，但在 BFF/adapter 层做字段映射

当前更稳的建议是 A 或 B，不建议直接“裸生成后即用”。

## 7. 鉴权与部署相关约束

### 7.1 当前后端鉴权事实

- 登录成功后：
  - `access_token` 在 JSON 响应体中返回
  - `refresh_token` 默认通过 `HttpOnly Cookie` 返回
- 刷新接口：
  - 优先从 Cookie 读取 `refresh_token`
  - 若 Cookie 不存在，当前后端也支持从请求体 `refresh_token` 字段读取
- refresh cookie 默认路径为 `/api/v1/auth`
- access token 默认有效期 30 分钟
- refresh token 默认有效期 7 天

### 7.2 当前未解决的问题

- 后端目前没有看到 CORS 中间件
- 如果前后端分不同域名或端口：
  - 登录和刷新需要处理跨域 Cookie
  - 前端请求需要显式 `withCredentials`
  - 后端需要补 CORS 和 Cookie 策略

### 7.3 对前端架构的影响

前端需要尽早确定部署形态：

- 同域部署或由反向代理统一域名
- 本地开发通过 dev proxy 转发
- 还是完全分离部署并改造 CORS/Cookie 策略

这会直接影响：

- HTTP 客户端配置
- 刷新令牌方案
- 本地开发体验
- 登录状态恢复逻辑

## 8. 推荐的前端架构讨论基线

以下不是最终结论，而是结合当前项目状态给出的建议基线，便于协商。

### 8.1 应用形态

建议优先考虑两种路线：

- 路线 A：先做单一用户端 SPA，管理员端后置
- 路线 B：单仓库双应用，用户端和管理端分别构建

不太建议一开始就把用户端和管理端完全揉成一个单 SPA，因为：

- 两端角色边界清晰
- 当前原型只覆盖用户端
- 管理端后端接口尚未实现
- 后期权限和打包边界更容易失控

### 8.2 模块划分建议

若采用 Vue 3 工程，建议从一开始按领域拆分，而不是按页面文件堆叠。

建议目录思路：

```text
frontend/
  src/
    app/
    router/
    layouts/
    modules/
      auth/
      chat/
      knowledge-base/
      document/
      profile/
      admin/
    shared/
      api/
      components/
      composables/
      stores/
      styles/
      utils/
```

### 8.3 前端需要重点预留的基础能力

- API Client 与统一错误处理
- 鉴权存储与自动刷新
- SSE 流式事件解析
- 文件上传与大文件状态反馈
- 页面级 loading、empty、error 状态组件
- 路由权限守卫
- 任务状态轮询

### 8.4 UI 实现方向建议

基于现有原型，建议不要做成传统“表格后台模板风格”。

更适合的实现方向：

- 以自定义主题样式为主
- 基础表单与弹层可借助成熟组件库
- 聊天区、布局、卡片、导航和状态组件尽量自定义

原因：

- 登录页和聊天页都有明显视觉风格，不适合完全套通用后台模板
- 知识库问答产品核心体验在聊天页，默认管理后台 UI 风格会显得割裂

## 9. 当前建议的开发顺序

在不等待全部设计细节一次性定完的前提下，建议采用以下推进方式：

1. 先完成前端正式设计文档
2. 再初始化工程基座
3. 优先落地鉴权和全局布局
4. 然后落知识库与文档管理
5. 最后接入真实聊天 SSE
6. 管理端单独安排阶段

## 10. 第一轮已确认结论

以下结论基于 2026-03-13 本轮协商，后续前端正式设计默认以此为基线推进。

### 10.1 范围与应用拆分

- 前端设计范围：用户端 + 管理端一起设计
- 应用拆分方式：单仓库双应用
- 缺失原型页面的处理：本轮先不补页面草图说明，优先完成架构与规范设计

对设计的直接影响：

- 正式文档需要同时覆盖用户端与管理端的信息架构
- 但页面视觉设计的细化深度，会优先集中在已有原型支撑的用户端
- 管理端更适合先定结构、权限、布局和组件规范，再进入视觉细化

### 10.2 技术栈与工程基座

- 前端框架：`Vue 3 + TypeScript + Vite`
- 服务端数据管理：`Pinia + Vue Query`
- 样式体系：`TailwindCSS`
- 组件策略：引入一个基础组件库做通用控件，聊天与主要布局自定义
- 包管理器：`pnpm`
- 测试基线：`ESLint + Prettier + Vitest + Playwright`
- 国际化策略：当前只做中文，但在设计上可预留扩展空间
- 响应式策略：桌面优先，移动端只保证可用

说明：

- “样式体系”当前只确认到 `TailwindCSS`，尚未进一步确认是否搭配 `SCSS`、`CSS Modules` 或仅保留少量全局样式文件。
- “组件库”目前只确认会引入，但尚未确认具体是 `Element Plus`、`Naive UI`、`Arco Design Vue` 还是其他方案。

### 10.3 API、鉴权与开发阶段策略

- 接口契约基线：以修订后的 `OpenAPI` 为准
- 部署形态基线：同域部署或开发环境通过 `dev proxy` 转发，优先复用 `HttpOnly Cookie`
- `access token` 存放策略：`localStorage`
- 聊天接口未完成前的开发策略：聊天页真实布局先落地，发送消息区域展示“功能开发中”

对设计的直接影响：

- 正式文档需要明确“前端以 OpenAPI 契约为主，当前后端实现偏差需单独列兼容清单”
- API Client 层需要预留 token 注入、自动刷新和失败重试机制
- 本地开发规范需要写清 dev proxy 的推荐做法

### 10.4 用户端页面方向

- 聊天页左侧区域：一级导航 + 会话列表混合
- 知识库入口：集成在聊天页侧边抽屉
- 注册流程：独立注册页
- 个人资料与修改密码：单独“个人中心”
- 引用展示：助手消息下方显示文档名列表
- “未命中知识库”提示：作为助手消息头部标签
- 文档处理状态：在知识库详情页轮询文档列表状态
- `pdf` 上传：前端首期不做额外预处理，按后端结果提示

### 10.5 视觉方向补充

你已额外补充一个重要要求：

- 当前原型图样式较简陋，正式设计时需要做视觉美化

这意味着正式文档不能只描述功能结构，还需要补充：

- 视觉系统方向
- 色彩与字体策略
- 布局节奏与留白原则
- 用户端与管理端的统一品牌感，以及它们之间的风格差异边界

## 11. 第二轮已确认结论

以下结论基于 2026-03-13 第二轮协商。

### 11.1 设计系统与工程基座

- 基础组件库：`Element Plus`
- 样式体系：`TailwindCSS + SCSS`
- 双应用共享方式：共享一套 `shared/` 基础层，用户端和管理端各自独立业务模块
- API Client 生成策略：生成完整客户端代码，再做轻量封装
- 登录态恢复策略：应用启动时直接调用 `/auth/refresh` 获取新 token

设计含义：

- 共享层可沉淀 `api`、`types`、`auth`、`utils`、`shared components`
- 用户端与管理端的业务模块、路由和布局保持分离
- 样式层建议以 `TailwindCSS` 负责原子布局，`SCSS` 负责主题变量、复杂组件样式和动画

### 11.2 联调契约的一个待澄清冲突

你在第一轮选择了：

- “接口契约基线：以修订后的 `OpenAPI` 为准”

你在第二轮回复里的 `44B`，我按常规输入错误暂时理解为：

- 第 4 题选择 `B`
- 即“先以当前后端真实响应为准，`OpenAPI` 后补”

这两项彼此冲突，因此目前不能都作为最终结论写入正式文档。这个点保留到下一轮单独确认。

### 11.3 用户端结构与交互细节

- 聊天页左侧结构采用你给出的精确版本：
  - 顶部是会话列表
  - 中部是一级导航
  - 底部是用户头像和昵称，点击进入个人中心
- 知识库抽屉职责：只负责选择当前会话绑定的知识库
- 聊天接口未完成阶段：使用本地假数据演示聊天界面，但禁用真实发送
- 个人中心范围：个人资料 + 修改密码

这意味着：

- 知识库的 CRUD 与文档管理必须仍然落在独立页面或独立模块，而不是全部塞进聊天抽屉
- 聊天页需要区分“真实消息历史”和“未接通后端时的演示态”

### 11.4 管理端范围

- 管理端文档深度：信息架构 + 路由结构 + 页面职责
- 管理端视觉策略：与用户端同品牌，但更克制、更偏管理界面

说明：

- 这意味着管理端本轮文档不需要深入到每个表格列和筛选项的详细交互
- 但需要把菜单结构、页面边界、权限分层和布局方案定清楚

### 11.5 视觉方向与素材

- 用户端视觉方向：延续原型的轻柔、空气感、插画气质，但做高级化处理
- 登录页背景策略：延续插画/场景背景
- 登录页背景素材已存在：
  - [登录界面背景图.jpg](/home/kitra/Projects/SoftwareEngineeringCourseDesign/frontend/UI原型/登录界面背景图.jpg)
- 登录页背景需覆盖一层黑色半透明玻璃层，用于降低亮度、提高表单可读性
- 聊天页气质：沉浸式、偏产品感的 AI 对话空间
- 字体策略：使用更有识别度的 Web 字体组合
- 没有原型的页面：正式文档里要补低保真结构说明
- 没有原型的页面允许继续与你协商设计
- 更多图片素材后续由你提供

## 12. 第三轮已确认结论

以下结论基于 2026-03-13 第三轮协商。

### 12.1 契约与命名

- 正式设计文档以当前后端真实实现为准，`OpenAPI` 后续补齐
- 产品名称和 AI 助手名称当前都未固定，正式文档先使用占位名

说明：

- 这意味着正式前端设计文档会明确标注“当前联调契约来源于后端真实实现”
- `docs/openapi-v1-draft.yaml` 在当前阶段只作为参考资料，不作为唯一真相源

### 12.2 用户端无原型页面的信息架构

- 用户端一级导航主项：`知识库 / 关于`
- 知识库模块形态：知识库列表页 + 知识库详情页
- 知识库详情页结构：顶部知识库信息，下面文档列表和上传区
- 新建会话：V1 不暴露模型选择，前端固定默认模型
- 注册页字段范围：用户名、邮箱、密码、确认密码、勾选协议
- 个人中心：首期不开放头像编辑

这里隐含一个重要结构结论：

- “聊天”不作为中部一级导航项，而是由左侧顶部的会话列表承担进入与切换职责
- 一级导航更像“除聊天会话区之外的功能入口”

### 12.3 管理端信息架构

- 管理端一级菜单偏向：
  - `用户`
  - `模型配置`
  - `任务`
  - `审计`
  - `系统设置`
- 管理端需要单独首页 Dashboard，用于展示系统摘要卡片

说明：

- 你当前给出的分组没有单独列出“配额”，这会和主设计文档里的配额能力产生一个菜单归属问题，保留到下一轮单独确认

### 12.4 聊天页与知识库关系

- 聊天页需要显示当前绑定知识库的状态条
- 如果会话没有绑定知识库，前端不额外做提示

## 13. 第四轮已确认结论

以下结论基于 2026-03-13 第四轮协商。

### 13.1 关于页与导航补充

- “关于”页同时承载：
  - 产品介绍
  - 使用帮助
  - 项目/版本信息
- “关于”页需要登录后才能访问，作为应用内页面存在

### 13.2 管理端菜单补充

- 配额管理放在 `系统设置` 下作为二级菜单
- Dashboard 以系统摘要卡片为主，核心内容包括：
  - 用户数
  - 知识库数
  - 文档数
  - 任务状态

### 13.3 视觉补充

- 登录页表单区采用中央悬浮玻璃卡片
- 聊天页消息区采用：
  - 用户消息不显示头像
  - 助手消息显示头像

## 14. 当前状态

前端协商阶段已经收敛完成，下一步进入正式前端详细设计文档整理。

正式文档将以本文件中的已确认结论为输入，并补充：

- 前端总体架构
- 路由与模块划分
- 页面信息架构
- 组件分层策略
- 状态管理方案
- API 与 SSE 接入规范
- 鉴权与错误处理规范
- 目录结构规范
- 代码规范与测试规范
- 分阶段开发计划
