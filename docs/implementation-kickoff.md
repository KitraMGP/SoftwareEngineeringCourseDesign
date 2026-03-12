# Implementation Kickoff

本文件用于把设计文档、数据库迁移和 OpenAPI 草案串起来，方便新会话直接进入系统实现。

## 1. 资料优先级

后续实现时，按以下顺序读取资料：

1. `docs/detailed-design.md`
2. `docs/migrations/`
3. `docs/openapi-v1-draft.yaml`
4. `docs/schema-ddl-draft.sql`
5. `docs/design.md`

说明：

- `docs/detailed-design.md` 是当前唯一主设计文档
- `docs/schema-ddl-draft.sql` 是总览草案，真正的版本化结构以 `docs/migrations/` 为准
- `docs/design.md` 只作历史草稿参考

## 2. 默认工程选型

如果后续新会话未单独确认技术栈，默认采用以下组合：

- Go：`1.24+`
- HTTP 路由：`chi`
- 数据访问：`pgx + sqlc`
- Migration：`goose`
- 日志：标准库 `slog`
- 对象存储：S3 兼容接口，开发环境 `MinIO`
- 数据库：`PostgreSQL 15+` with `pgvector`

## 3. 建议实现顺序

1. 初始化 Go 模块和目录结构
2. 接入 `goose` 并执行 `docs/migrations/`
3. 基于 `docs/openapi-v1-draft.yaml` 固化 handler 层接口签名
4. 优先实现账户与认证模块
5. 实现会话、消息、SSE 主链路
6. 实现知识库上传、任务系统和对象存储接入
7. 实现文档解析、分块、embedding 和检索
8. 实现管理员任务、配额和系统参数接口

## 4. Migration 说明

当前 migration 采用 Goose 单文件 up/down 风格：

- `00001_extensions_and_functions.sql`
- `00002_core_schema.sql`
- `00003_indexes_and_triggers.sql`
- `00004_seed_system_settings.sql`

建议执行顺序由 Goose 自动保证。首期需要特别留意：

- `VECTOR(1536)` 需要与最终 embedding 模型维度一致
- 如后续切换为不同维度模型，需要先调整设计，而不是直接改代码
- `00004_seed_system_settings.sql` 写入的是实现默认值，不是最终生产配置

## 5. OpenAPI 使用建议

`docs/openapi-v1-draft.yaml` 目前适合做三件事：

- 作为前后端联调契约
- 生成前端类型定义
- 约束 handler 的请求和响应结构

建议做法：

- 在正式生成代码前，先复核所有错误码映射是否与 `docs/detailed-design.md` 一致
- 流式接口不要强求完整代码生成，可手写 SSE handler
- 生成的 DTO 不要直接等同于数据库模型

## 6. 仍需在实现前确认的少量工程细节

这些问题不影响继续设计，但进入编码前最好在新会话里固定下来：

- PDF 解析库选型
- DOCX 解析库选型
- 本地开发环境的 `docker compose` 方案
- `sqlc` 包输出目录和命名约定
- refresh token 在前后端部署形态下使用 Cookie 还是显式传输

如果未单独确认，建议优先做最小可用方案，并把结论回写到 `docs/detailed-design.md`。
