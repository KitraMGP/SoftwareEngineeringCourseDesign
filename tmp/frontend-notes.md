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

## 2026-03-19 品牌名定稿同步

- 已将前端产品名占位统一替换为 `流光问答`
- 已将前端助手名占位统一替换为 `流光`
- 已同步更新用户端 / 管理端 HTML 标题与前端设计文档中的命名说明

## 2026-03-19 关于页文案调整

- 已更新用户端关于页底部深色信息条
- 已移除 `Workspace` 和 `Account` 栏位
- 已新增“HNUST 23计科四班课程设计团队精心打造”说明与 `CREATORS` 成员名单

## 2026-03-16 知识库页提示文案清理

- 已删除知识库列表页和知识库详情页中不必要的“实现说明 / 当前后端限制 / 设计理由”类提示文案
- 本轮精简包括：
  - 列表页标题区的后端联调说明
  - 列表页搜索区的结构说明
  - 详情页统计卡片下的解释性提示
  - 详情页状态卡中的 PDF 限制提示
  - 详情页上传区的“自动刷新 / PDF 会失败”等说明
- 已执行：
  - `cd frontend && pnpm build`
- 已用浏览器 MCP 实际检查：
  - `http://127.0.0.1:5173/knowledge-bases`
  - `http://127.0.0.1:5173/knowledge-bases/bf4b147e-bbfc-4889-bcd3-03ea01d96e7b`
- 结果：
  - 知识库列表页不再显示后端实现说明
  - 知识库详情页不再显示文档体量、PDF 失败、自动刷新等提示
  - 文档上传 / 列表 / 详情 / 删除 / 重建索引
- 对照后确认，之前前端仍缺少的可实现项主要是：
  - 用户端显式“退出登录”入口
  - 安全设置页的显式导航入口
  - `avatar_url` 的资料编辑与展示
- 上述缺口本轮已全部补齐

## 2026-03-16 全站提示文案收敛

- 已继续扩展到全站范围，清理用户端与管理端中“实现状态 / 接口进度 / 占位说明 / 后续计划”类文案
- 本轮覆盖：
  - 用户端聊天页、知识库抽屉、关于页、登录页左侧说明、个人资料页、安全设置页
  - 管理端登录布局、后台主布局、Dashboard、用户、任务、审计、系统设置、配额、模型配置、无权限页
- 处理方式：
  - 删除“已接通 / 当前版本 / 501 / 占位 / 待接入 / 后续替换”等开发态提示
  - 保留真正必要的页面标题、操作按钮、基础字段和简短用途描述
  - 将部分卡片改为中性标签，避免把研发进度直接暴露到界面
- 已执行：
  - `cd frontend && pnpm build`
  - 关键词残留扫描未再发现上述类型文案
- 已用浏览器 MCP 实际检查用户端页面：
  - `http://127.0.0.1:5173/sessions/3b2676ef-61c6-4d1c-b9d7-c2f98ab8f617`
  - `http://127.0.0.1:5173/about`
  - `http://127.0.0.1:5173/me/profile`
  - `http://127.0.0.1:5173/me/security`
  - `http://127.0.0.1:5173/knowledge-bases`
  - `http://127.0.0.1:5173/knowledge-bases/bf4b147e-bbfc-4889-bcd3-03ea01d96e7b`
- 备注：
  - 管理端本地 dev server 当前未在浏览器可访问地址上运行，因此本轮对管理端做了代码扫描与构建验证，但未做浏览器直查

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

## 2026-03-16 工作台视觉收敛与输入交互修正

### 本轮目标

- 按页面验收结果收敛用户端工作台首页、聊天页和知识库管理页的视觉密度
- 修复用户端“整页刷新或直达子路由可能掉回登录页”的鉴权恢复问题
- 修复聊天输入框不能直接回车发送的问题，并明确 `Shift+Enter` 换行提示

### 本轮实现

- 已调整用户端工作区整体视觉：
  - 降低全局玻璃拟态强度、阴影和背景装饰对比度
  - 收窄左侧侧栏，压缩会话卡信息层级
  - 主工作区容器圆角、背景与阴影已统一收敛
- 已重做聊天首页信息结构：
  - 改为“主动作 + 次动作 + 简短状态卡”布局
  - 去掉首屏厚重的能力大段说明块
  - 保留会话创建与知识库选择两个核心入口
- 已优化聊天页：
  - 收缩顶部状态区和提示条
  - 降低消息卡和输入区的垂直占用
  - 用户消息改为更低饱和度样式，assistant 区块减少操作噪声
  - `重新生成` 按钮已改为悬停显现，避免默认打断阅读
- 已修复聊天输入交互：
  - 普通 `Enter` 发送
  - `Shift+Enter` 保留换行
  - 输入框占位符已明确标注发送/换行规则
  - 已避免输入法组合态误发送
- 已优化知识库列表页：
  - 搜索区压缩为更扁平的工具条
  - 卡片标题、元信息和操作区已收敛
  - 详情入口继续保留为主按钮，编辑/删除降为次级文本操作
- 已优化知识库详情页：
  - 取消对 embedding model 和时间字段的超大号统计卡展示
  - 改为“文档数高强调 + 其余参数中等强调”的信息结构
  - PDF 限制改为 badge 和简短说明，不再占据大段正文
  - 上传区和文档列表区已调整为更适合工作台的密度
- 已修复鉴权恢复问题：
  - `bootstrap()` 现优先使用 `localStorage` 中的 access token 拉取当前用户
  - 仅在 access token 无效时回退到 refresh 流程
  - 因此整页刷新和直达子路由时，不再无条件依赖 refresh 才能恢复登录态

### 本轮自动化测试

- 已新增 `packages/shared/src/auth/useAuthStore.test.ts`
  - 验证存量 access token 可直接恢复用户态
  - 验证 access token 失效后会回退 refresh 再恢复用户态
- 已新增 `apps/user/src/modules/chat/components/MessageComposer.test.ts`
  - 验证 `Enter` 触发提交
  - 验证 `Shift+Enter` 不触发提交
  - 验证输入法组合态与 stop 态不触发提交
- 已执行 `cd frontend && pnpm test`
  - 通过
- 已执行 `cd frontend && pnpm lint`
  - 通过
- 已执行 `cd frontend && pnpm typecheck`
  - 用户端通过
  - 管理端通过
- 已执行 `cd frontend && pnpm build`
  - 用户端通过
  - 管理端通过

### 本轮浏览器验收

- 已使用浏览器 MCP 验证用户端知识库详情页整页直达
  - 页面保持已登录态，未跳回登录页
- 已使用浏览器 MCP 验证用户端会话详情页整页直达
  - 页面保持已登录态，未跳回登录页
- 已验证聊天输入框占位符已显示：
  - `Enter` 发送
  - `Shift+Enter` 换行
- 已在真实会话中通过键盘事件触发 `Enter` 发送
  - 未点击发送按钮
  - 用户消息与 assistant 回复已成功渲染
- 已验证 `Shift+Enter` 不触发发送
  - 事件未被 `preventDefault`
  - 消息条数保持不变
- 已分别检查并截图确认：
  - 工作台首页
  - 聊天页
  - 知识库列表页
  - 知识库详情页

### 当前遗留提示

- 构建仍有既有非阻塞提示：
  - Sass legacy JS API deprecation
  - Tailwind content pattern 扫描过宽
  - user/admin chunk size warning
- 上述问题与本轮视觉和交互修复无直接冲突，后续可单独治理
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
  - Dashboard
  - 用户管理
  - 系统设置
- 用户端移动端：
  - 空会话详情
  - 移动端侧栏抽屉
  - 登录页

说明：
- 当前本地没有真实管理员账号，因此管理端壳层页是通过浏览器上下文内临时写入 Pinia 管理员状态后进行截图验收，目的是确认纯前端布局与占位页结构正常

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
  - 管理端 Dashboard / 用户管理 / 系统设置壳层页正常

### 本轮最终验证

- `pnpm lint` 通过
- `pnpm typecheck` 通过
- `pnpm build` 通过
- 非阻塞告警仍保留：
  - Sass legacy JS API deprecation
  - Tailwind content pattern 性能提示
  - Rollup chunk size warning

## 2026-03-13 会话页高度修复

### 问题现象

- 在桌面端创建并进入会话后，左侧会话工具栏高度正常
- 右侧会话主区域按内容自然撑高，导致页面整体超出视口
- 发送按钮会落到首屏下方，需要整页向下滚动才能看到

### 修复内容

- 已将用户工作区桌面端外层容器改为固定视口高度，避免 `main` 继续跟随内容增长
- 已将右侧 `main` 高度与左侧侧栏统一为 `100vh - 外层纵向 padding`
- 已为会话页根节点、消息列表和输入区补齐 `min-h-0 / shrink-0` 约束
  - 现在由消息列表承担滚动
  - 页头和输入区保持固定可见

### 本轮验证

- 已在 Firefox MCP 中以桌面端 `1366 x 768` 视口复测空会话详情页
  - 页面不再产生整页纵向滚动
  - 左右两栏高度已对齐
  - 发送按钮在首屏内可见
- 已在同一视口下复测绑定知识库会话
  - 右侧主区域同样未再溢出视口
- 已完成 `pnpm --filter @private-kb/user typecheck`
  - 通过

### 补充修正

- 后续复测发现右侧主区域下边缘仍比左侧高 `16px`
- 根因是 `main` 在桌面端仍继承移动端 `min-h-[calc(100vh-2rem)]`
  - 该值大于桌面端目标高度 `calc(100vh - 3rem)`
  - 导致 `main` 实际高度被最小高度顶大
- 已将桌面端 `main` 的 `min-height` 改为与 `height` 和左侧栏一致的 `calc(100vh - 3rem)`
- Firefox MCP 复测结果：
  - 左右两栏高度差已从 `16px` 降为 `0px`

## 2026-03-13 侧栏最近会话区域修复

### 问题现象

- 左侧“最近会话”区域在仅有两个会话卡片时依然过矮
- 第二个会话会被区域底部截断
- 即使滚动也无法把最后一个会话完整滚入可视区

### 根因

- “最近会话”外层容器中存在标题行
- 内部滚动列表错误使用了 `h-full`
- 结果是列表高度按“父容器整高”计算，叠加标题行后超出父容器，被外层 `overflow-hidden` 裁掉

### 修复内容

- 已将“最近会话”区域改为标准列布局
- 滚动列表从 `h-full` 改为 `min-h-0 flex-1`
- 同时压缩了侧栏底部工具区的垂直占用，释放更多列表可视高度

### 验证结果

- 已在 Firefox MCP 中以桌面端 `1366 x 768` 视口复测
- 列表滚动到底时，最后一个会话卡片现已可完整显示
- `pnpm --filter @private-kb/user typecheck` 通过

## 2026-03-13 会话输入框高度修复

### 问题现象

- 聊天页底部问题输入框默认高度过大
- 在桌面端会占用过多垂直空间，压缩消息显示区域

### 修复内容

- 已将输入框默认高度从 `rows=5` 下调为 `rows=3`
- 已为输入框增加：
  - 更低的最小高度
  - 最大高度限制
  - `overflow-y-auto` 内部滚动
- 同时略微收紧了输入区卡片与按钮区的上下内边距

### 验证结果

- 已在 Firefox MCP 中以桌面端 `1366 x 768` 视口复测
- 输入区可见高度已明显下降
- 注入 18 行测试文本后，输入框会在自身内部滚动，不再继续撑高底部区域
- `pnpm --filter @private-kb/user typecheck` 通过

## 2026-03-13 侧栏宽度抖动修复

### 问题现象

- 在“无会话首页”与“会话详情页”之间切换时，左侧栏宽度会发生轻微变化
- 导致整个工作区布局看起来有横向抖动

### 根因

- 侧栏组件桌面端同时存在 `lg:w-auto` 与条件宽度类
- 且作为 flex 子项仍允许被收缩
- 当右侧内容最小宽度变大时，左侧栏会被压窄

### 修复内容

- 已移除侧栏桌面端的 `lg:w-auto`
- 已为侧栏增加 `shrink-0`
- 侧栏在桌面端现在固定使用显式宽度，不再受右侧内容挤压

### 验证结果

- 已通过 Firefox MCP 用两个真实账号复测：
  - 无会话首页
  - 有会话详情页
- 在桌面端 `1366 x 768` 视口下，两种状态侧栏宽度现均为 `300px`
- `pnpm --filter @private-kb/user typecheck` 通过

## 2026-03-14 用户端聊天闭环同步到最新后端

### 对照后端新增能力后的前端补齐

- 已根据当前后端真实实现补齐用户端聊天页缺口：
  - 知识库绑定会话现在可直接发送问题，不再错误提示“待开放”
  - `GET /sessions/{sessionId}` 返回的 `citations` 已接入前端消息模型
  - assistant 消息已支持 `重新生成`
  - 会话生成中已支持 `停止生成`
- 已同步调整用户端首页、登录页和关于页中的阶段性说明文案
  - 不再继续展示“知识库问答未开放”的过期描述

### 本轮代码实现

- 共享层：
  - `packages/shared/src/types/domain.ts` 已补 `MessageCitation`、`StreamStopResult`
  - `packages/shared/src/api/chat.ts` 已补：
    - `regenerateSessionMessage`
    - `stopSessionStream`
    - 通用 SSE 请求封装
- 用户端聊天页：
  - `SessionPage.vue` 现已统一管理发送 / 重生成 / 停止生成三类流式状态
  - 普通会话与知识库会话统一使用真实后端 SSE
  - 发送中的按钮会切换为“停止生成”
  - 重生成会原位覆盖对应 assistant 消息卡片
  - 知识库命中与未命中会显示不同状态徽标
- 消息区：
  - `MessageThread.vue` 已新增 assistant 消息“重新生成”按钮
  - 已按真实后端字段展示引用来源文档名与页码
- 输入区：
  - `MessageComposer.vue` 已支持发送态与停止态切换

### 本轮静态检查

- 已完成 `pnpm --filter @private-kb/user typecheck`
  - 通过
- 已完成 `pnpm --filter @private-kb/user build`
  - 通过

### Firefox MCP 运行时验收

- 已使用真实页面创建新的知识库：
  - `RAG UI Smoke 1773419000`
- 已通过页面上传 `txt` 文档并确认状态到达 `available`
- 已通过页面创建绑定该知识库的新会话
- 已在知识库会话中发送真实问题，验证结果：
  - 页面状态条显示“知识库问答已接通”
  - assistant 消息显示“命中知识库”徽标
  - 回答下方出现引用来源区域，显示文档 `rag-ui-smoke`
- 已触发真实 `regenerate` 请求并在浏览器网络面板确认：
  - `POST /api/v1/sessions/{sessionId}/messages/{messageId}/regenerate`
  - 返回 `200`
- 已触发真实 `stream/stop` 请求并确认：
  - `POST /api/v1/sessions/{sessionId}/stream/stop`
  - 返回 `200`
  - 被停止的那一轮消息最终只保留 user 消息，未落库 assistant 回复
- 本轮浏览器检查未发现新的控制台错误

## 2026-03-16 前端构建告警收敛

### 本轮目标

- 解决上一轮遗留的非阻塞工具链问题：
  - Sass legacy JS API deprecation
  - Tailwind content 扫描范围警告
  - Rollup chunk size warning
  - 手工分包引入的 circular chunk warning
  - Vitest 的 Vite CJS Node API deprecation 提示

### 本轮实现

- 已新增共享工具：
  - `packages/shared/src/utils/installElementPlus.ts`
- Element Plus 已从两个应用入口的全量 `app.use(ElementPlus)` 改为按需组件注册
  - 同时继续通过 `provideGlobalConfig` 注入中文 locale
  - 这样避免把整套组件插件一起打进首包
- Vite 共享配置已调整：
  - 保留 `scss.api = "modern-compiler"`，Sass legacy JS API 告警已消失
  - 移除上一轮用于强拆 vendor 的 `manualChunks` 逻辑，避免继续产生循环分包告警
- Tailwind 配置已改为基于当前构建应用的精确扫描：
  - 当前 app 的 `src/**/*.{vue,html}`
  - `packages/shared/src/**/*.{vue,html}`
  - 不再扫描整个 `apps/**/*` 或过宽的父级路径
- 前端 workspace 根包已声明 `type: module`
  - Vitest 不再触发 Vite CJS Node API deprecation 提示
- `packages/shared/package.json` 已补齐 `element-plus` 依赖声明
- 已执行一次 `pnpm install --no-frozen-lockfile`
  - 用于同步 workspace 依赖链接和 lockfile

### 本轮验证

- 已完成 `cd frontend && pnpm lint`
  - 通过
- 已完成 `cd frontend && pnpm typecheck`
  - user / admin 均通过
- 已完成 `cd frontend && pnpm test`
  - 4 个测试文件、8 个测试全部通过
- 已完成 `cd frontend && pnpm build`
  - user / admin 均通过
- 当前构建输出已不再出现：
  - Sass legacy JS API 警告
  - Tailwind content pattern 警告
  - Rollup chunk size warning
  - circular chunk warning
  - Vitest 的 Vite CJS Node API 警告

### 当前产物观察

- user 端主入口 JS 约 `417.19 kB`
- admin 端主入口 JS 约 `366.17 kB`
- 两侧主入口都已低于默认的 Rollup chunk warning 阈值

### 额外说明

- 本轮尝试对 `http://127.0.0.1:5173` 与 `http://127.0.0.1:5174` 做运行态烟雾检查时，端口均未启动
- 因遵循“默认不替用户长期启动 dev server”的约定，本轮未额外拉起前端服务

## 2026-03-16 聊天输入框高度与 Markdown 渲染修复

### 问题现象

- 用户端聊天页底部问题输入框默认高度仍然偏高
- 在消息较长时，会进一步压缩可视消息区域
- 会话消息当前仍按纯文本展示
- assistant 回复中的标题、列表、代码块、链接等 Markdown 结构无法阅读

### 本轮实现

- 已收紧聊天输入区视觉占高：
  - `MessageComposer.vue` 默认从 `rows=2` 调整为 `rows=1`
  - 缩小输入区外层内边距、按钮行间距和按钮垂直 padding
  - 将默认 placeholder 改为更短文案，减少初始换行
  - 新增 textarea 自适应高度逻辑
    - 初始高度更低
    - 随输入内容增长自动扩展
    - 最高限制为 `128px`
- 已新增本地安全 Markdown 渲染工具：
  - `apps/user/src/modules/chat/utils/renderMarkdown.ts`
  - 支持：
    - 标题
    - 段落与换行
    - 无序/有序列表
    - 引用块
    - 行内代码与 fenced code block
    - 粗体 / 斜体 / 删除线
    - 安全链接
  - 原始 HTML 会被转义
  - `javascript:` 等危险链接不会注入
- `MessageThread.vue` 已改为渲染安全 Markdown HTML
  - 并补充了标题、列表、引用、代码块、链接等消息样式

### 本轮测试

- 已新增：
  - `apps/user/src/modules/chat/utils/renderMarkdown.test.ts`
  - `apps/user/src/modules/chat/components/MessageThread.test.ts`
- 已更新：
  - `apps/user/src/modules/chat/components/MessageComposer.test.ts`
- 已完成 `cd frontend && pnpm lint`
  - 通过
- 已完成 `cd frontend && pnpm typecheck`
  - user / admin 均通过
- 已完成 `cd frontend && pnpm test`
  - 6 个测试文件、12 个测试全部通过
- 已完成 `cd frontend && pnpm build`
  - user / admin 均通过

### 运行态说明

- 当前本地未检测到可直接复用的前端 dev server 或后端接口服务
- 因此本轮主要以组件级单测、类型检查和生产构建作为验收依据

## 2026-03-16 Markdown 列表缩进兼容修复

### 问题现象

- 上一轮接入 Markdown 渲染后，普通标题、段落和代码块已可显示
- 但真实回复中若列表项带有前导缩进，`* **技术实现**` 这类内容仍会退回为普通段落
- 页面上会直接看到原始星号，而不是渲染后的列表项

### 根因

- 自定义 Markdown 解析器在 block 级识别列表、引用和标题时，对前导空白不够宽容
- 导致模型常见的“缩进后再输出列表”的文本形态未被识别为列表块

### 修复内容

- 已调整 `apps/user/src/modules/chat/utils/renderMarkdown.ts`
  - block 级匹配前统一使用 `trimStart()`
  - 无序/有序列表、引用块、标题、block boundary 判断均已兼容前导缩进
- 已补充针对真实回复形态的单测
  - 覆盖“标题 + 缩进列表 + 加粗文本”的混合内容

### 本轮验证

- 已完成 `cd frontend && pnpm test -- renderMarkdown MessageThread`
  - 通过
- 已完成 `cd frontend && pnpm test`
  - 6 个测试文件、13 个测试全部通过
- 已完成 `cd frontend && pnpm lint`
  - 通过
- 已完成 `cd frontend && pnpm typecheck`
  - user / admin 均通过
- 已完成 `cd frontend && pnpm build`
  - user / admin 均通过

## 2026-03-16 PDF 后端能力同步

- 已根据后端最新实现同步认知：
  - 文本型 `pdf` 当前已可上传并完成 ingest
  - 扫描件或纯图片型 `pdf` 仍会在后端处理阶段失败，因为 OCR 尚未实现
- 前端后续提示语应以此为准：
  - 不再将所有 `pdf` 一概视为失败
  - 仅在后端返回无可提取文本时，提示“当前 PDF 为扫描件，OCR 尚未实现”

## 2026-03-16 管理后台真实联调与管理员登录修复

### 本轮实现

- 已补齐共享层管理后台能力导出：
  - `packages/shared/src/api/admin.ts`
  - `packages/shared/src/constants/queryKeys.ts`
  - `packages/shared/src/index.ts`
- 已修复管理端管理员切换账号链路：
  - `admin` 路由守卫不再把“已登录但非管理员”的用户从 `/login` 强行打回 `/forbidden`
  - 后台登录页在普通用户凭据登录成功后会立即执行 `logout`，清理后台登录态并停留在登录页
  - 无权限页已增加“切换账号 / 退出当前账号”操作
- 已将以下页面从静态占位替换为真实接口驱动页面：
  - `Dashboard`
  - `用户管理`
  - `任务管理`
  - `模型配置`
  - `系统设置`
  - `配额策略`
  - `审计日志`
- 已修复运行时组件注册缺口：
  - `packages/shared/src/utils/installElementPlus.ts`
  - 新增注册 `ElSelect / ElOption / ElTable / ElTableColumn / ElPagination`

### 本轮测试

- 已完成：
  - `cd frontend && pnpm --filter @private-kb/admin typecheck`
  - `cd frontend && pnpm --filter @private-kb/admin build`
  - `cd frontend && pnpm --filter @private-kb/user typecheck`
  - `cd frontend && pnpm --filter @private-kb/user build`
- 已使用浏览器 MCP 做管理端真实回归：
  - 在保留非管理员刷新令牌的旧状态下，访问 `http://127.0.0.1:5174/login` 不再被错误重定向到 `/forbidden`
  - 使用普通用户 `adminswitch0316 / Adminswitch123` 登录后台时，会触发 `login -> logout`，页面停留在 `/login`，密码框清空
  - 使用管理员 `admin / 12345678abc` 登录成功并进入 Dashboard
  - Dashboard 已正确显示真实统计数据
  - 用户页已正确加载 16 条账号，且“当前账号”冻结按钮为禁用态
  - 已在用户页成功执行一次冻结与一次恢复操作，目标用户：
    - `adminswitch0316`
  - 审计页已正确展示上述 `admin.user.freeze / admin.user.unfreeze` 记录
  - 任务页已正确加载 21 条任务；当前环境无 `failed` 任务，因此仅验证了列表展示、分页和“重试”按钮禁用态
  - 模型配置页、系统设置页、配额策略页、无权限页均已验证渲染正常
- 浏览器控制台检查结果：
  - 修复前存在 `Failed to resolve component: el-select / el-table / el-pagination`
  - 修复后页面切换不再出现上述运行时告警
## 2026-03-16 用户端知识库抽屉跨账号串数据修复

- 用户端聊天/知识库查询改为附带当前 `user scope`，避免跨账号复用 Vue Query 缓存。
- 用户身份切换时自动清空已选知识库和抽屉状态，避免上一账号的残留上下文。
- 浏览器 MCP 已完成同页切换回归：
  - 普通用户 `cachefixuser0316` 登录后可看到其测试知识库 `Cache Scope KB`
  - 退出后在同一标签页登录管理员 `admin / 12345678abc`，知识库抽屉恢复为空列表

## 2026-03-19 管理端 API Key 编辑补齐

### 本轮实现

- 已更新共享层管理端 provider 类型与请求：
  - `packages/shared/src/types/domain.ts`
  - `packages/shared/src/api/admin.ts`
  - `ProviderConfig` 新增 `api_key_source`
  - `created_at / updated_at / id` 改为可空，以兼容“仅由当前运行配置合成、尚未落库”的 provider 行
- 已将管理端“模型配置”页改为支持 API key 覆盖写入：
  - `apps/admin/src/modules/providers/pages/ProvidersPage.vue`
  - 页面已收敛为单一默认 provider：`DeepSeek`
  - 配置来源展示改为：`.env / 数据库 / 未配置`
  - 编辑通过弹窗完成，输入框始终为空，不回显旧 key
  - 页面显式提示：系统会优先读取 `.env`，只有 `.env` 未提供时才回退读取数据库
  - 因此当 `.env` 已提供 `DEEPSEEK_API_KEY` 时，页面会直接视为已配置
- 顺手清理了前端现存 lint 问题，避免本轮验收被历史问题阻塞：
  - `apps/admin/src/modules/users/pages/UsersPage.vue`
  - `apps/user/src/modules/chat/pages/SessionPage.vue`
  - `packages/shared/src/types/domain.ts`

### 本轮测试

- 已完成：
  - `cd frontend && pnpm lint`
  - `cd frontend && pnpm build`
- 当前未做浏览器 MCP 的最终页面点击验收，原因是本机 `:8080` 上仍是旧 API 进程：
  - 旧进程的 `GET /api/v1/admin/provider-configs` 仍返回空列表
  - 需要用户重启 API（以及如需验证数据库后备读取则一并重启 worker）后，再在管理端页面验证新的 DeepSeek 单卡片与编辑弹窗
