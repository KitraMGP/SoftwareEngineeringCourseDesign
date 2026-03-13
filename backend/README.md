# Backend

本目录是“基于大模型的私有知识库问答系统”后端当前后端实现。

## 当前已落地

- `cmd/api`: API 入口
- `cmd/worker`: Worker 入口
- `internal/platform`: 配置、数据库连接、HTTP 响应、中间件、鉴权、本地文件存储
- `internal/account`: 注册、登录、刷新令牌、退出登录、当前用户信息、修改密码
- `internal/chat`: 会话创建、列表、详情、删除；无知识库场景下的 DeepSeek SSE 聊天
- `internal/kb`: 知识库 CRUD、文档上传、文档列表/详情/删除、重建索引任务创建
- `internal/task`: 统一任务模型、任务创建与状态推进
- `internal/worker`: 异步任务消费执行器
- `internal/admin`: 管理端路由占位
- `migrations`: Goose 迁移 SQL

## 启动前准备

1. 复制环境变量模板

```bash
cp backend/.env.example backend/.env
```

说明：

- 默认 `DATABASE_DSN` 已经和 `compose.dev.yaml` 里的 PostgreSQL 配置对齐
- 当前默认宿主机端口为 `55432`
- 如果你不改端口、用户名、密码和数据库名，那么可以直接使用默认值
- 现在 API/Worker 会自动读取 `backend/.env` 和 `backend/.env.local`，不需要手动 `source .env`

2. 至少确认以下配置

- `DATABASE_DSN`
- `AUTH_JWT_SECRET`
- `DEEPSEEK_API_KEY`
- `STORAGE_PROVIDER`
- `STORAGE_BUCKET`
- `STORAGE_LOCAL_ROOT`

3. 准备 PostgreSQL 15+，推荐直接用仓库内的开发容器

Podman Compose:

```bash
cd backend
podman-compose -f compose.dev.yaml up -d
```

Docker Compose:

```bash
cd backend
docker compose -f compose.dev.yaml up -d
```

说明：

- 容器镜像已内置 `pgvector`
- migration 会负责创建 `pgcrypto`、`vector` 扩展
- 默认上传文件会保存在 `backend/data/storage`，目录会自动创建

4. 执行数据库迁移

```bash
cd backend
./scripts/migrate-up.sh
```

如果你更习惯 `make`：

```bash
cd backend
make db-up
make migrate-up
```

## 本地运行

API:

```bash
cd backend
./scripts/run-api.sh
```

Worker:

```bash
cd backend
./scripts/run-worker.sh
```

如果你更习惯 `make`：

```bash
cd backend
make run-api
make run-worker
```

## 测试

```bash
cd backend
GOCACHE=/tmp/go-build go test ./...
```

## 当前基础 AI 聊天能力

- 聊天接口：`POST /api/v1/sessions/{sessionId}/messages`
- 返回类型：`text/event-stream`
- 当前已接入 provider：`DeepSeek`
- 当前支持场景：
  - 会话未绑定知识库时的通用聊天
  - 历史消息自动截取最近若干条拼接到上下文
  - 用户消息和 assistant 消息落库
- 当前限制：
  - 若会话绑定了 `knowledge_base_id`，发送消息会返回 `501`
  - `regenerate` 和 `stream/stop` 仍未实现

建议至少补齐以下环境变量：

- `DEEPSEEK_API_KEY`
- `AI_DEFAULT_CHAT_MODEL`
- `AI_CHAT_TIMEOUT`
- `AI_MAX_HISTORY_MESSAGES`
- `AI_SSE_HEARTBEAT_INTERVAL`

### 最小联调示例

下面示例展示“注册 -> 登录 -> 创建普通会话 -> 发起流式聊天”的最短链路。

1. 注册

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo_user",
    "email": "demo_user@example.com",
    "password": "StrongPass123!"
  }'
```

2. 登录并保存 `access_token`

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "account": "demo_user",
    "password": "StrongPass123!"
  }'
```

登录成功后，从返回 JSON 的 `data.access_token` 中取出令牌。

3. 创建一个不绑定知识库的普通会话

```bash
curl -X POST http://127.0.0.1:8080/api/v1/sessions \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "DeepSeek Curl Smoke",
    "model": "deepseek-chat"
  }'
```

创建成功后，从返回 JSON 的 `data.session_id` 中取出会话 ID。

4. 发起 SSE 聊天

```bash
curl -N -X POST http://127.0.0.1:8080/api/v1/sessions/<session_id>/messages \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "你好，请用三句话做一个简短的自我介绍。"
  }'
```

预期返回示例：

```text
event: meta
data: {"message_id":"uuid","grounded":false,"model":"deepseek-chat"}

event: delta
data: {"content":"你好"}

event: delta
data: {"content":"！我是 DeepSeek ..."}

event: done
data: {"finish_reason":"stop"}
```

5. 可选：回查会话详情，确认消息已经落库

```bash
curl http://127.0.0.1:8080/api/v1/sessions/<session_id> \
  -H "Authorization: Bearer <access_token>"
```

## 当前文档上传能力

- 上传接口：`POST /api/v1/knowledge-bases/{kbId}/documents`
- 请求格式：`multipart/form-data`，字段名固定为 `file`
- 上传大小上限默认 `20 MiB`，可通过 `STORAGE_MAX_UPLOAD_BYTES` 调整
- 当前允许上传的 MIME / 扩展名：
  - `text/plain` / `.txt`
  - `text/markdown` / `.md` / `.markdown`
  - `application/vnd.openxmlformats-officedocument.wordprocessingml.document` / `.docx`
  - `application/pdf` / `.pdf`
- 当前 worker 已实现：
  - `document_ingest`
  - `knowledge_base_reindex`
  - `resource_cleanup`
- 当前 ingest 已实现的文本提取：
  - `txt`
  - `markdown`
  - `docx`
- 当前限制：
  - `pdf` 上传会被接受，但 worker 目前会把该文档任务标记为 `failed`，因为 PDF 解析器尚未接入
  - embedding 目前仍是占位零向量，用于先打通上传、任务、分块和状态流转

## 下一阶段建议

1. 将当前无知识库聊天扩展为带检索的 RAG 问答
2. 替换占位 embedding，实现真实向量化与 pgvector 检索
3. 完成消息重生成与中断
4. 完成管理员的用户、任务、系统参数、配额和审计接口
