# 基于大模型的知识问答系统详细设计文档

## 1. 详细功能描述

本系统面向普通用户和管理员两类角色，提供用户管理、AI会话管理、问答交互、知识库管理和系统管理五大核心功能。以下按模块描述主要业务流程。

### 1.1 用户管理模块

**1.1.1 注册流程**  
- 用户访问系统注册页面，填写用户名、密码、邮箱等必要信息。  
- 前端提交注册请求，后端验证信息格式（如密码强度、邮箱唯一性）。  
- 验证通过后，系统创建新用户账户，状态默认为正常，返回注册成功提示。

**1.1.2 登录流程**  
- 用户输入账号（用户名/邮箱）和密码，前端提交登录请求。  
- 后端验证凭证，若正确则生成 JWT 令牌返回给客户端。  
- 客户端保存令牌，后续请求携带在 Authorization 头中。

**1.1.3 个人信息管理**  
- 用户可查看、修改个人信息（昵称、邮箱、头像等），修改时需验证新信息的有效性。  
- 修改密码需验证原密码，新密码符合复杂度要求后更新。

**1.1.4 退出登录**  
- 客户端清除本地令牌，后续请求失效；服务端无需额外处理（无状态 JWT）。

### 1.2 AI会话管理模块

**1.2.1 会话列表**  
- 用户进入会话管理页面，系统查询该用户所有会话，按最后更新时间倒序排列。  
- 支持按会话名称或关键词搜索会话。

**1.2.2 会话切换与查看**  
- 用户点击某个会话，系统加载该会话的历史消息列表（按时间正序显示）。  
- 会话界面顶部显示会话名称、当前使用的模型、知识库等配置信息。

**1.2.3 会话删除**  
- 用户可删除单个或多个会话，删除后相关消息一并清除，并更新会话列表。

**1.2.4 会话导出**  
- 用户导出会话内容，系统生成 Markdown 或 JSON 格式文件供下载，包含用户问题和系统回答。

### 1.3 问答交互模块

**1.3.1 开始新会话**  
- 用户在问答页面点击“新建会话”，系统创建空白会话并进入该会话界面。  
- 新会话默认使用系统全局默认模型和知识库（若无则无知识库）。

**1.3.2 发送消息与流式回答**  
- 用户在会话输入框中输入问题，可选上传文件、选择知识库、切换模型。  
- 前端通过 POST 请求发送消息内容及配置参数。  
- 后端启动大模型调用，通过 Server-Sent Events (SSE) 流式返回生成的文本片段，前端实时渲染。  
- 回答完成后，系统保存完整的问答对（用户问题和模型回答）到数据库。

**1.3.3 中断回答**  
- 在流式生成过程中，用户可点击“停止”按钮，前端发送中断请求，后端终止模型生成任务，保留已生成的部分。

**1.3.4 重新生成回答**  
- 对于已有回答的对话，用户可触发重新生成，系统将使用相同的问题和配置再次调用模型，生成新回答覆盖原回答（或保留多个版本？需求中未明确，可设计为覆盖或新增一条回答记录，但通常为覆盖当前回答）。

**1.3.5 上传文件（扩展）**  
- 用户可上传文件（如 PDF、TXT、Word）作为上下文。后端接收文件，解析文本内容，并与问题一起提交给模型（或先进行知识库检索）。

**1.3.6 使用知识库（扩展）**  
- 用户可在发送消息前选择一个或多个知识库，系统将问题与知识库中的相关文档片段结合，增强模型回答。

**1.3.7 切换模型（扩展）**  
- 用户可在会话中随时切换大模型（如 DeepSeek-V2, DeepSeek-R1 等），后续消息使用新模型回答。

### 1.4 知识库管理模块

**1.4.1 知识库维护**  
- 用户可创建知识库，填写名称、描述等基本信息。  
- 可查看、修改、删除自己创建的知识库。

**1.4.2 文档管理**  
- 进入某个知识库后，可上传文档（支持 PDF、TXT、Markdown、Word 等格式）。  
- 后端对文档进行文本提取、分块、向量化，并存入向量数据库（如 PostgreSQL 的 pgvector 扩展）。  
- 用户可查看文档列表、删除文档。删除文档时同步移除向量索引。

**1.4.3 重建索引**  
- 当文档更新或向量化算法改进时，用户可触发重建索引，系统重新处理所有文档，更新向量库。

### 1.5 系统管理模块（仅管理员）

**1.5.1 API 密钥配置**  
- 管理员可配置大模型服务的 API 密钥（如 DeepSeek 的 API Key），系统调用模型时使用。

**1.5.2 系统状态查看**  
- 展示系统运行状态：CPU/内存使用率、API 调用次数、最近错误数等。

**1.5.3 系统运行日志**  
- 提供日志查询界面，支持按级别（INFO、ERROR）、时间范围筛选。

**1.5.4 用户账户管理**  
- 管理员可查看所有用户列表，支持新建用户、编辑用户信息、重置密码、冻结/解冻账户、删除用户。

**1.5.5 系统参数配置**  
- 配置全局默认模型、知识库检索参数（如 top_k）、会话超时时间等。

---

## 2. 接口设计 (RESTful API)

所有接口均以 `/api/v1` 为前缀，使用 JSON 格式传输，认证通过 JWT 实现（除登录、注册外）。

### 2.1 用户管理

| 端点 | 方法 | 描述 | 请求参数 | 响应 |
|------|------|------|----------|------|
| `/auth/register` | POST | 用户注册 | `{username, password, email}` | `{code, message, data: {user_id}}` |
| `/auth/login` | POST | 用户登录 | `{username, password}` | `{code, message, data: {token, user_info}}` |
| `/auth/logout` | POST | 退出登录 | 无（客户端丢弃 token） | `{code, message}` |
| `/users/me` | GET | 获取当前用户信息 | 无 | `{code, data: {id, username, email, avatar, created_at}}` |
| `/users/me` | PUT | 修改个人信息 | `{nickname?, email?, avatar?}` | `{code, message}` |
| `/users/me/password` | PUT | 修改密码 | `{old_password, new_password}` | `{code, message}` |

### 2.2 会话管理

| 端点 | 方法 | 描述 | 请求参数 | 响应 |
|------|------|------|----------|------|
| `/sessions` | GET | 获取当前用户会话列表 | `?keyword=搜索词&page=1&size=20` | `{code, data: {total, list: [{id, name, created_at, updated_at}]}}` |
| `/sessions` | POST | 创建新会话 | `{name?}` | `{code, data: {session_id}}` |
| `/sessions/{sessionId}` | GET | 获取会话详情（含消息） | 无 | `{code, data: {session, messages: [{role, content, created_at}]}}` |
| `/sessions/{sessionId}` | DELETE | 删除会话 | 无 | `{code, message}` |
| `/sessions/{sessionId}/export` | GET | 导出会话内容 | `?format=markdown` | 文件下载 |
| `/sessions/search` | GET | 搜索会话 | `?q=关键词` | 同列表接口 |

### 2.3 问答交互

| 端点 | 方法 | 描述 | 请求参数 | 响应 |
|------|------|------|----------|------|
| `/sessions/{sessionId}/messages` | POST | 发送消息（流式） | `{content, files?: [file_id], knowledge_base_ids?: [uuid], model?: string}` | 使用 SSE 流式返回文本片段 |
| `/sessions/{sessionId}/messages/{messageId}/regenerate` | POST | 重新生成回答 | 无 | 同发送消息流式响应 |
| `/sessions/{sessionId}/generate/stop` | POST | 中断当前回答 | 无 | `{code, message}` |
| `/files` | POST | 上传文件（用于问答上下文） | multipart/form-data: file | `{code, data: {file_id, name, size}}` |

### 2.4 知识库管理

| 端点 | 方法 | 描述 | 请求参数 | 响应 |
|------|------|------|----------|------|
| `/knowledge-bases` | GET | 获取知识库列表 | `?page=1&size=20` | `{code, data: {total, list: [{id, name, description, created_at}]}}` |
| `/knowledge-bases` | POST | 创建知识库 | `{name, description}` | `{code, data: {id}}` |
| `/knowledge-bases/{kbId}` | GET | 获取知识库详情 | 无 | `{code, data: {id, name, description, created_at, documents: []}}` |
| `/knowledge-bases/{kbId}` | PUT | 修改知识库 | `{name?, description?}` | `{code, message}` |
| `/knowledge-bases/{kbId}` | DELETE | 删除知识库 | 无 | `{code, message}` |
| `/knowledge-bases/{kbId}/documents` | GET | 获取文档列表 | `?page=1&size=20` | `{code, data: {total, list: [{id, filename, status, created_at}]}}` |
| `/knowledge-bases/{kbId}/documents` | POST | 上传文档 | multipart/form-data: file | `{code, data: {document_id}}` |
| `/knowledge-bases/{kbId}/documents/{docId}` | DELETE | 删除文档 | 无 | `{code, message}` |
| `/knowledge-bases/{kbId}/index/rebuild` | POST | 重建知识库索引 | 无 | `{code, message}`（异步任务，可返回任务 ID） |

### 2.5 系统管理（管理员）

| 端点 | 方法 | 描述 | 请求参数 | 响应 |
|------|------|------|----------|------|
| `/admin/api-keys` | GET | 获取当前 API 密钥配置 | 无 | `{code, data: {provider, key_preview}}` |
| `/admin/api-keys` | PUT | 配置 API 密钥 | `{provider, api_key}` | `{code, message}` |
| `/admin/status` | GET | 获取系统状态 | 无 | `{code, data: {cpu, memory, api_calls_today, errors_today}}` |
| `/admin/logs` | GET | 获取系统日志 | `?level=ERROR&start=2025-01-01&end=2025-01-02` | `{code, data: {logs: [{timestamp, level, message}]}}` |
| `/admin/users` | GET | 获取所有用户 | `?page=1&size=20&role=user` | `{code, data: {total, list: [...]}}` |
| `/admin/users` | POST | 新建用户 | `{username, password, email, role}` | `{code, data: {id}}` |
| `/admin/users/{userId}` | PUT | 修改用户信息 | `{username?, email?, role?, status?}` | `{code, message}` |
| `/admin/users/{userId}/password` | PUT | 重置密码 | `{new_password}` | `{code, message}` |
| `/admin/users/{userId}/freeze` | POST | 冻结账户 | 无 | `{code, message}` |
| `/admin/users/{userId}/unfreeze` | POST | 解冻账户 | 无 | `{code, message}` |
| `/admin/users/{userId}` | DELETE | 删除用户 | 无 | `{code, message}` |
| `/admin/settings` | GET | 获取系统参数 | 无 | `{code, data: {key: value}}` |
| `/admin/settings` | PUT | 更新系统参数 | `{settings: {...}}` | `{code, message}` |

---

## 3. 数据库设计

使用 PostgreSQL 15+，启用 `uuid-ossp`、`pgvector` 扩展（用于向量存储）。遵循以下设计原则：
- 主键统一使用 `UUID` 类型，默认 `gen_random_uuid()`。
- 时间戳使用 `TIMESTAMPTZ` 类型，默认 `CURRENT_TIMESTAMP`。
- 软删除使用 `deleted_at TIMESTAMPTZ` 字段（可为空）。
- 外键约束使用 `ON DELETE CASCADE` 或 `SET NULL` 根据业务决定。

### 3.1 表结构

#### 3.1.1 用户表 (users)

| 字段名        | 类型          | 约束                      | 说明                          |
|---------------|---------------|---------------------------|-------------------------------|
| id            | UUID          | PRIMARY KEY DEFAULT gen_random_uuid() | 用户ID                        |
| username      | TEXT          | UNIQUE NOT NULL           | 用户名                        |
| email         | TEXT          | UNIQUE NOT NULL           | 邮箱                          |
| password_hash | TEXT          | NOT NULL                  | 加密后的密码                  |
| nickname      | TEXT          |                           | 昵称                          |
| avatar        | TEXT          |                           | 头像 URL                      |
| role          | TEXT          | NOT NULL DEFAULT 'user'   | 角色：user / admin            |
| status        | TEXT          | NOT NULL DEFAULT 'active' | 状态：active / frozen         |
| created_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 创建时间                      |
| updated_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 更新时间                      |
| deleted_at    | TIMESTAMPTZ   |                           | 软删除时间                    |

索引：
- `username`, `email` (唯一)
- `role`, `status`

#### 3.1.2 会话表 (sessions)

| 字段名        | 类型          | 约束                      | 说明                          |
|---------------|---------------|---------------------------|-------------------------------|
| id            | UUID          | PRIMARY KEY DEFAULT gen_random_uuid() | 会话ID                        |
| user_id       | UUID          | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 所属用户                      |
| name          | TEXT          |                           | 会话名称（可为空，自动生成）  |
| model         | TEXT          |                           | 当前使用的模型标识            |
| knowledge_base_ids | UUID[]   |                           | 关联的知识库ID列表（数组）    |
| created_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 创建时间                      |
| updated_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 最后活动时间                  |
| deleted_at    | TIMESTAMPTZ   |                           | 软删除时间                    |

索引：
- `user_id`, `updated_at` (倒序)
- `name` 支持全文搜索（可使用 GIN 索引）

#### 3.1.3 消息表 (messages)

| 字段名        | 类型          | 约束                      | 说明                          |
|---------------|---------------|---------------------------|-------------------------------|
| id            | UUID          | PRIMARY KEY DEFAULT gen_random_uuid() | 消息ID                        |
| session_id    | UUID          | NOT NULL REFERENCES sessions(id) ON DELETE CASCADE | 所属会话                      |
| role          | TEXT          | NOT NULL                  | 角色：user / assistant        |
| content       | TEXT          | NOT NULL                  | 消息内容                      |
| files         | JSONB         |                           | 关联的文件信息（如ID列表）    |
| model_used    | TEXT          |                           | 生成此回答使用的模型（assistant时）|
| tokens        | INTEGER       |                           | 消耗的 token 数（可选）       |
| created_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 创建时间                      |

索引：
- `session_id`, `created_at`
- 会话内消息顺序由 `created_at` 决定

#### 3.1.4 文件表 (files)

| 字段名        | 类型          | 约束                      | 说明                          |
|---------------|---------------|---------------------------|-------------------------------|
| id            | UUID          | PRIMARY KEY DEFAULT gen_random_uuid() | 文件ID                        |
| user_id       | UUID          | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 上传者                        |
| filename      | TEXT          | NOT NULL                  | 原始文件名                    |
| storage_path  | TEXT          | NOT NULL                  | 存储路径（如对象存储 key）    |
| size          | BIGINT        | NOT NULL                  | 文件大小（字节）              |
| mime_type     | TEXT          |                           | MIME 类型                     |
| created_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 上传时间                      |
| deleted_at    | TIMESTAMPTZ   |                           | 软删除时间                    |

#### 3.1.5 知识库表 (knowledge_bases)

| 字段名        | 类型          | 约束                      | 说明                          |
|---------------|---------------|---------------------------|-------------------------------|
| id            | UUID          | PRIMARY KEY DEFAULT gen_random_uuid() | 知识库ID                      |
| user_id       | UUID          | NOT NULL REFERENCES users(id) ON DELETE CASCADE | 创建者                        |
| name          | TEXT          | NOT NULL                  | 知识库名称                    |
| description   | TEXT          |                           | 描述                          |
| embedding_model | TEXT        |                           | 使用的向量化模型              |
| created_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 创建时间                      |
| updated_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 更新时间                      |
| deleted_at    | TIMESTAMPTZ   |                           | 软删除时间                    |

#### 3.1.6 文档表 (documents)

| 字段名        | 类型          | 约束                      | 说明                          |
|---------------|---------------|---------------------------|-------------------------------|
| id            | UUID          | PRIMARY KEY DEFAULT gen_random_uuid() | 文档ID                        |
| knowledge_base_id | UUID    | NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE | 所属知识库                    |
| file_id       | UUID          | REFERENCES files(id) ON DELETE SET NULL | 关联的文件（可为空，若直接文本）|
| title         | TEXT          |                           | 文档标题（从内容提取）        |
| content       | TEXT          | NOT NULL                  | 提取的纯文本内容              |
| chunk_count   | INTEGER       |                           | 分块数量                      |
| status        | TEXT          | NOT NULL DEFAULT 'pending'| 处理状态：pending, processing, completed, failed |
| created_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 创建时间                      |
| deleted_at    | TIMESTAMPTZ   |                           | 软删除时间                    |

#### 3.1.7 文档块表 (document_chunks)

此表用于存储分块后的文本及其向量，使用 pgvector。

| 字段名        | 类型          | 约束                      | 说明                          |
|---------------|---------------|---------------------------|-------------------------------|
| id            | UUID          | PRIMARY KEY DEFAULT gen_random_uuid() | 块ID                         |
| document_id   | UUID          | NOT NULL REFERENCES documents(id) ON DELETE CASCADE | 所属文档                      |
| chunk_index   | INTEGER       | NOT NULL                  | 块序号                        |
| content       | TEXT          | NOT NULL                  | 块文本                        |
| embedding     | VECTOR(1536)  |                           | 向量（维度视模型而定，此处示例1536）|
| created_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 创建时间                      |

索引：
- `document_id`, `chunk_index` 唯一
- 向量索引：`CREATE INDEX ON document_chunks USING ivfflat (embedding vector_cosine_ops)` 或 HNSW

#### 3.1.8 API 密钥表 (api_keys)

| 字段名        | 类型          | 约束                      | 说明                          |
|---------------|---------------|---------------------------|-------------------------------|
| id            | UUID          | PRIMARY KEY DEFAULT gen_random_uuid() | 主键                          |
| provider      | TEXT          | UNIQUE NOT NULL           | 服务商：deepseek, openai 等   |
| encrypted_key | TEXT          | NOT NULL                  | 加密后的 API Key              |
| created_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 创建时间                      |
| updated_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 更新时间                      |

#### 3.1.9 系统参数表 (system_settings)

| 字段名        | 类型          | 约束                      | 说明                          |
|---------------|---------------|---------------------------|-------------------------------|
| key           | TEXT          | PRIMARY KEY               | 参数键                        |
| value         | JSONB         | NOT NULL                  | 参数值（支持各种类型）        |
| description   | TEXT          |                           | 描述                          |
| updated_at    | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 更新时间                      |

示例参数：
- `default_model`: `"deepseek-v2"`
- `knowledge_base_retrieval_top_k`: `5`
- `session_timeout_minutes`: `30`

#### 3.1.10 系统日志表 (system_logs)

| 字段名        | 类型          | 约束                      | 说明                          |
|---------------|---------------|---------------------------|-------------------------------|
| id            | BIGSERIAL     | PRIMARY KEY               | 自增ID                        |
| timestamp     | TIMESTAMPTZ   | NOT NULL DEFAULT NOW()    | 日志时间                      |
| level         | TEXT          | NOT NULL                  | 级别：DEBUG, INFO, WARN, ERROR |
| message       | TEXT          | NOT NULL                  | 日志内容                      |
| metadata      | JSONB         |                           | 额外结构化信息                |

索引：
- `timestamp`, `level`

### 3.2 关系说明

- 用户与会话：一对多（用户可拥有多个会话）
- 会话与消息：一对多
- 用户与文件：一对多（用户上传的文件）
- 用户与知识库：一对多（用户创建的知识库）
- 知识库与文档：一对多
- 文档与文档块：一对多
- 文件与文档：可选关联（一个文件可对应一个文档，一个文档也可没有文件，如直接输入文本）

### 3.3 注意事项

- 使用 `UUID` 作为主键避免暴露自增 ID 带来的安全风险，且便于分布式生成。
- 时间戳统一使用带时区的 `TIMESTAMPTZ`，存储 UTC 时间，应用层按需转换。
- PostgreSQL 中字符串字段优先使用 `TEXT`；长度限制和枚举限制建议通过 `CHECK` 约束而不是 `VARCHAR(n)` 表达。
- 数组类型 `UUID[]` 用于存储会话关联的知识库 ID，简化多对多关系（若知识库与会话是多对多，也可使用关联表，但需求中一个会话可使用多个知识库，用数组更简单）。
- 向量字段使用 `VECTOR(n)`，需提前安装 pgvector 扩展，维度根据选择的嵌入模型确定（如 1536、1024 等）。
- 敏感信息如 API Key 在存入数据库前需加密（如使用 AES-256-GCM），`encrypted_key` 字段存储密文。

---

以上设计覆盖了需求规约中的所有功能点，并符合 PostgreSQL 的数据类型最佳实践。接口设计遵循 RESTful 风格，便于前后端分离开发。
