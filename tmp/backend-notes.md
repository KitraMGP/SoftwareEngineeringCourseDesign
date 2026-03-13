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
- SSE 问答、文档上传、解析、向量检索和实际 worker 消费留到后续阶段；本轮先放好接口和占位逻辑。
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
  - SSE 发消息 / 重生成 / 中断接口先返回占位
- 已完成知识库模块基础能力：
  - 知识库 CRUD
  - 文档列表 / 文档详情 / 删除
  - 重建索引任务创建
  - 上传接口先返回占位
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
- SSE 流式回答主链路
- 消息重生成与中断
- 管理后台真实业务接口
- 审计、配额统计、provider 配置和系统参数管理
- 更完整的 worker 任务治理与管理端重试接口

## 下一阶段推荐顺序

1. 接入真实 embedding provider，替换当前占位零向量
2. 打通向量检索 + SSE 问答主链路
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
  - 数据库确认该文档已有 `2` 条 `document_ingest` 任务记录，最终文档状态仍为 `available`
