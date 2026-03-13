# Frontend Work Notes

更新时间：2026-03-13

## 当前目的

- 完成 V1 前端工程初始化、双应用基座和用户端核心页面开发
- 用本文件记录前端开发进度、契约同步情况、已知限制和后续联调动作

## 本轮已完成

- 已阅读和复核前端相关资料：
  - `docs/detailed-design.md`
  - `docs/frontend-detailed-design.md`
  - `docs/implementation-kickoff.md`
  - `backend/README.md`
  - `tmp/backend-notes.md`
- 已查看原型图：
  - `frontend/UI原型/login.png`
  - `frontend/UI原型/chat.png`
  - `frontend/UI原型/登录界面背景图.jpg`
- 已重新读取更新后的 `docs/openapi-v1-draft.yaml`
  - 当前文档已改为“以后端真实实现为准”的契约快照
  - 已据此调整前端共享类型和部分状态枚举理解
- 已初始化前端 monorepo：
  - `frontend/package.json`
  - `frontend/pnpm-workspace.yaml`
  - `frontend/tsconfig.base.json`
  - `frontend/tailwind.config.ts`
  - `frontend/postcss.config.cjs`
  - `frontend/.eslintrc.cjs`
  - `frontend/prettier.config.cjs`
  - `frontend/vitest.config.ts`
  - `frontend/playwright.config.ts`
- 已建立共享层：
  - `frontend/packages/shared/`
  - 包含 API client、认证 store、共享类型、工具函数、基础组件、样式 token
- 已建立用户端应用：
  - `frontend/apps/user/`
  - 已实现登录页、注册页、聊天工作区壳层、知识库列表页、知识库详情页、关于页、个人资料页、安全设置页
- 已建立管理端应用：
  - `frontend/apps/admin/`
  - 已实现后台登录页、管理员鉴权守卫、Dashboard、用户、模型配置、任务、审计、系统设置、配额页骨架

## 已确认事实

- 项目 V1 目标仍是“私有知识库问答系统”
- 前端采用“单仓库双应用 + 共享基础层”方案
- 已确认工程基线：
  - `Vue 3 + TypeScript + Vite`
  - `Pinia + Vue Query`
  - `TailwindCSS + SCSS`
  - `Element Plus`
  - `pnpm`
  - `ESLint + Prettier + Vitest + Playwright`
- 用户端当前已接上真实后端能力：
  - 登录
  - 注册
  - refresh 启动恢复
  - 当前用户信息
  - 修改资料
  - 修改密码
  - 会话创建、列表、详情、删除
  - 知识库创建、列表、详情、更新、删除
  - 文档上传、列表、详情、删除
  - 知识库重建索引任务创建
- 管理端后端接口当前仍未实现
  - 但独立后台应用、管理员角色守卫和页面骨架已就位
- 登录页已沿用背景图 `frontend/UI原型/登录界面背景图.jpg`
- 聊天页继续采用“消息预览 + 禁用发送”的阶段策略
- 知识库详情页已按真实后端能力接入文档状态轮询
- OpenAPI 已更新后，文档状态应按以下枚举理解：
  - `pending`
  - `processing`
  - `available`
  - `failed`
  - `deleting`

## 当前后端限制

- SSE 聊天接口尚未实现
- 重生成和停止生成接口尚未实现
- 管理端接口整体尚未实现
- `pdf` 当前会上传成功，但 ingest 任务会失败
- embedding、向量检索和真实 RAG 仍是占位能力
- 当前没有“更新会话知识库绑定”的真实接口
  - 前端已调整为“切换知识库时从新会话开始”
- 后端目前没有看到 CORS 中间件

## 契约与联调说明

- 当前前端请求层与共享类型应优先跟随更新后的 `docs/openapi-v1-draft.yaml`
- 该 OpenAPI 文档已可作为当前阶段前端契约快照使用
- 若后端继续变更真实实现，需要同步更新共享类型和请求层
- 若前后端分域部署，需要重新确认 refresh cookie、CORS、`withCredentials` 和 dev proxy

## 当前阶段结论

- 前端已从“仅有设计文档”进入“正式工程代码已落地”阶段
- 用户端 V1 主路径已基本具备：
  - 认证
  - 聊天工作区壳层
  - 知识库管理
  - 个人中心
  - 关于页
- 管理端已具备可继续承接真实接口的独立应用壳层

## 当前未完成项

- 尚未执行 `lint`
- 尚未启动 user/admin dev server
- 尚未使用浏览器 MCP 对页面进行实际渲染检查和样式微调

## 本轮测试结论

- 已完成 `pnpm typecheck`
  - 用户端通过
  - 管理端通过
- 已完成 `pnpm build`
  - 用户端通过
  - 管理端通过
- 本轮修复内容：
  - 修正了知识库上传自定义请求的 `Element Plus UploadRequestOptions` 类型错误
  - 补充了 `Element Plus` 全局样式引入，表单控件恢复正常显示
  - 修正了若干查询缓存失效 key 不精确的问题
  - 修正了上传 `FormData` 时不应手动覆盖 `Content-Type` 的问题
  - 调整了 Tailwind/PostCSS 配置，使 utility class 已能实际生成到产物 CSS 中
- 已使用浏览器 MCP 检查登录页实际渲染
  - 用户端登录页已做一轮可读性调整
  - 管理端登录页基础布局正常
  - 用户端移动端登录页布局正常
- 当前本地后端未在 `127.0.0.1:8080` 运行
  - 真实注册/登录/知识库链路尚未完成浏览器验收
- 当前仍有非阻塞告警：
  - Sass legacy JS API deprecation 警告，属于工具链层提示
  - Rollup chunk size warning，当前包体偏大但不阻塞继续开发
  - Tailwind 对 workspace 相对扫描模式仍有性能提示，但当前样式已正确生成

## 后续约定

- 下一步优先让用户手动在 `frontend/` 下执行 `pnpm install`
- 安装完成后执行：
  - `pnpm build`
  - `pnpm typecheck`
  - 必要时 `pnpm lint`
- 如需页面验收：
  - 用户端启动 `pnpm dev:user`
  - 管理端启动 `pnpm dev:admin`
- 待服务启动后，继续使用浏览器 MCP 做页面检查、交互调试和细节修正

## 2026-03-13 前端 V1 第二轮联调更新

### 契约认知修正

- 之前“聊天页继续采用消息预览 + 禁用发送”的结论已过期
- 结合更新后的 `docs/openapi-v1-draft.yaml` 与后端 README，可确认当前真实后端能力为：
  - 未绑定知识库的空会话已支持 `POST /api/v1/sessions/{sessionId}/messages` SSE 聊天
  - 绑定知识库的会话发送消息仍会返回 `501`
  - `regenerate` 与 `stream/stop` 仍未实现

### 本轮前端实现

- 已将共享默认聊天模型从占位值改为与后端一致的 `deepseek-chat`
- 已在 `packages/shared/src/api/chat.ts` 增加真实 SSE 请求与事件流解析
  - 支持 `meta / delta / done / error`
  - 支持带 bearer token 的 `fetch`
  - 支持在 `401` 时尝试 refresh 后重试一次
- 用户端会话详情页已从“纯占位输入区”升级为“空会话可真实发送”
  - 未绑定知识库的会话可直接发送消息
  - 绑定知识库的会话继续禁用发送，并展示限制说明
- 已补充聊天区自动滚动、发送态按钮文案、流式占位文案与状态提示
- 首页聊天能力说明文案已同步改为当前真实范围

### 本轮静态检查

- 已完成 `pnpm lint`
  - 通过
- 已完成 `pnpm typecheck`
  - 用户端通过
  - 管理端通过
- 已完成 `pnpm build`
  - 用户端通过
  - 管理端通过

### 浏览器 MCP 验收

- 用户端：
  - 已进入已有绑定知识库会话，确认输入区保持禁用，提示文案正确
  - 已创建新的空会话：`8fad748b-6832-42c1-910c-0e9d1a9c9b0d`
  - 已在该空会话中发送真实消息，页面收到 assistant 回复
  - 刷新同一会话后，用户消息与 assistant 消息仍然存在，确认消息已真实落库而非仅前端临时态
- 管理端：
  - 使用普通用户已仍会进入 `/forbidden`，管理员守卫未受本轮共享层改动影响

### 当前后端限制（更新后）

- 空会话 SSE 聊天已实现并已联调通过
- 绑定知识库的检索式问答仍未实现
- `regenerate` 和 `stream/stop` 仍未实现
- 管理端真实业务接口整体仍未完成
- `pdf` 上传后仍可能在 ingest 阶段失败
- embedding / 向量检索 / 真正 RAG 仍是后续能力

### 后续可继续项

- 继续优化聊天页和知识库页移动端细节
- 若后端开放知识库问答，再将当前会话页扩展为“知识库命中引用”真实展示
- 视需要再做 chunk 拆分，降低当前 user/admin 首屏 JS 体积

## 2026-03-13 前端 V1 第三轮收尾

### 与后端进度记录对照

- 已再次核对 `tmp/backend-notes.md`
- 当前后端已可真实提供且应由前端承接的用户端能力包括：
  - 注册 / 登录 / refresh / logout
  - 当前用户信息 / 修改资料 / 修改密码
  - 会话 CRUD
  - 未绑定知识库的空会话 SSE 问答
  - 知识库 CRUD
  - 文档上传 / 列表 / 详情 / 删除 / 重建索引
- 对照后确认，之前前端仍缺少的可实现项主要是：
  - 用户端显式“退出登录”入口
  - 安全设置页的显式导航入口
  - `avatar_url` 的资料编辑与展示
- 上述缺口本轮已全部补齐

### 本轮新增实现

- 已新增共享组件 `UserAvatar`
  - 统一处理头像 URL 展示与首字母回退
- 用户端侧栏已补：
  - 安全设置入口
  - 退出登录按钮
  - 头像展示
- 个人资料页已补：
  - `avatar_url` 输入框
  - 前端 URI 校验
  - 头像预览与最近更新时间展示
- 管理端壳层账户卡片已改为复用头像组件
- 已同步修正若干过时文案：
  - 登录页信息卡
  - 关于页 FAQ / 版本说明
  - 会话空态预览文案
- 已修复一个实际页面问题：
  - 知识库列表页“搜索”按钮在桌面视口下出现竖排换行，现已改为正常单行展示

### 截图验收范围

- 用户端桌面：
  - 空会话详情
  - 绑定知识库会话详情
  - 知识库列表
  - 知识库详情
  - 关于页
  - 个人资料页
  - 安全设置页
  - 登录页
  - 注册页
- 管理端桌面：
  - 后台登录页
  - 无权限页
- 用户端移动端：
  - 空会话详情
  - 移动端侧栏抽屉
  - 登录页

### 截图验收结论

- 本轮截图检查后，未发现新的结构性布局错误
- 已发现并修复的唯一明确显示问题：
  - 知识库列表页搜索按钮文本换行
- 其余页面在当前真实数据下显示正常：
  - 登录/注册页正常
  - 用户端侧栏与新增安全设置/退出登录入口正常
  - 个人资料页头像回退显示正常
  - 知识库列表、详情、会话页、关于页正常
  - 管理端登录页与无权限页正常

### 本轮最终验证

- `pnpm lint` 通过
- `pnpm typecheck` 通过
- `pnpm build` 通过
- 非阻塞告警仍保留：
  - Sass legacy JS API deprecation
  - Tailwind content pattern 性能提示
  - Rollup chunk size warning
