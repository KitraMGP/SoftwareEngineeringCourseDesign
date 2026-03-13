# Project Agent Notes

## Scope

- 当前项目已进入前后端并行设计/开发阶段。
- 后端设计与实现以 `docs/detailed-design.md` 为最高优先级文档。
- 前端设计与实现以 `docs/frontend-detailed-design.md` 为最高优先级文档。
- 前后端协作时，若接口草案与当前后端真实实现冲突：
  - 前端当前以“后端真实实现”为联调真相源
  - `docs/openapi-v1-draft.yaml` 后续再补齐

## Key Docs

- `docs/detailed-design.md`: 详细设计与模块边界，继续开发前先读这里。
- `docs/frontend-detailed-design.md`: 前端正式详细设计文档，包含双应用架构、页面信息架构、状态管理、鉴权与视觉规范。
- `docs/frontend-design-prep.md`: 前端协商过程与已确认结论沉淀。
- `docs/implementation-kickoff.md`: 当前阶段拆解与落地顺序参考。
- `docs/openapi-v1-draft.yaml`: 接口草案，联调和补接口时参考。
- `backend/README.md`: 本地环境、迁移、启动、测试说明。
- `backend/Makefile`: 常用开发命令入口。
- `backend/compose.dev.yaml`: 本地 PostgreSQL + pgvector 开发容器。
- `backend/migrations/`: 数据库迁移脚本。
- `frontend/UI原型/`: 前端已有原型图与登录背景素材。

## Run And Test

- 由用户自己执行系统启动命令；除非用户明确要求，否则不要替用户长期启动 `run-api`、`run-worker`、数据库容器。
- 若后续开始前端工程初始化，同样默认由用户自己执行长期运行命令；除非用户明确要求，否则不要替用户长期启动前端 dev server。
- 最短启动路径：
  - `cp backend/.env.example backend/.env`
  - 修改 `backend/.env` 中的 `AUTH_JWT_SECRET`
  - `cd backend && make db-up && make migrate-up`
  - 由用户在两个终端分别执行 `make run-api` 和 `make run-worker`
- 测试命令：
  - `cd backend && make test`
  - 或 `cd backend && GOCACHE=/tmp/go-build go test ./...`
- 若服务已由用户启动，可继续做接口冒烟测试。

## Frontend Conventions

- 前端工程基线：
  - `Vue 3 + TypeScript + Vite`
  - 单仓库双应用：用户端 + 管理端
  - `pnpm`
  - `Pinia + Vue Query`
  - `Element Plus`
  - `TailwindCSS + SCSS`
  - `ESLint + Prettier + Vitest + Playwright`
- 推荐目录基线以 `docs/frontend-detailed-design.md` 为准，不要脱离该文档自行发散目录结构。
- 用户端与管理端共享基础层，但业务模块、路由和布局分离。
- 聊天页、登录页、主布局和品牌化视觉组件优先自定义，不要整页套后台模板。
- 当前只有登录页和聊天页原型；没有原型的页面应先按文档中的低保真结构实现，必要时继续与用户协商。
- 登录页沿用 `frontend/UI原型/登录界面背景图.jpg`，并覆盖黑色半透明玻璃层降低亮度。
- 当前产品名和 AI 助手名未固定；新增前端文案时避免把占位名写死为最终品牌名。
- 当前聊天 SSE、消息重生成、停止生成和管理端真实接口尚未完成；前端开发时要区分“目标结构”和“当前可联调范围”。
- 当前 `pdf` 上传后端会失败；若前端开始实现上传流程，需明确提示而不是假设 `pdf` 已可用。

## Progress Memory

- 工作进度主记录文件：`tmp/backend-notes.md`
- 前端工作进度主记录文件：`tmp/frontend-notes.md`
- 每次继续开发前先读取它，避免上下文压缩后遗漏约束、已完成项和测试结果。
- 完成重要开发、环境调整、测试结论或后续计划后，及时追加更新到对应文件。
