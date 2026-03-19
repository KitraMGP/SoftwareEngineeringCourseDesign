# Backend Work Notes

更新时间：2026-03-12

## 已确认约束

- 主设计文档以 `docs/detailed-design.md` 为准，高于其他草稿。
- 本轮只实现后端代码，前端后续再做。
- 架构采用“模块化单体 API + 独立 worker”。
- 当前代码优先落地 V1 范围，不额外加入 OCR、多知识库联合检索、多设备登录等超范围能力。
- 需要持续把规划、进度和重要决策写入 `tmp`，避免上下文压缩后丢失。

## 当前工程选型

- Go: `1.26`
- HTTP Router: `chi`
- DB Access: `pgx/v5`
- Auth: JWT access token + opaque refresh token(hash stored in DB)
- Password Hash: Argon2id
- Logger: `slog`
- Migration: 直接复用 Goose 格式 SQL

## 本阶段目标

第一阶段先完成一个可运行后端骨架：

1. 初始化 `backend` 工程目录和基础配置
2. 落地平台层：配置、日志、HTTP 响应、数据库、鉴权
3. 实现账户认证与当前用户模块
4. 补齐会话/知识库/任务/管理端路由骨架
5. 复制迁移文件并完成基础构建验证

## 关键实现策略

- 认证模块优先实现真实数据库读写，不先做 mock。
- 第一阶段先优先完成账户、会话、知识库和任务骨架；聊天主链路和文档处理后续再逐步补齐。
- `sqlc` 暂不在第一阶段引入，先用 `pgx` 手写 repository，把模块边界和接口先稳定下来；后续如需要再平滑切到 `sqlc`。

## 进行中

- 正在根据详细设计文档搭建第一阶段后端骨架。

## 当前进度

- 已创建根目录 `.gitignore`
- 已初始化 `backend` Go 模块，并补充 `.env.example`
- 已复制 `docs/migrations/*.sql` 到 `backend/migrations/`
- 已完成平台层基础设施：
  - 配置加载
  - `pgx` 数据库连接
  - 统一 JSON 响应与错误结构
  - request id / recover / access log 中间件
  - JWT access token + opaque refresh token
  - Argon2id 密码哈希
- 已完成账户模块首个纵切：
  - 注册
  - 登录
  - 刷新令牌
  - 退出登录
  - 当前用户信息
  - 修改个人资料
  - 修改密码
- 已完成会话模块基础能力：
  - 创建会话
  - 会话列表
  - 会话详情
  - 删除会话
  - 无知识库场景下的 SSE 发消息
  - 重生成 / 中断接口暂未实现
- 已完成知识库模块基础能力：
  - 知识库 CRUD
  - 文档列表 / 文档详情 / 删除
  - 重建索引任务创建
  - 文档上传
- 已完成任务模块第一阶段：
  - 通用任务模型
  - 创建 reindex / cleanup 任务
  - worker 轮询骨架
- 已补管理员、RAG、模型、配额、审计、对象存储抽象的骨架接口
- 已完成基础测试与构建验证：
  - `GOCACHE=/tmp/go-build go test ./...`
  - 认证包已有基础单元测试
- 已补本地运行环境辅助：
  - `backend/.env` / `.env.local` 自动加载
  - `backend/compose.dev.yaml` 开发数据库容器
  - `backend/scripts/migrate-up.sh` 迁移脚本
  - `backend/scripts/run-api.sh`、`backend/scripts/run-worker.sh`
  - `backend/Makefile`
  - `backend/README.md` 已补本地启动步骤
- 开发数据库宿主机默认端口已从 `5432` 调整为 `55432`，以避免与本机已有 PostgreSQL 冲突

## 仍待完成

- PDF 解析器接入
- 真实 embedding、向量检索、引用记录
- 消息重生成与中断
- 管理后台真实业务接口
- 审计、配额统计、provider 配置和系统参数管理
- 更完整的 worker 任务治理与管理端重试接口

## 下一阶段推荐顺序

1. 接入真实 embedding provider，替换当前占位零向量
2. 打通向量检索，把当前普通聊天扩展为 RAG 问答
3. 完成 PDF 解析与引用记录
4. 最后补 admin / quota / audit / settings

## 运行环境最短启动路径

1. `cp backend/.env.example backend/.env`
2. 修改 `backend/.env` 中的 `AUTH_JWT_SECRET`
3. `cd backend`
4. `podman-compose -f compose.dev.yaml up -d`
5. `./scripts/migrate-up.sh`
6. 新开终端执行 `./scripts/run-api.sh`
7. 再开一个终端执行 `./scripts/run-worker.sh`

## 2026-03-13 冒烟测试结果

- API 健康检查通过：`GET /api/v1/healthz` 返回 `200`
- 数据库连通性通过：本地 `psql` 可连接 `127.0.0.1:55432`
- 已执行通过的接口：
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `GET /api/v1/users/me`
  - `PUT /api/v1/users/me`
  - `POST /api/v1/knowledge-bases`
  - `GET /api/v1/knowledge-bases/{kbId}`
  - `PUT /api/v1/knowledge-bases/{kbId}`
  - `GET /api/v1/knowledge-bases`
  - `GET /api/v1/knowledge-bases/{kbId}/documents`
  - `POST /api/v1/knowledge-bases/{kbId}/reindex`
  - `POST /api/v1/sessions`
  - `GET /api/v1/sessions`
  - `GET /api/v1/sessions/{sessionId}`
  - `POST /api/v1/auth/refresh`
  - `PUT /api/v1/users/me/password`
  - `POST /api/v1/auth/logout`
- 已验证占位接口：
  - `POST /api/v1/knowledge-bases/{kbId}/documents` 返回 `501 Not Implemented`

## 本次测试数据

- 测试用户 ID：`b8809e55-3249-47da-9311-0cd700d90254`
- 测试知识库 ID：`9f107902-dca9-4a4e-b2be-625597dad41f`
- 测试任务 ID：`9106efa5-f7eb-44bb-a946-3167aadf387e`
- 测试会话 ID：`46188fe7-6c94-4826-a8ec-6dcd35afb43c`

## 测试备注

- 这轮测试脚本运行在 `zsh` 下，变量名误用了 shell 特殊变量 `USERNAME`，所以注册时实际使用的用户名是当前系统用户名 `kitra`；这不是后端 bug。
- 登录、刷新令牌、修改密码和重新登录都已通过。
- `refresh` 返回的 access token 与登录返回值相同，因为当前 JWT 载荷没有 `jti`，且两次签发发生在同一秒，导致 token 字符串相同；这不影响本阶段功能验证，但后续可考虑补 `jti` 提高可追踪性。

## 2026-03-13 AGENTS 约定

- 已在仓库根目录新增 `AGENTS.md`
- 该文件给 Codex 标注了关键文档位置、后端启动/测试方法，以及 `tmp/backend-notes.md` 的使用约定
- 明确约定：系统启动命令默认由用户手动执行，除非用户明确要求代为执行

## 2026-03-13 第二阶段实现进展

- 已完成本地对象存储实现，默认写入 `backend/data/storage`
- 已为配置系统补充存储相关环境变量：
  - `STORAGE_PROVIDER`
  - `STORAGE_BUCKET`
  - `STORAGE_LOCAL_ROOT`
  - `STORAGE_MAX_UPLOAD_BYTES`
- 已实现 `POST /api/v1/knowledge-bases/{kbId}/documents`
  - `multipart/form-data`
  - 文件大小校验
  - MIME / 扩展名校验
  - 同知识库内基于 `sha256` 的内容去重
  - 创建 `files`、`documents`、`tasks(document_ingest)` 记录
- 已实现 worker 第二阶段：
  - 事务抢占 runnable task，使用 `FOR UPDATE SKIP LOCKED`
  - `document_ingest` 实际执行
  - `knowledge_base_reindex` 为文档补发 ingest 任务
  - `resource_cleanup` 基础执行
- 已实现文本入库流程：
  - `txt`
  - `markdown`
  - `docx`
  - 文本规范化
  - 基于段落和窗口的简化切块
  - 写入 `document_chunks`
- 当前已知限制：
  - `pdf` 上传会成功入库并创建任务，但 worker 会将其标记为 `failed`，因为 PDF 解析尚未实现
  - 向量写入当前使用 `1536` 维零向量占位，目的是先打通上传和异步任务链路
- 本轮本地验证已通过：
  - `cd backend && GOCACHE=/tmp/go-build go test ./...`
  - `cd backend && GOCACHE=/tmp/go-build go build ./...`

## 2026-03-13 重启后第二阶段冒烟测试

- 已确认 `GET /api/v1/healthz` 返回 `200`
- 已确认上传链路通过：
  - 新测试用户：`4028a489-d52f-495a-8c81-2a628ef547eb`
  - 测试知识库：`8a42ae65-d6bd-4077-a70d-e62ec8999733`
  - 测试文档：`cff91e9a-a2c4-4639-81a5-088692a91a18`
  - 上传任务：`002d1a8f-c206-4ba4-a451-da0cbfa8ad78`
  - 文档状态轮询结果：`pending -> pending -> pending -> pending -> available`
  - 数据库确认：
    - `documents.status = available`
    - `documents.chunk_count = 1`
    - `tasks(document_ingest).status = succeeded`
    - 本地对象文件在删除前存在
- 已确认重复上传校验通过：
  - 同一知识库上传同内容文档返回 `409`
  - 业务码：`40903 duplicate document`
- 已确认删除与清理链路通过：
  - 删除接口返回 `200`
  - 删除后查询文档返回 `404`
  - 关联 `files.deleted_at` 已被置值
  - 本地对象文件已被 worker 删除
- 已确认重建索引任务通过：
  - 第二轮专门测试知识库：`9b1f1557-c018-4655-bef0-090891fecf9e`
  - 文档：`7a0e7983-c446-42c8-8ae1-c769efe02c8b`
  - reindex 任务：`a3188b85-4b77-4981-a46d-1de4b2995c7d`
  - 轮询状态：`pending -> pending -> pending -> pending -> pending -> succeeded`

## 2026-03-17 Prompt 调整

- 已将聊天默认系统提示词改为中文基线版本，不再使用 `You are a helpful assistant.`
- 聊天请求组装时现在会自动注入运行时日期时间信息：
  - 当前 UTC 时间
  - 当前服务端本地时间、时区名和 UTC 偏移
- 知识库相关的 system prompt 已改为中文，明确：
  - 命中检索时优先依据检索上下文回答
  - 上下文不足时可补充通用回答，但不得伪装成知识库事实
  - 未命中检索时不得声称答案来自知识库，也不得编造引用
- 已同步更新 `backend/.env.example` 中的 `AI_CHAT_SYSTEM_PROMPT`
- 本轮验证已通过：
  - `cd backend && GOCACHE=/tmp/go-build go test ./...`
  - 数据库确认该文档已有 `2` 条 `document_ingest` 任务记录，最终文档状态仍为 `available`

## 2026-03-13 OpenAPI 同步

- 已根据当前后端实现重写 `docs/openapi-v1-draft.yaml`
- 本次同步原则：以后端真实行为为准，而不是设计草案中的目标状态
- 已修正的重点包括：
  - 列表接口从 `data.list` 改为当前后端真实返回的 `data.items`
  - 知识库创建、详情、更新等接口的响应结构已改为当前后端真实结构
  - `PUT /users/me` 请求体已移除当前后端不支持的 `email`
  - 登录响应中的 `user` 结构已改为当前后端真实字段集合
  - 会话消息 SSE 接口描述会随着真实实现继续同步
  - 管理端接口已标注为当前返回 `501 Not Implemented`
  - 错误详情字段从 `reason` 改为当前后端真实字段 `message`
- 已使用本地 YAML 解析器校验通过

## 2026-03-13 当前阶段判断

- 当前后端可以视为完成了“基础业务骨架 + 文档入库异步链路”阶段
- 已真实可用的能力：
  - 认证与当前用户
  - 会话基础 CRUD
  - 知识库 CRUD
  - 文档上传、去重、入库、删除、重建索引
  - worker 异步任务消费与清理
- 已验证状态：
  - `go test ./...` 与 `go build ./...` 通过
  - 本地 API + worker 联调冒烟通过
  - OpenAPI 草案已同步到当前后端实现
- 当前后端最关键的未完成部分已经从“工程骨架”转为“RAG 主链路能力”：
  - 真实 embedding
  - pgvector 检索
  - SSE 流式问答
  - 消息重生成 / 中断
  - PDF 解析
- 如果按课程设计可演示优先级推进，推荐下一阶段顺序：
  1. 实现真实 embedding 与检索
  2. 打通聊天问答主链路
  3. 补 PDF
  4. 最后补 admin / audit / quota / settings

## 2026-03-13 无知识库聊天阶段

- 已按当前优先级切换到“先做基础 AI 聊天，再继续完整文档问答”的路线
- 已完成的聊天能力：
  - `POST /api/v1/sessions/{sessionId}/messages` 不再是占位接口
  - 已接入 DeepSeek Chat Completions 流式接口
  - 后端向前端输出真实 SSE 事件：`meta`、`delta`、`done`
  - 当前用户消息与 assistant 回复都会写入 `messages`
  - 聊天上下文会带最近若干条历史消息
  - SSE 心跳已加入，默认 `15s`
- 当前范围限制：
  - 仅支持 `knowledge_base_id IS NULL` 的普通会话
  - 若会话绑定知识库，发送消息会返回 `501`
  - `regenerate` 与 `stream/stop` 仍然保持未实现
- 新增环境变量：
  - `AI_PROVIDER`
  - `DEEPSEEK_API_KEY`
  - `DEEPSEEK_BASE_URL`
  - `AI_DEFAULT_CHAT_MODEL`
  - `AI_CHAT_SYSTEM_PROMPT`
  - `AI_CHAT_TIMEOUT`
  - `AI_MAX_HISTORY_MESSAGES`
  - `AI_CHAT_TEMPERATURE`
  - `AI_SSE_HEARTBEAT_INTERVAL`
- 本轮本地验证已通过：
  - `cd backend && GOCACHE=/tmp/go-build go test ./...`
  - `cd backend && GOCACHE=/tmp/go-build go build ./...`
- 本轮未做真实联网调用验证：
  - 因为当前执行环境没有配置可用的 `DEEPSEEK_API_KEY`
  - DeepSeek provider 通过了本地无网络单元测试，验证了请求组装、流式解析和错误映射

## 2026-03-13 SSE 联调问题修复

- 真实 `curl` 联调已跑到聊天接口
- 暴露出一个后端缺陷：
  - `POST /api/v1/sessions/{sessionId}/messages` 返回 `500 failed to initialize sse stream`
  - 根因是 `httpx.Logger` 中间件中的 `statusRecorder` 包装了 `ResponseWriter`，但没有继续实现 `http.Flusher`
  - 这会导致 SSE 初始化阶段对 `http.Flusher` 的类型断言失败
- 已完成修复：
  - 为 `statusRecorder` 补充 `Flush`
  - 顺手补充 `Hijack` / `Push` 透传，避免继续破坏底层 writer 能力
  - 新增 `backend/internal/platform/httpx/middleware_test.go` 回归测试
- 修复后本地验证已通过：
  - `cd backend && GOCACHE=/tmp/go-build go test ./...`
  - `cd backend && GOCACHE=/tmp/go-build go build ./...`
- 注意：
  - 要继续验证真实聊天流，需要用户重启当前 `run-api`

## 2026-03-13 DeepSeek 聊天真实联调通过

- 在修复 `http.Flusher` 透传问题并重启 API 后，已完成真实 `curl` 联调
- 联调路径：
  - 注册新用户
  - 登录获取 access token
  - 创建未绑定知识库的普通会话
  - 调用 `POST /api/v1/sessions/{sessionId}/messages`
  - 再调用 `GET /api/v1/sessions/{sessionId}` 校验消息持久化
- 本次测试数据：
  - 用户：`390ba885-4606-4c0f-84f8-8377aacb6d85`
  - 会话：`ed21c2b6-1a54-4882-ba36-fb415621e4e5`
  - assistant 消息：`55cd70eb-1cff-40ed-a25e-24b4b8f3ebf0`
- 已确认结果：
  - SSE 返回顺序正确：`meta -> delta -> done`
  - `meta.message_id` 与最终入库的 assistant 消息 ID 一致
  - 会话详情中消息数为 `2`
  - assistant 消息已落库，`status = completed`
  - `model_used = deepseek-chat`
  - token 用量已写入：
    - `prompt_tokens = 21`
    - `completion_tokens = 37`
    - `total_tokens = 58`
- 本轮可以确认：
  - “无知识库基础 AI 聊天 + DeepSeek API” 已经是可运行状态

## 2026-03-13 最小联调文档补充

- 已在 `backend/README.md` 中新增“最小联调示例”
- 内容覆盖：
  - 注册
  - 登录
  - 创建普通会话
  - `curl -N` 调用 SSE 聊天
  - 回查会话详情确认消息落库
- 该示例用于前后端联调和课程演示时快速验证当前 DeepSeek 聊天能力

## 2026-03-13 当前状态复核

- 当前仓库后端相关工作区状态干净，未发现新的后端未提交改动
- 再次执行通过：
  - `cd backend && GOCACHE=/tmp/go-build go test ./...`
- 当前后端完成度判断：
  - 已完成“账户认证 + 会话基础 + 知识库基础 CRUD + 文档上传/异步入库 + 无知识库 DeepSeek 流式聊天”
  - 已具备课程演示所需的基础后端主干
  - 尚未完成的核心能力仍集中在“知识库问答主链路”和“管理侧真实业务”
- 当前最重要未完成项：
  - 知识库绑定会话下的 RAG 聊天
  - 真实 embedding 与 pgvector 检索
  - 聊天 `regenerate` / `stream/stop`
  - PDF 解析
  - admin / audit / quota / settings 真实实现

## 2026-03-13 知识库会话 RAG 聊天实现完成

- 已完成 `POST /api/v1/sessions/{sessionId}/messages` 的知识库会话分支：
  - 会话绑定知识库时会先执行检索，再调用聊天模型
  - 命中知识库上下文时，SSE `meta.grounded = true`
  - 未命中时会自动回退为通用回答，SSE `meta.grounded = false`
- 已完成 RAG 检索链路：
  - 生成查询 embedding
  - 在 `document_chunks` 上执行 `pgvector` 检索
  - 应用知识库级 `retrieval_top_k` / `similarity_threshold`
  - 将命中的 chunk 拼装到模型提示词中
- 已完成 assistant 引用落库：
  - `message_citations` 会在 assistant 消息创建时一并写入
  - `GET /api/v1/sessions/{sessionId}` 的消息详情已新增 `citations`
- 已完成文档入库向量化：
  - worker 在 `document_ingest` 时不再写入零向量占位
  - 默认 embedding provider 改为本地 `local_hash`
  - 该 provider 生成与查询统一的 `1536` 维确定性向量，适合本地开发和课程演示
  - 同时已预留 `openai_compatible` embedding provider，可通过环境变量切换为远程真实 embedding 服务
- 新增/更新环境变量：
  - `AI_EMBEDDING_PROVIDER`
  - `AI_EMBEDDING_API_KEY`
  - `AI_EMBEDDING_BASE_URL`
  - `AI_EMBEDDING_TIMEOUT`
  - `AI_RAG_MAX_CONTEXT_CHUNKS`
- 已同步文档：
  - `backend/.env.example`
  - `backend/README.md`
  - `docs/openapi-v1-draft.yaml`
- 本轮本地验证已通过：
  - `cd backend && GOCACHE=/tmp/go-build go test ./...`
  - `cd backend && GOCACHE=/tmp/go-build go build ./...`
  - `python -c 'import yaml; yaml.safe_load(open("docs/openapi-v1-draft.yaml", "r", encoding="utf-8")); print("openapi yaml ok")'`

## 2026-03-13 当前剩余重点

- `POST /api/v1/sessions/{sessionId}/messages/{messageId}/regenerate`
- `POST /api/v1/sessions/{sessionId}/stream/stop`
- PDF 解析
- 若课程演示需要更高检索质量，可将 `AI_EMBEDDING_PROVIDER` 切换为 `openai_compatible` 并配置远程 embedding 服务
- admin / audit / quota / settings 真实业务实现

## 2026-03-13 知识库 RAG curl 冒烟测试通过

- 测试路径：
  - `GET /api/v1/healthz`
  - `POST /api/v1/auth/register`
  - `POST /api/v1/auth/login`
  - `POST /api/v1/knowledge-bases`
  - `POST /api/v1/knowledge-bases/{kbId}/documents`
  - `GET /api/v1/knowledge-bases/{kbId}/documents/{docId}`
  - `POST /api/v1/sessions`
  - `POST /api/v1/sessions/{sessionId}/messages`
  - `GET /api/v1/sessions/{sessionId}`
- 本轮测试数据：
  - 用户：`dff03018-4f7b-4d29-b3b4-38e3e657196a`
  - 用户名：`rag_smoke_1773414882`
  - 知识库：`c61603cf-e67b-4b43-b78c-712b512230c6`
  - 文档：`80616db0-1c22-4170-9d6d-c6352191fbdd`
  - 入库任务：`3a742bea-a9bb-461b-b171-f5cce900d193`
  - 会话：`9bd2e79b-2f4c-42e1-8b42-070cb6973610`
  - assistant 消息：`7e5c3b21-1b34-4a46-b01f-cb2492a6798c`
- 已确认结果：
  - 文档状态已到 `available`
  - SSE 首个事件：
    - `meta.message_id = 7e5c3b21-1b34-4a46-b01f-cb2492a6798c`
    - `meta.grounded = true`
    - `meta.model = deepseek-chat`
  - SSE 后续正常输出 `delta`，最终收到 `done`
  - assistant 回复成功落库，且 `grounded = true`
  - assistant 消息已写入引用：
    - `document_chunk_id = 25cc96d6-1530-46ce-94b3-2f30944a6292`
    - `document_id = 80616db0-1c22-4170-9d6d-c6352191fbdd`
    - `document_name = rag-smoke`
    - `rank_no = 1`
  - token 用量已写入：
    - `prompt_tokens = 155`
    - `completion_tokens = 60`
    - `total_tokens = 215`

## 2026-03-13 聊天闭环第二阶段完成

- 已完成 `POST /api/v1/sessions/{sessionId}/messages/{messageId}/regenerate`
  - 返回 `text/event-stream`
  - 复用原 user 消息重新执行检索与生成
  - 覆盖原 assistant 消息内容、token 用量和 `message_citations`
  - SSE `meta.message_id` 等于原 assistant 消息 ID
- 已完成 `POST /api/v1/sessions/{sessionId}/stream/stop`
  - 返回 `{"stopped": true|false}`
  - 若成功停止活跃流，不会落库不完整 assistant 答案
- 已为聊天模块补充流生命周期管理：
  - 同一会话同一时刻只允许一个活跃生成任务
  - 若同一会话已有活跃流，再次发送消息或重生成会返回 `409`
  - 主动 stop 时，原 SSE 会收到 `done.finish_reason = cancelled`
- 已同步更新：
  - `backend/README.md`
  - `docs/openapi-v1-draft.yaml`
- 本轮本地验证已通过：
  - `cd backend && GOCACHE=/tmp/go-build go test ./...`
  - `cd backend && GOCACHE=/tmp/go-build go build ./...`
  - OpenAPI YAML 解析通过

## 2026-03-13 当前剩余重点（更新）

- PDF 解析
- 若课程演示需要更高检索质量，可将 `AI_EMBEDDING_PROVIDER` 切换为 `openai_compatible`
- admin / audit / quota / settings 真实业务实现
- quota / audit 统计落库

## 2026-03-13 regenerate / stop curl 冒烟测试通过

- 本轮测试用户：
  - 用户：`0fc03aa0-765d-4f08-a0c2-45d7cfb22414`
  - 用户名：`regen_stop_smoke_1773416143`
- 本轮测试资源：
  - 知识库：`cffe0e2d-bee3-42bc-86b2-db4187821639`
  - 文档：`46b8dc9a-47ea-4128-b2e9-a9ecfeba2fec`
  - 知识库会话：`0a072f79-4e4c-4e6a-a904-6e881e073267`
  - 普通会话：`2ec21a31-3079-451e-84e6-467f9eeb57f4`
- `regenerate` 测试结果：
  - 原 assistant 消息：`d78a3173-665b-43e1-8446-d705aaaaf2b9`
  - `POST /api/v1/sessions/{sessionId}/messages/{messageId}/regenerate` 返回 SSE
  - `meta.message_id` 与原 assistant 消息 ID 一致
  - 回查会话详情后，assistant 仍是同一条消息 ID，但内容和 `updated_at` 已更新
  - 引用仍存在：
    - `document_chunk_id = 379099ef-f171-4aeb-b641-51ac2580e788`
    - `document_id = 46b8dc9a-47ea-4128-b2e9-a9ecfeba2fec`
    - `document_name = chat-stage3-smoke`
  - 更新后的 token 用量：
    - `prompt_tokens = 152`
    - `completion_tokens = 162`
    - `total_tokens = 314`
- `stream/stop` 测试结果：
  - 第一次 stop 测试时，正在生成的 assistant 消息 ID：`c4be7d3c-012d-4cac-9e50-cb58897095ea`
  - `POST /api/v1/sessions/{sessionId}/stream/stop` 返回 `{"stopped":true}`
  - 原 SSE 最终收到 `done.finish_reason = cancelled`
  - 回查会话详情后，仅有 user 消息，无 assistant 消息落库
- 并发互斥测试结果：
  - 第二次长生成过程中，assistant 临时消息 ID：`9ec29afa-cd79-434d-abf3-77152275c988`
  - 在该流未结束时再次 `POST /api/v1/sessions/{sessionId}/messages`
  - 后端返回 `409`
  - 错误消息：`a message is already being generated for this session`
  - 随后调用 `stream/stop` 成功停止该流，且回查会话详情仍未产生 assistant 落库

## 2026-03-16 PDF 文本提取接入

- 已在 `backend/internal/kb/ingest.go` 接入 Go 库 `github.com/ledongthuc/pdf`
- 当前 `pdf` 文本提取能力已支持：
  - 带可提取文本层的常规 PDF
  - 对部分厂商生成的宽松 PDF 头做兼容修正
- 已补充测试：
  - 最小可重复 PDF 构造测试
  - 宽松 PDF 头兼容测试
  - 通过环境变量注入真实 PDF 样本路径的手动 smoke test
- 已验证：
  - `cd backend && GOCACHE=/tmp/go-build go test ./internal/kb`
  - `cd backend && GOCACHE=/tmp/go-build go test ./...`
  - `cd backend && GOCACHE=/tmp/go-build go build ./...`
- 使用用户提供的真实样本：
  - `/run/media/kitra/win_par1/04_文档/2026/保研/计算机科学与工程学院2026年推荐优秀应届本科毕业生免试攻读硕士学位研究生工作细则.pdf`
  - 已确认该文件是扫描图像 PDF
  - `pdftotext` 诊断输出仅有换页符，没有任何正文文本
  - 当前 Go PDF 文本提取同样无法得到正文，因此会返回：
    - `pdf contains no extractable text; OCR is not implemented`
- 结论更新：
  - “PDF 解析未实现”已不再准确
  - 当前状态应改为：
    - 文本型 PDF 已支持
    - 扫描件/纯图片型 PDF 仍不支持，因为 OCR 尚未实现

## 2026-03-16 PDF 文本可读性修复

- 初版 `GetTextByRow()` 方案能提取出文本，但在真实英文论文 PDF 中会出现大量单词粘连，导致 chunk 可读性不足
- 已改为基于 `page.Content().Text` 的字符级坐标重建：
  - 按字符间距补空格
  - 按坐标变化补行分隔
  - 保留 ligature 归一化
- 已新增针对字符级重建的单元测试：
  - `TestRebuildPDFPageText`
- 已使用用户提供的文本型 PDF 做真实链路验证：
  - `/run/media/kitra/win_par1/04_文档/2026/文献/Informer Beyond Efficient Transformer for Long.pdf`
  - `ParseDocumentContent` 样本测试通过
  - 实际启动 `make run-api` + `make run-worker` 后上传成功
  - 文档状态：`available`
  - 任务状态：`succeeded`
  - 数据库中 `document_chunks` 前缀已恢复为可读文本，例如：
    - `Informer: Beyond Efficient Transformer for Long Sequence`
    - `Time-Series Forecasting`

## 2026-03-16 管理后台接口落地与冒烟

- 已将 `backend/internal/admin/handler.go` 从占位 `501` 改为真实路由处理，并新增：
  - `GET /api/v1/admin/overview`
  - `GET /api/v1/admin/users`
  - `POST /api/v1/admin/users/{userId}/freeze`
  - `POST /api/v1/admin/users/{userId}/unfreeze`
  - `GET /api/v1/admin/tasks`
  - `POST /api/v1/admin/tasks/{taskId}/retry`
  - `GET /api/v1/admin/provider-configs`
  - `GET /api/v1/admin/settings`
  - `GET /api/v1/admin/quota-policies`
  - `GET /api/v1/admin/audit-logs`
- 已新增：
  - `backend/internal/admin/model.go`
  - `backend/internal/admin/repository.go`
  - `backend/internal/admin/service.go`
- 已在 `backend/cmd/api/main.go` 接入真实 `admin` service / repository
- 当前仍保持未实现的接口为显式 `feature_not_ready`：
  - 创建/更新用户
  - 重置用户密码
  - 更新 provider 配置
  - 更新系统设置
  - 更新配额策略
  - 用户用量查询

### 本轮验证

- 已完成：
  - `cd backend && GOCACHE=/tmp/go-build go build ./...`
  - `cd backend && GOCACHE=/tmp/go-build go test ./...`
- 已重新启动 API 新进程做联调
  - 发现原 `:8080` 上仍是旧进程，返回旧版 `501/404`
  - 已切换到新编译进程后继续验证
- 已使用管理员账号 `admin / 12345678abc` 进行接口冒烟：
  - `GET /api/v1/admin/overview` 返回真实统计
  - `GET /api/v1/admin/users?page=1&size=10` 返回用户列表
  - `GET /api/v1/admin/tasks?page=1&size=10` 返回任务列表
  - `GET /api/v1/admin/provider-configs` 返回空列表
  - `GET /api/v1/admin/settings` 返回 10 条系统设置
  - `GET /api/v1/admin/quota-policies` 返回空列表
  - `GET /api/v1/admin/audit-logs?page=1&size=10` 在前端冻结/恢复后返回新增审计记录
- 已通过前端真实操作间接验证：
  - `POST /api/v1/admin/users/{userId}/freeze`
  - `POST /api/v1/admin/users/{userId}/unfreeze`
  - 两次操作均成功，且审计日志写入正常

## 2026-03-19 管理后台 API Key 配置补齐

### 本轮实现

- 已新增 `backend/internal/providerconfig/`：
  - 统一封装 `provider_configs` 表读写
  - `GET /admin/provider-configs` 现固定只返回默认 provider：`DeepSeek`
  - 返回仅暴露 `has_api_key` 与 `api_key_source(environment/database/missing)`，不回显真实 key
- 已将 `PUT /api/v1/admin/provider-configs/{provider}` 从 `feature_not_ready` 改为真实实现：
  - 请求体仅接受 `api_key`
  - 必须重新输入非空 key 才能覆盖
  - 不支持读取旧值或回显旧值
  - 更新后写入 `admin.provider.update` 审计日志
- 已将 API 与 worker 的 provider key 解析改为“`.env` 优先，数据库后备”：
  - 聊天链路：`backend/internal/chat/dynamic_provider.go`
  - embedding 链路：`backend/internal/model/dynamic_provider.go`
  - 若 `.env` 中已提供 key，则直接视为已配置并优先使用
  - 仅当 `.env` 未提供 key 时，才回退读取数据库中的 key
  - 若两者都无 key，则仍按 misconfigured 返回，需在管理后台补配置
- 当前后端仍未实现“后台编辑 base url / 默认模型 / enable 状态”；本轮只补 API key 覆盖写入，满足当前诉求

### 本轮验证

- 已完成：
  - `cd backend && GOCACHE=/tmp/go-build go test ./...`
- 已对本机现有 `127.0.0.1:8080` 做只读冒烟：
  - `GET /api/v1/healthz` 返回 `200`
  - 使用管理员 `admin / 12345678abc` 登录成功
  - 当前运行中的旧 API 进程 `GET /api/v1/admin/provider-configs` 仍返回空列表，说明需要重启 API 才会加载本轮新逻辑
- 本轮未直接调用 `PUT /api/v1/admin/provider-configs/{provider}` 做在线写入验证，避免把测试 key 写入用户当前数据库
