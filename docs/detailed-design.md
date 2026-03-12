# 基于大模型的私有知识库问答系统详细设计文档

## 1. 文档说明

### 1.1 文档目标

本文档用于替代现有的 AI 生成草稿，输出一份可用于课程设计落地实现的详细设计文档。文档重点覆盖以下内容：

- 系统总体架构与模块划分
- 前后端交互接口设计
- 账户系统与认证授权设计
- 数据库与对象存储设计
- 大模型调用与知识库问答设计
- 异步任务、配额、审计与管理后台设计
- Go 语言实现时的落地建议

### 1.2 设计范围

本设计覆盖首期版本（V1），目标是支持用户基于私有知识库完成问答交互，并提供基础的后台管理能力。

V1 明确包含：

- 用户注册、登录、退出登录、刷新令牌、修改密码
- 单设备登录控制
- 会话管理与 SSE 流式问答
- 私有知识库创建、文档上传、异步解析、向量化、检索
- 标准 RAG 问答
- DeepSeek 作为首期大模型接入方
- 管理员的用户管理、模型配置、配额管理、审计与失败任务重试

V1 明确不包含：

- 第三方登录
- 邮箱验证码激活与找回密码
- 聊天场景下的临时文件上传
- OCR 与扫描版 PDF 识别
- 混合检索、rerank、多知识库联合检索
- 多设备同时在线

### 1.3 已确认设计结论

本轮讨论已经确认以下关键约束，后续设计均以此为前提：

- 文档按语言无关的方式编写，但补充 Go 落地建议
- 系统为“用户私有知识库问答系统”
- 后端形态为“模块化单体 API + 独立 worker”
- 认证采用 `Access Token + Refresh Token`
- 登录策略为单设备登录
- 会话最多绑定一个知识库
- 问答链路采用标准 RAG
- 检索失败时允许回退到通用大模型回答
- 引用在后端按文档块记录，前端首期仅显示文档名
- 文档接入支持 `PDF / TXT / Markdown / DOCX`
- 文档处理、索引重建、资源清理统一走异步任务
- 管理员只能查看用户资源元数据与任务状态，不能查看用户文档正文和聊天正文

### 1.4 文档使用与优先级说明

如果后续开启新的会话，并以本文档为基础继续推进系统设计、数据库设计或编码实现，应遵循以下规则：

- 本文档是当前版本的唯一主设计文档，优先级高于 `docs/design.md`
- `docs/design.md` 仅作为历史草稿参考，不应再作为实现依据
- 如果后续会话中的讨论结论与本文档冲突，应先更新本文档，再开始实现
- V1 范围以本文档为准，不应在未确认的情况下自行加入多知识库联合问答、OCR、多设备登录、聊天临时文件等超范围能力
- 如果新会话需要做具体实现，建议优先阅读第 1 章、第 2 章、第 3 章、第 9 章和第 11 章
- 若进入编码阶段，建议同时阅读 `docs/implementation-kickoff.md`、`docs/migrations/` 和 `docs/openapi-v1-draft.yaml`

### 1.5 后续工作默认原则

为避免新会话在未明确约束的地方做出彼此冲突的假设，补充以下默认原则：

- 未明确指定的实现细节，优先选择满足 V1 目标的最简单方案，不做过早扩展
- 新增接口、表名、状态值时，遵循本文档已有命名方式，避免混用多种命名风格
- 涉及用户正文、文档正文、聊天内容的能力，默认受隐私边界约束，管理员不可见
- 涉及大模型与对象存储的第三方依赖，应通过接口抽象接入，不在业务层直接绑定 SDK
- 后续如果新增关键约束或删改既有约束，应同步回写本文档，而不是仅保留在聊天记录中

## 2. 系统概述与总体架构

### 2.1 业务目标

系统面向两类角色：

- 普通用户：维护自己的知识库，基于知识库进行问答
- 管理员：维护系统配置、配额策略、用户状态、模型接入配置与任务运行状态

系统的核心业务目标如下：

- 让用户能够上传文档构建私有知识库
- 让系统能够自动完成文档解析、分块、向量化与检索
- 让用户在会话中通过 SSE 获得流式回答
- 在回答中保留基础可追踪信息，包括模型、token 用量和引用块
- 提供最小可运维能力，包括失败任务重试、配额管控和审计

### 2.2 架构选型

系统采用模块化单体架构，对外暴露一个 API 服务，并配套一个独立 worker。这样设计的原因如下：

- 相比微服务，模块化单体更适合课程设计和中小规模项目，开发成本更低
- 相比把所有慢任务放到同步请求里，独立 worker 能显著改善文件上传与重建索引的用户体验
- 业务边界仍然清晰，后续若体量扩大，可继续将某些模块拆成独立服务

### 2.3 部署视图

```text
+-------------------+         +------------------------+
| Web Frontend      |  HTTPS  | API Service            |
| Admin Frontend    +-------->+ - Auth / User          |
+-------------------+         | - Chat / Session       |
                              | - KB / Document        |
                              | - Model Gateway        |
                              | - Admin / Quota        |
                              +----+-------------+-----+
                                   |             |
                                   |             |
                             +-----v----+   +----v----------------+
                             | PostgreSQL|   | Object Storage     |
                             | +pgvector |   | MinIO / OSS / S3   |
                             +-----+----+   +---------------------+
                                   |
                             +-----v-------------------------------+
                             | Worker                              |
                             | - Document parse / chunk / embed    |
                             | - Reindex                           |
                             | - Cleanup                           |
                             | - Retry                             |
                             +-------------+-----------------------+
                                           |
                                     +-----v-----+
                                     | DeepSeek  |
                                     | API       |
                                     +-----------+
```

### 2.4 模块划分

#### 2.4.1 账户与认证模块

职责：

- 注册、登录、刷新令牌、退出登录
- 修改密码
- 用户状态校验（active / frozen）
- 单设备登录控制

边界：

- 只负责“谁可以访问系统”，不负责配额与知识库权限
- 令牌签发与会话失效由本模块统一控制

#### 2.4.2 会话与消息模块

职责：

- 创建、删除、查询会话
- 管理会话中的用户消息和助手消息
- 通过 SSE 输出流式回答
- 重新生成回答并覆盖已有 assistant 消息

边界：

- 不负责文档解析和向量化
- 会话最多关联一个知识库，由会话配置决定是否启用 RAG

#### 2.4.3 知识库与文档模块

职责：

- 创建、修改、删除知识库
- 上传文档并创建入库任务
- 维护文档处理状态、索引状态、文档去重
- 提供检索所需的文档元数据

边界：

- 文档解析和向量化由 worker 执行
- 文档内容对管理员不可见

#### 2.4.4 RAG 检索模块

职责：

- 根据问题生成查询向量
- 在指定知识库内检索相关文档块
- 应用相似度阈值、top_k 和元数据过滤
- 组织引用信息并传给问答模块

边界：

- 只在“单知识库范围内”检索
- 检索失败时返回空命中结果，由聊天模块决定是否回退通用问答

#### 2.4.5 模型网关模块

职责：

- 屏蔽不同大模型提供商的调用差异
- 提供统一的聊天生成与 embedding 接口
- 负责超时、重试、错误映射和 token 用量采集

边界：

- V1 只实现 DeepSeek 适配器
- 业务层不直接依赖具体 SDK

#### 2.4.6 任务与调度模块

职责：

- 统一管理异步任务状态
- 调度 worker 执行文档解析、重建索引、清理任务
- 失败任务自动重试 3 次

边界：

- 不承担复杂分布式编排，只提供数据库驱动的轻量任务机制

#### 2.4.7 管理后台模块

职责：

- 用户账户管理
- 模型配置管理
- 系统参数管理
- 配额策略管理
- 审计日志查询
- 失败任务重试

边界：

- 管理员可以查看资源元数据与摘要统计，但不能查看用户正文内容

## 3. 关键业务流程设计

### 3.1 用户注册与登录流程

#### 3.1.1 注册

1. 用户提交用户名、邮箱和密码。
2. 系统校验：
   - 用户名格式是否合法
   - 邮箱格式是否合法
   - 用户名和邮箱是否唯一
   - 密码是否满足复杂度要求
3. 校验通过后，系统创建用户，状态为 `active`。
4. 由于 V1 不做邮箱激活，注册成功后用户可直接登录。

#### 3.1.2 登录

1. 用户以“用户名或邮箱 + 密码”的方式登录。
2. 后端校验密码哈希、账户状态。
3. 若用户已有活跃登录会话，则立即废弃旧 `refresh token`。
4. 生成新的 `access token` 和 `refresh token`。
5. 返回 `access token`，并将 `refresh token` 以安全方式返回客户端。

设计说明：

- `access token` 过期时间为 30 分钟
- `refresh token` 过期时间为 7 天
- 单设备登录只强制废弃旧 `refresh token`，旧 `access token` 自然过期，不维护 access token 黑名单

#### 3.1.3 刷新令牌

1. 客户端在 `access token` 过期后调用刷新接口。
2. 后端校验 `refresh token` 是否存在、是否已撤销、是否过期。
3. 校验通过后，执行 refresh token 轮换：
   - 生成新的 access token
   - 生成新的 refresh token
   - 废弃旧 refresh token
4. 返回新的令牌信息。

#### 3.1.4 退出登录

1. 客户端调用退出接口。
2. 服务端废弃当前登录会话对应的 `refresh token`。
3. 客户端清除本地的 `access token`。

#### 3.1.5 修改密码

1. 用户提交原密码和新密码。
2. 系统校验原密码正确性与新密码复杂度。
3. 更新密码哈希。
4. 废弃当前登录会话，要求用户重新登录。

### 3.2 文档上传与知识库入库流程

1. 用户进入某个知识库并上传文档。
2. API 服务完成以下同步处理：
   - 校验知识库归属权
   - 校验文件类型和大小
   - 计算文件内容哈希
   - 校验同一知识库内是否已经存在相同内容文档
   - 将原始文件上传到对象存储
   - 创建 `files`、`documents`、`tasks` 记录
3. Worker 拉取 `document_ingest` 任务并执行：
   - 解析文档文本
   - 基于标题/段落切块
   - 对超长块进行窗口切分
   - 调用 embedding 模型生成向量
   - 写入 `document_chunks`
   - 更新文档状态为 `available`
4. 若任何步骤失败：
   - 任务进入 `failed` 状态
   - 自动重试，最多 3 次
   - 超出次数后等待管理员手工重试

### 3.3 会话问答流程

1. 用户创建会话，会话可选择一个知识库，也可不绑定知识库。
2. 用户提交问题。
3. 系统检查：
   - 用户是否已登录且未被冻结
   - 用户是否未超出配额
   - 会话是否存在且归属当前用户
4. 如果会话绑定知识库：
   - 调用 RAG 检索模块在该知识库内检索相关块
   - 若命中块数大于 0，则构造“带上下文”的提示词
   - 若未命中，则标记 `grounded=false`，回退到通用大模型回答
5. 调用模型网关，使用 SSE 向前端流式返回文本。
6. 回答完成后：
   - 保存 assistant 消息内容
   - 保存 `model_used`
   - 保存 token 用量
   - 保存引用块记录
   - 更新会话 `updated_at`
   - 更新配额统计

### 3.4 重新生成回答流程

V1 采用“覆盖式重生成”。

1. 用户选择某条 assistant 消息执行重新生成。
2. 系统找到其关联的用户问题消息。
3. 重新执行与初次问答相同的检索和模型生成逻辑。
4. 使用新回答覆盖原 assistant 消息内容与引用信息。
5. 更新 token 统计和 `updated_at`。

设计说明：

- V1 不保留回答版本历史
- 为保证可追踪性，仍然记录最新回答的模型、token 和引用信息

### 3.5 知识库重建索引流程

1. 用户或管理员触发知识库重建索引。
2. 系统创建 `knowledge_base_reindex` 任务。
3. Worker 按知识库遍历文档：
   - 删除旧文档块
   - 重新解析与切块
   - 重新生成 embedding
4. 全部完成后更新知识库索引时间与任务状态。

### 3.6 资源删除与延迟清理流程

删除采用“软删除 + 异步清理”机制。

1. 用户删除会话、知识库或文档时，系统先写入 `deleted_at`。
2. 同时创建 `resource_cleanup` 任务。
3. Worker 在保留期后清理对象存储文件、向量块和孤立引用记录。

推荐保留期：

- 软删除保留 7 天

这样设计的目的：

- 降低误删风险
- 避免同步删除导致长耗时请求
- 便于失败重试和数据恢复

### 3.7 V1 推荐实施顺序

如果后续工作从“开始设计”进入“开始实现”，建议按以下顺序推进：

1. 完成数据库迁移基线、用户模型、认证模块和登录会话管理
2. 完成会话、消息、SSE 流式问答主链路，但先用模型网关 mock 或最小可用 provider 验证接口
3. 完成知识库、文档上传、对象存储接入和统一任务表
4. 完成文档解析、分块、embedding、向量检索和 RAG 问答闭环
5. 完成管理员能力，包括配额、审计、失败任务重试
6. 最后补充系统参数管理、观测、错误码细化和测试

按这个顺序推进的原因是：

- 可以先打通“登录 -> 发问 -> 流式返回”的主链路
- 再逐步把知识库和异步任务接入，降低一次性复杂度
- 管理后台和运维能力放在主流程稳定后再补更稳妥

## 4. 接口设计

### 4.1 接口设计原则

- 前缀统一为 `/api/v1`
- 普通接口使用 JSON
- 流式回答接口使用 `SSE`
- 非流式接口统一返回标准包裹结构
- 请求需携带 `Authorization: Bearer <access_token>`
- 刷新令牌建议通过 `HttpOnly Cookie` 传输；如前后端部署条件受限，可退化为请求体传输，但不是首选

### 4.2 通用响应格式

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

失败响应：

```json
{
  "code": 40001,
  "message": "invalid password",
  "data": null,
  "request_id": "trace-or-request-id",
  "details": []
}
```

错误响应字段约定：

- `code`：系统内部错误码
- `message`：面向前端展示或日志记录的简要说明
- `request_id`：请求链路追踪标识，便于日志排查
- `details`：字段级校验错误列表，可为空数组

### 4.3 认证与用户接口

#### 4.3.1 注册

- `POST /api/v1/auth/register`

请求：

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "StrongPassword123"
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "user_id": "uuid"
  }
}
```

#### 4.3.2 登录

- `POST /api/v1/auth/login`

请求：

```json
{
  "account": "alice",
  "password": "StrongPassword123"
}
```

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "access_token": "jwt",
    "expires_in": 1800,
    "user": {
      "id": "uuid",
      "username": "alice",
      "role": "user"
    }
  }
}
```

说明：

- `refresh token` 通过 `HttpOnly Cookie` 返回更安全
- 登录成功时废弃该用户旧登录会话

#### 4.3.3 刷新令牌

- `POST /api/v1/auth/refresh`

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "access_token": "new-jwt",
    "expires_in": 1800
  }
}
```

#### 4.3.4 退出登录

- `POST /api/v1/auth/logout`

作用：

- 废弃当前 refresh token 对应的登录会话

#### 4.3.5 当前用户信息

- `GET /api/v1/users/me`
- `PUT /api/v1/users/me`
- `PUT /api/v1/users/me/password`

### 4.4 会话与问答接口

#### 4.4.1 创建会话

- `POST /api/v1/sessions`

请求：

```json
{
  "name": "软件工程课程设计",
  "model": "deepseek-chat",
  "knowledge_base_id": "uuid-or-null"
}
```

设计约束：

- V1 中会话创建后不提供“切换知识库”能力
- 若需要切换知识库或模型，建议新建会话

#### 4.4.2 查询会话列表

- `GET /api/v1/sessions?page=1&size=20&keyword=xxx`

#### 4.4.3 查询会话详情

- `GET /api/v1/sessions/{sessionId}`

返回：

- 会话基础信息
- 消息列表
- 当前知识库摘要信息

#### 4.4.4 删除会话

- `DELETE /api/v1/sessions/{sessionId}`

#### 4.4.5 发送消息

- `POST /api/v1/sessions/{sessionId}/messages`
- 响应类型：`text/event-stream`

请求：

```json
{
  "content": "请根据知识库总结系统架构设计"
}
```

SSE 事件建议如下：

- `meta`：本次回答基础信息，如 `message_id`、`grounded`、`model`
- `delta`：增量文本
- `done`：结束事件
- `error`：错误事件

示例：

```text
event: meta
data: {"message_id":"uuid","grounded":true,"model":"deepseek-chat"}

event: delta
data: {"content":"系统采用"}

event: delta
data: {"content":"模块化单体架构"}

event: done
data: {"finish_reason":"stop"}
```

#### 4.4.6 重新生成回答

- `POST /api/v1/sessions/{sessionId}/messages/{messageId}/regenerate`

作用：

- 使用原用户问题重新生成回答
- 覆盖原 assistant 消息内容与引用记录

#### 4.4.7 中断生成

- `POST /api/v1/sessions/{sessionId}/stream/stop`

说明：

- 仅中断当前会话中正在执行的流式任务
- 已生成的部分内容可按产品策略选择丢弃或保留
- V1 建议：若在模型侧成功中断，则不提交不完整答案为正式 assistant 消息

### 4.5 知识库接口

#### 4.5.1 知识库管理

- `GET /api/v1/knowledge-bases`
- `POST /api/v1/knowledge-bases`
- `GET /api/v1/knowledge-bases/{kbId}`
- `PUT /api/v1/knowledge-bases/{kbId}`
- `DELETE /api/v1/knowledge-bases/{kbId}`

创建知识库请求建议：

```json
{
  "name": "课程设计资料库",
  "description": "存放项目说明书、需求文档和设计文档",
  "embedding_model": "text-embedding-3-small"
}
```

说明：

- `embedding_model` 在知识库创建时固化
- 若后续修改 embedding 模型，必须通过“替换并重建索引”的方式完成

#### 4.5.2 文档管理

- `GET /api/v1/knowledge-bases/{kbId}/documents`
- `POST /api/v1/knowledge-bases/{kbId}/documents`
- `GET /api/v1/knowledge-bases/{kbId}/documents/{docId}`
- `DELETE /api/v1/knowledge-bases/{kbId}/documents/{docId}`

上传返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "document_id": "uuid",
    "task_id": "uuid",
    "status": "pending"
  }
}
```

#### 4.5.3 重建索引

- `POST /api/v1/knowledge-bases/{kbId}/reindex`

响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "task_id": "uuid"
  }
}
```

### 4.6 管理后台接口

#### 4.6.1 用户与账户管理

- `GET /api/v1/admin/users`
- `POST /api/v1/admin/users`
- `PUT /api/v1/admin/users/{userId}`
- `PUT /api/v1/admin/users/{userId}/password`
- `POST /api/v1/admin/users/{userId}/freeze`
- `POST /api/v1/admin/users/{userId}/unfreeze`

#### 4.6.2 模型配置与系统参数

- `GET /api/v1/admin/provider-configs`
- `PUT /api/v1/admin/provider-configs/{provider}`
- `GET /api/v1/admin/settings`
- `PUT /api/v1/admin/settings`

#### 4.6.3 任务、审计与配额

- `GET /api/v1/admin/tasks`
- `POST /api/v1/admin/tasks/{taskId}/retry`
- `GET /api/v1/admin/audit-logs`
- `GET /api/v1/admin/quota-policies`
- `PUT /api/v1/admin/quota-policies/{policyId}`
- `GET /api/v1/admin/usage/users/{userId}`

## 5. 账户系统设计

### 5.1 用户模型

系统定义两类角色：

- `user`：普通用户
- `admin`：管理员

系统定义两类账户状态：

- `active`：可正常登录和使用
- `frozen`：禁止登录、禁止刷新令牌、禁止访问业务接口

### 5.2 密码策略

密码存储采用 `Argon2id` 哈希，不可逆加密，不保存明文密码。

推荐密码规则：

- 长度不少于 8 位
- 同时包含字母和数字
- 禁止与用户名相同

### 5.3 令牌设计

#### 5.3.1 Access Token

- 用途：接口鉴权
- 有效期：30 分钟
- 携带方式：`Authorization` 头
- 内容：`user_id`、`role`、`session_id`、`exp`

#### 5.3.2 Refresh Token

- 用途：刷新 access token
- 有效期：7 天
- 建议存储：`HttpOnly + Secure` Cookie
- 数据库存储：只保存 refresh token 哈希，不保存明文

### 5.4 单设备登录策略

系统采用单设备登录。

实现逻辑：

1. 用户登录成功后，服务端将该用户旧登录会话标记为已撤销。
2. 新建新的登录会话并签发新的 refresh token。
3. 旧 access token 不做黑名单处理，最多在 30 分钟内自然失效。

这样设计的原因：

- 避免 access token 黑名单带来的复杂度
- 保证“踢旧设备”语义在下次刷新时生效
- 对课程设计规模来说实现成本和安全性较平衡

### 5.5 修改密码与账户冻结

以下场景需要强制使当前登录状态失效：

- 用户主动修改密码
- 管理员重置密码
- 管理员冻结账户

失效策略：

- 立即撤销 refresh token
- 已签发 access token 自然过期

### 5.6 Go 落地建议

对有 Spring Boot 背景的开发者，可按下面的方式理解：

- `handler` 类似 Controller，负责参数解析和响应输出
- `service` 负责注册、登录、刷新、修改密码等业务逻辑
- `repository` 负责读写 `users`、`user_sessions`
- `middleware` 负责 JWT 鉴权与角色校验

## 6. 知识库与 RAG 设计

### 6.1 知识库模型

每个知识库属于一个用户，具备以下关键属性：

- 名称、描述
- embedding 模型标识
- 提示词模板
- 检索参数，如 `top_k`、相似度阈值
- 索引状态与最近重建时间

设计约束：

- 一个知识库只能使用一个 embedding 模型
- 该模型在知识库创建时固化
- 若要切换 embedding 模型，必须重建整个知识库索引

### 6.2 文档接入范围

V1 支持的文档类型：

- PDF
- TXT
- Markdown
- DOCX

V1 不支持：

- 扫描版 PDF 的 OCR
- 图片提取文字
- 聊天临时文件

### 6.3 文档去重策略

系统使用文件内容哈希做去重。

规则：

- 在同一知识库内，若上传文件的 `sha256` 与已有文档一致，则判定为重复
- 对重复文档直接拒绝重复索引
- 若用户希望更新文档，应走“替换并重建”流程

这样设计的原因：

- 比文件名判重更可靠
- 可避免重复占用对象存储和向量空间
- 有利于后续做幂等处理

### 6.4 文档分块策略

V1 采用“两阶段分块”：

1. 先按文档结构切分：
   - 标题
   - 段落
   - 自然换行
2. 对过长文本块再按窗口切分：
   - 固定目标长度
   - 保留少量重叠

设计原则：

- 优先保留语义边界，而不是机械定长切块
- 对超长段落进行二次切块，避免单块过大导致检索效果下降

块元数据建议包括：

- `document_id`
- `knowledge_base_id`
- `chunk_index`
- `heading_path`
- `token_count`
- `source_page` 或 `source_section`

### 6.5 检索流程

问答阶段检索链路如下：

1. 根据用户问题生成查询向量
2. 在会话绑定的知识库范围内检索相似文档块
3. 应用相似度阈值与 `top_k` 限制
4. 按相关性顺序取前若干块作为上下文
5. 将命中文档块记录为引用

V1 约束：

- 仅做单知识库向量检索
- 不做关键词检索
- 不做 rerank
- 不做多知识库合并召回

### 6.6 检索失败的回退策略

若知识库未检索到有效内容，系统允许回退到通用大模型回答。

但必须同时满足：

- 返回字段中标记 `grounded=false`
- 前端明确展示“未命中知识库，以下为通用回答”
- 后端仍记录此次回答使用了哪个模型，但不生成知识库引用

这样设计的目的：

- 保证系统可回答开放问题
- 避免用户将无依据回答误认为来源于私有知识库

### 6.7 引用记录设计

后端按“文档块”粒度存储引用，前端首期仅显示文档名。

这样设计的好处：

- 首期前端实现简单
- 后续若想展示命中片段或定位到具体段落，无需改数据库

### 6.8 提示词模板分层

V1 采用两级提示词：

- 全局系统提示词
- 知识库级提示词模板

组合方式：

1. 先拼接系统级规则，如“回答必须准确，不得捏造引用”
2. 再拼接知识库业务提示词，如“优先总结课程设计相关章节”
3. 最后拼接用户问题和检索结果

### 6.9 关于 pgvector 的实现限制

需要特别说明一个容易被忽略的问题：

- `pgvector` 的 `VECTOR(n)` 列要求向量维度固定
- 如果同一张表混用不同维度的 embedding 模型，会带来实现复杂度

因此 V1 的实际约束建议为：

- 同一部署实例内仅启用“维度一致”的 embedding 模型集合
- `knowledge_bases.embedding_model` 记录逻辑模型名
- 如果未来需要真正支持不同维度模型并存，可考虑分表、分库或独立向量引擎

这也是为什么文档中虽然允许“知识库创建时固化 embedding 模型”，但工程实现上仍建议优先控制模型维度一致。

### 6.10 默认运行参数建议

为了让后续新会话能够直接开始详细设计或编码，补充一组可作为首期默认值的运行参数。若后续未单独确认，可按以下值启动实现：

- 单文件上传大小上限：`20 MiB`
- 支持的 MIME 类型：`application/pdf`、`text/plain`、`text/markdown`、`application/vnd.openxmlformats-officedocument.wordprocessingml.document`
- 分块目标大小：约 `500 tokens`
- 分块重叠：约 `80 tokens`
- 默认检索 `top_k`：`5`
- 参与提示词拼装的最大引用块数：`5`
- 会话中拼接到模型上下文的历史消息范围：最近 `5` 轮问答
- SSE 心跳间隔：`15 秒`

说明：

- 这些参数属于“实现默认值”，不等于最终生产参数
- 正式落地时应将它们做成系统参数，而不是写死在代码中
- 相似度阈值与具体 embedding 模型强相关，首期建议做成可配置项，并通过测试数据调整

## 7. 大模型调用设计

### 7.1 设计目标

模型调用层需要解决三个问题：

- 业务层不直接绑定具体模型提供商
- 流式回答与 embedding 生成采用统一接口
- 对调用失败、超时、重试和 token 用量进行统一处理

### 7.2 Provider 抽象

建议在业务层定义统一接口：

```go
type ChatProvider interface {
    StreamChat(ctx context.Context, req ChatRequest) (StreamResult, error)
}

type EmbeddingProvider interface {
    Embed(ctx context.Context, req EmbeddingRequest) (EmbeddingResult, error)
}
```

V1 实现：

- `DeepSeekChatProvider`
- `DeepSeekEmbeddingProvider` 或兼容的 embedding 提供方适配器

说明：

- 即使首期只接一个 provider，也建议做这一层抽象
- 后续若切换到 OpenAI、本地模型或代理网关，不需要修改聊天和知识库业务逻辑

### 7.3 聊天请求组装

聊天请求由以下部分组成：

- 全局系统提示词
- 知识库级提示词模板
- 检索命中的上下文片段
- 历史消息摘要或最近若干轮消息
- 当前用户问题

V1 建议：

- 会话历史仅取最近若干轮，避免上下文无限增长
- 在提示词中明确要求模型区分“知识库命中回答”和“通用回答”

### 7.4 超时、重试与错误处理

建议策略：

- 聊天生成超时：例如 60 秒
- embedding 生成超时：例如 20 秒
- 网络抖动或短暂 5xx 错误可有限次重试
- 参数错误、鉴权错误不重试

错误映射：

- 上游网络错误 -> `provider_unavailable`
- 密钥无效 -> `provider_auth_failed`
- 限流 -> `provider_rate_limited`
- 上下文过长 -> `prompt_too_large`

### 7.5 Token 用量采集

每次模型调用后，需要记录：

- `prompt_tokens`
- `completion_tokens`
- `total_tokens`
- 调用耗时
- 模型标识
- 是否命中知识库

这些数据用于：

- 用户配额统计
- 审计和运维观察
- 成本分析

### 7.6 配置与密钥管理

管理员维护模型提供商配置，而不是简单只存一个 API Key。

建议维护内容：

- provider 名称
- base URL
- 默认 chat model
- 默认 embedding model
- 加密后的 API Key
- 是否启用

敏感数据要求：

- API Key 入库前使用对称加密
- 解密只发生在调用模型前的最小范围内
- 管理端只展示脱敏值

## 8. 异步任务与状态机设计

### 8.1 任务分类

V1 使用统一 `tasks` 表表示异步任务，任务类型包括：

- `document_ingest`
- `knowledge_base_reindex`
- `resource_cleanup`

### 8.2 任务状态

建议状态机：

- `pending`
- `running`
- `succeeded`
- `failed`
- `cancelled`

重试相关字段：

- `attempt_count`
- `max_attempts`
- `next_run_at`

### 8.3 Worker 执行模型

Worker 周期性拉取可执行任务，采用如下约束：

- 使用数据库行锁或状态更新避免重复消费
- 同一文档不允许并发执行多个入库任务
- 同一知识库不允许并发执行多个重建任务

### 8.4 文档状态机

除通用任务状态外，文档本身也需维护业务状态：

- `pending`
- `processing`
- `available`
- `failed`
- `deleting`

关系如下：

- 创建文档时默认 `pending`
- ingest 任务开始时改为 `processing`
- 完成后改为 `available`
- 连续失败后改为 `failed`
- 删除流程中改为 `deleting`

### 8.5 失败重试策略

失败重试规则：

- 自动重试最多 3 次
- 每次重试间隔可逐步增加
- 超过 3 次后由管理员手工触发重试

不建议无限自动重试，因为：

- 可能掩盖脏数据或坏文件
- 会持续占用 worker 资源
- 容易造成重复外部调用成本

### 8.6 幂等性与并发约束

如果后续开启新会话直接进入实现阶段，这一节需要作为强制约束使用：

- `document_ingest` 任务必须以 `document_id` 为幂等键，重复执行时不能产生重复块数据
- 文档重新入库前，应先清理该文档旧的 `document_chunks`，再写入新的切块结果
- 同一知识库同一时刻最多允许一个 `knowledge_base_reindex` 任务处于运行态
- `resource_cleanup` 任务必须允许重复执行，即使目标对象已不存在也应安全完成
- Worker 抢任务时必须通过事务和状态变更防止重复消费

这样设计的原因是：

- 异步任务失败重试是常态，没有幂等性就无法安全重跑
- 文档重建和资源清理都天然存在重复执行可能
- 这些约束如果不提前写进文档，后续实现极易在边界行为上出现不一致

## 9. 数据库设计

### 9.1 数据库选型

数据库采用 `PostgreSQL 15+`，启用如下扩展：

- `pgcrypto`：生成 UUID、辅助加密
- `pgvector`：存储向量

所有业务表统一遵循：

- 主键优先使用 UUID
- 时间字段使用 `TIMESTAMPTZ`
- 支持软删除的实体保留 `deleted_at`

### 9.1.1 字符串字段类型约定

考虑到本项目数据库选型为 PostgreSQL，字符串字段采用如下约定：

- 默认使用 `TEXT`，不将 `VARCHAR(n)` 作为常规选择
- 长度限制属于业务约束，不通过 `VARCHAR(n)` 表达，而通过应用层校验和数据库 `CHECK (char_length(...) <= n)` 约束表达
- 状态类和枚举类字段优先使用 `TEXT + CHECK`，例如 `role`、`status`、`task_type`
- 需要大小写不敏感唯一性的字段，如 `username`、`email`，建议通过 `lower(column)` 唯一索引实现

采用这一约定的原因是：

- 在 PostgreSQL 中，`TEXT` 与 `VARCHAR` 在存储与性能上通常没有实质差异
- `TEXT + CHECK` 更能清晰区分“存储类型”和“业务规则”
- 后续如果字段长度上限调整，不需要修改列类型定义

### 9.2 主要表设计

#### 9.2.1 users

用途：保存用户基础信息。

关键字段：

- `id`
- `username`
- `email`
- `password_hash`
- `nickname`
- `avatar_url`
- `role`
- `status`
- `created_at`
- `updated_at`
- `deleted_at`

约束：

- `username`、`email` 使用 `TEXT`
- `username`、`email` 使用大小写不敏感唯一索引
- `role` 使用 `TEXT CHECK (role IN ('user', 'admin'))`
- `status` 使用 `TEXT CHECK (status IN ('active', 'frozen'))`

#### 9.2.2 user_sessions

用途：保存当前登录会话与 refresh token。

关键字段：

- `id`
- `user_id`
- `refresh_token_hash`
- `device_label`
- `user_agent`
- `ip_address`
- `expires_at`
- `last_active_at`
- `revoked_at`
- `created_at`
- `updated_at`

设计说明：

- V1 是单设备登录，但仍保留独立会话表，便于审计和后续扩展
- 通过 `revoked_at` 表示会话失效
- `device_label`、`user_agent`、`ip_address` 均使用 `TEXT`

#### 9.2.3 sessions

用途：保存聊天会话。

关键字段：

- `id`
- `user_id`
- `name`
- `model`
- `knowledge_base_id`
- `created_at`
- `updated_at`
- `deleted_at`

说明：

- `knowledge_base_id` 可为空，表示通用问答会话
- 不再使用 `UUID[]`
- `name`、`model` 使用 `TEXT`

#### 9.2.4 messages

用途：保存会话中的消息。

关键字段：

- `id`
- `session_id`
- `role`
- `reply_to_message_id`
- `content`
- `status`
- `model_used`
- `grounded`
- `prompt_tokens`
- `completion_tokens`
- `total_tokens`
- `created_at`
- `updated_at`

设计说明：

- 用户消息 `reply_to_message_id` 为空
- assistant 消息通过 `reply_to_message_id` 指向对应的用户问题
- 重新生成回答时覆盖该 assistant 消息内容
- `role`、`status`、`model_used` 均使用 `TEXT`，其中 `role`、`status` 建议配套 `CHECK` 约束

#### 9.2.5 message_citations

用途：保存 assistant 消息所引用的知识块。

关键字段：

- `id`
- `message_id`
- `document_chunk_id`
- `document_id`
- `knowledge_base_id`
- `rank_no`
- `created_at`

说明：

- 首期前端仅展示文档名
- 但数据库按块级记录，便于未来扩展

#### 9.2.6 knowledge_bases

用途：保存知识库信息。

关键字段：

- `id`
- `user_id`
- `name`
- `description`
- `embedding_model`
- `prompt_template`
- `retrieval_top_k`
- `similarity_threshold`
- `last_indexed_at`
- `created_at`
- `updated_at`
- `deleted_at`

设计说明：

- `name`、`description`、`embedding_model`、`prompt_template` 均使用 `TEXT`

#### 9.2.7 files

用途：保存原始文件的对象存储信息。

关键字段：

- `id`
- `user_id`
- `storage_provider`
- `bucket_name`
- `object_key`
- `original_filename`
- `mime_type`
- `size_bytes`
- `sha256`
- `created_at`
- `deleted_at`

说明：

- `sha256` 用于文件内容识别和去重
- 不建议直接把文件二进制放数据库
- `storage_provider`、`bucket_name`、`object_key`、`original_filename`、`mime_type` 均使用 `TEXT`

#### 9.2.8 documents

用途：保存知识库内的逻辑文档。

关键字段：

- `id`
- `knowledge_base_id`
- `file_id`
- `title`
- `status`
- `error_message`
- `content_text`
- `content_length`
- `chunk_count`
- `created_at`
- `updated_at`
- `deleted_at`

设计说明：

- `content_text` 存储解析后的纯文本
- 管理员无权查看 `content_text`
- 对于大文件，若担心数据库膨胀，可后续将正文转移到对象存储并仅保留摘要
- `title`、`status`、`error_message` 使用 `TEXT`，其中 `status` 建议配套 `CHECK` 约束

#### 9.2.9 document_chunks

用途：保存切块文本及向量。

关键字段：

- `id`
- `knowledge_base_id`
- `document_id`
- `chunk_index`
- `heading_path`
- `content`
- `token_count`
- `source_page`
- `embedding`
- `created_at`

索引建议：

- `(document_id, chunk_index)` 唯一索引
- `embedding` 上建立 HNSW 或 ivfflat 向量索引
- `heading_path`、`content` 使用 `TEXT`

#### 9.2.10 tasks

用途：统一保存异步任务。

关键字段：

- `id`
- `task_type`
- `resource_type`
- `resource_id`
- `user_id`
- `status`
- `payload`
- `result`
- `attempt_count`
- `max_attempts`
- `next_run_at`
- `started_at`
- `finished_at`
- `error_code`
- `error_message`
- `created_at`
- `updated_at`

设计说明：

- `task_type`、`resource_type`、`status`、`error_code`、`error_message` 均使用 `TEXT`
- `task_type`、`status` 建议配套 `CHECK` 约束或统一常量定义

#### 9.2.11 provider_configs

用途：保存模型提供商配置。

关键字段：

- `id`
- `provider`
- `base_url`
- `default_chat_model`
- `default_embedding_model`
- `encrypted_api_key`
- `is_enabled`
- `created_at`
- `updated_at`

设计说明：

- `provider`、`base_url`、`default_chat_model`、`default_embedding_model` 使用 `TEXT`

#### 9.2.12 system_settings

用途：保存系统级可配置参数。

关键字段：

- `key`
- `category`
- `value`
- `description`
- `updated_at`

设计说明：

- `key`、`category`、`description` 使用 `TEXT`
- `value` 使用 `JSONB`
- 用于保存默认模型、SSE 心跳、上传大小限制、默认检索参数等系统级配置

#### 9.2.13 quota_policies

用途：保存配额策略。

关键字段：

- `id`
- `scope_type`
- `scope_id`
- `daily_total_tokens_limit`
- `storage_bytes_limit`
- `document_count_limit`
- `warn_ratio`
- `created_at`
- `updated_at`

说明：

- `scope_type` 可为 `system_default` 或 `user`
- 用户级策略覆盖系统默认策略
- `scope_type` 使用 `TEXT CHECK (scope_type IN ('system_default', 'user'))`

#### 9.2.14 daily_usage_counters

用途：保存用户每日用量。

关键字段：

- `id`
- `user_id`
- `usage_date`
- `request_count`
- `prompt_tokens`
- `completion_tokens`
- `total_tokens`
- `created_at`
- `updated_at`

#### 9.2.15 user_resource_usage

用途：保存用户当前资源占用。

关键字段：

- `user_id`
- `knowledge_base_count`
- `document_count`
- `storage_bytes`
- `updated_at`

#### 9.2.16 audit_logs

用途：保存审计记录。

关键字段：

- `id`
- `actor_user_id`
- `actor_role`
- `action`
- `resource_type`
- `resource_id`
- `target_user_id`
- `result`
- `metadata`
- `created_at`

设计约束：

- 允许记录登录事件、冻结用户、修改系统设置、任务重试、模型调用摘要和用量
- 不记录用户原始正文内容
- `actor_role`、`action`、`resource_type`、`result` 使用 `TEXT`

### 9.3 关系说明

- `users` 与 `user_sessions`：一对多，V1 通过逻辑约束只保留一个有效会话
- `users` 与 `sessions`：一对多
- `sessions` 与 `messages`：一对多
- `messages` 与 `message_citations`：一对多
- `users` 与 `knowledge_bases`：一对多
- `knowledge_bases` 与 `documents`：一对多
- `documents` 与 `document_chunks`：一对多
- `users` 与 `files`：一对多
- `users` 与 `tasks`：一对多

### 9.4 与原 AI 草稿相比的关键修正

- 将 `sessions.knowledge_base_ids` 改为 `knowledge_base_id`
- 去掉 `messages.files JSONB`
- 新增 `user_sessions` 管理 refresh token
- 新增 `message_citations` 表
- 新增 `tasks`、`quota_policies`、`daily_usage_counters`、`user_resource_usage`
- 将简单 `api_keys` 提升为 `provider_configs`
- 明确 `admin/settings` 对应独立的 `system_settings` 表
- 明确 PostgreSQL 中字符串字段默认使用 `TEXT`，长度与枚举约束通过 `CHECK` 和索引表达

## 10. 配额、审计与管理后台设计

### 10.1 配额设计

V1 配额控制三类资源：

- 每日 token 总量
- 当前存储容量
- 当前文档数量

超限策略：

- 达到预警阈值时提示用户与管理员
- 真正超限时拒绝继续调用或上传

### 10.2 审计设计

需要审计的关键事件：

- 登录成功 / 失败
- 退出登录
- 修改密码
- 冻结 / 解冻用户
- 修改模型配置
- 修改系统参数
- 触发任务重试
- 模型调用摘要与用量

审计边界：

- 可以记录调用时间、模型、token 用量、成功失败状态
- 不可记录用户问题正文、用户回答正文、文档正文

### 10.3 管理员界面能力边界

管理员可以：

- 查看用户列表与状态
- 查看知识库、文档、任务、配额的元数据
- 重试失败任务
- 调整系统参数和配额策略

管理员不可以：

- 查看用户文档正文
- 查看用户提问和回答正文
- 冒充用户直接操作聊天内容

## 11. Go 落地建议

### 11.1 建议的目录结构

```text
cmd/
  api/
  worker/
internal/
  account/
  admin/
  chat/
  kb/
  rag/
  model/
  task/
  quota/
  audit/
  platform/
    config/
    db/
    storage/
    auth/
    httpx/
```

### 11.2 对 Spring Boot 开发者的理解映射

- `cmd/api/main.go` 类似 Spring Boot 启动类
- `internal/*/handler` 类似 Controller
- `internal/*/service` 类似 Service
- `internal/*/repository` 类似 Repository
- `internal/platform/auth` 类似认证配置和拦截器
- `cmd/worker/main.go` 类似单独的异步任务消费进程

### 11.3 分层建议

建议遵循以下调用方向：

```text
handler -> service -> repository / provider
```

约束：

- handler 不直接写 SQL
- service 不直接依赖 HTTP 框架细节
- provider 不反向依赖业务模块

### 11.4 工程实践建议

- 统一错误码
- 统一请求日志和 trace id
- 使用数据库事务保护关键写操作
- 对 SSE 连接和模型调用设置超时
- 对对象存储和大模型调用做接口抽象，便于测试

### 11.5 编码前仍需固定的工程选型

本文档已经固定了业务和架构边界，但以下工程选型尚未强制写死。若后续新会话要直接开始编码，建议先确认这些内容，再进入具体实现：

- Go 版本，例如 `Go 1.24+`
- HTTP 路由框架，例如 `chi` 或 `gin`
- 数据库访问方式，例如 `pgx + sqlc` 或 `GORM`
- 数据库迁移工具，例如 `goose` 或 `golang-migrate`
- 文档解析库选型，尤其是 `PDF` 与 `DOCX` 的解析方案
- 本地开发使用的对象存储方案，例如 `MinIO`

如果后续会话未单独确认，我建议默认采用下面这组实现组合：

- 路由层：`chi`
- 数据访问：`pgx + sqlc`
- 迁移工具：`goose`
- 日志：标准库 `slog`
- 对象存储：S3 兼容接口，开发环境使用 `MinIO`

推荐理由：

- 这套组合更接近 Go 社区常见的“显式、轻量、少魔法”风格
- 对有 Spring Boot 背景但刚接触 Go 的开发者来说，分层关系清晰，调试成本较低

## 12. 非功能设计

### 12.1 安全性

- 密码使用 `Argon2id`
- refresh token 仅保存哈希
- API Key 入库加密
- 上传文件做类型与大小校验
- 管理接口做角色校验

### 12.2 可维护性

- 使用模块化单体降低复杂度
- 模型调用、对象存储、任务调度均做接口抽象
- 审计、配额、任务状态独立建模

### 12.3 可扩展性

后续可扩展方向：

- 多设备登录
- 混合检索与 rerank
- OCR
- 多提供商模型切换
- 多知识库联合问答
- 更细粒度的答案反馈与效果评估

## 13. DDL 草案说明

为了让后续新会话能够直接进入数据库建模或迁移文件编写，本项目补充了一份 PostgreSQL DDL 草案文件：

- `docs/schema-ddl-draft.sql`
- `docs/migrations/` 目录下的版本化 Goose migration

该草案遵循本文档的设计约束，包含：

- 扩展启用、时间更新触发器
- 核心表结构、主外键、唯一索引、部分索引
- 常见 `CHECK` 约束
- `pgvector` 向量列与示例索引

使用时需要注意：

- DDL 中 `VECTOR(1536)` 仅作为首期草案默认值，正式实现前应与实际选定的 embedding 模型维度核对
- 某些业务约束在课程设计阶段可以放在应用层和数据库层共同维护，不建议只依赖单一层
- 若后续决定采用 `sqlc` 等工具，建议以该草案为基础继续拆分为版本化 migration
- 当前仓库已提供首版 `docs/migrations/`，后续新增表或约束时应继续追加版本化 migration，而不是直接改历史文件

## 14. API 错误码表

为保证接口联调和日志排查的一致性，V1 统一使用整数错误码。推荐区间划分如下：

- `0`：成功
- `400xx`：请求参数或业务校验错误
- `401xx`：认证失败或令牌失效
- `403xx`：权限不足
- `404xx`：资源不存在
- `409xx`：资源冲突
- `422xx`：文件、模型或业务状态不满足要求
- `429xx`：配额或限流
- `500xx`：系统内部错误
- `502xx`：第三方 provider 错误
- `503xx`：服务不可用

| 错误码 | 名称 | 含义 | 常见接口 |
|--------|------|------|----------|
| `0` | `ok` | 请求成功 | 全部 |
| `40000` | `invalid_request` | 请求体或查询参数整体不合法 | 全部 |
| `40001` | `validation_failed` | 字段级校验失败 | 注册、创建知识库、发送消息 |
| `40002` | `bad_pagination` | 分页参数非法 | 列表接口 |
| `40003` | `unsupported_sort` | 排序字段不支持 | 管理列表接口 |
| `40101` | `auth_required` | 未提供 access token | 登录态接口 |
| `40102` | `access_token_invalid` | access token 非法 | 登录态接口 |
| `40103` | `access_token_expired` | access token 已过期 | 登录态接口 |
| `40104` | `refresh_token_invalid` | refresh token 非法 | `/auth/refresh` |
| `40105` | `refresh_token_expired` | refresh token 已过期 | `/auth/refresh` |
| `40106` | `session_revoked` | 登录会话已失效 | `/auth/refresh`、退出后访问 |
| `40107` | `account_frozen` | 账户被冻结 | 登录、刷新、业务接口 |
| `40301` | `forbidden` | 当前用户无权访问该资源 | 跨用户资源访问 |
| `40401` | `user_not_found` | 用户不存在 | 管理接口 |
| `40402` | `session_not_found` | 会话不存在 | 会话与消息接口 |
| `40403` | `knowledge_base_not_found` | 知识库不存在 | 知识库接口 |
| `40404` | `document_not_found` | 文档不存在 | 文档接口 |
| `40405` | `message_not_found` | 消息不存在 | 重生成接口 |
| `40406` | `task_not_found` | 任务不存在 | 任务重试接口 |
| `40901` | `username_exists` | 用户名已存在 | 注册、管理员创建用户 |
| `40902` | `email_exists` | 邮箱已存在 | 注册、管理员创建用户 |
| `40903` | `duplicate_document` | 知识库中已存在相同内容文档 | 文档上传 |
| `40904` | `reindex_in_progress` | 当前知识库已有重建任务执行中 | 重建索引 |
| `42201` | `unsupported_file_type` | 文件类型不支持 | 文档上传 |
| `42202` | `file_too_large` | 文件大小超出限制 | 文档上传 |
| `42203` | `invalid_password` | 原密码错误或密码规则不满足 | 修改密码 |
| `42204` | `invalid_credentials` | 用户名/邮箱与密码不匹配 | 登录 |
| `42205` | `model_not_available` | 指定模型未启用或不存在 | 创建会话、发送消息 |
| `42206` | `prompt_too_large` | 提示词上下文超限 | 发送消息 |
| `42207` | `embedding_dimension_mismatch` | embedding 模型维度与向量列不匹配 | 文档入库、重建索引 |
| `42901` | `daily_token_quota_exceeded` | 当日 token 配额已耗尽 | 发送消息 |
| `42902` | `storage_quota_exceeded` | 存储容量超限 | 文档上传 |
| `42903` | `document_count_quota_exceeded` | 文档数量超限 | 文档上传 |
| `42904` | `provider_rate_limited` | 上游模型服务限流 | 问答、向量化 |
| `50001` | `internal_error` | 未分类内部错误 | 全部 |
| `50002` | `database_error` | 数据库操作失败 | 全部 |
| `50003` | `storage_error` | 对象存储操作失败 | 文档上传、清理任务 |
| `50004` | `task_dispatch_failed` | 异步任务创建或调度失败 | 上传、重建、删除 |
| `50201` | `provider_unavailable` | 上游模型服务不可用 | 问答、向量化 |
| `50202` | `provider_auth_failed` | 上游 API Key 无效 | 问答、向量化 |
| `50203` | `provider_timeout` | 上游调用超时 | 问答、向量化 |
| `50301` | `service_unavailable` | 系统维护中或核心依赖不可用 | 全部 |

使用原则：

- 对外返回 `code` 与 `message`，内部日志同时记录 `request_id`
- 字段级错误优先使用 `40001 validation_failed`，并在 `details` 中列出字段与原因
- 上游模型调用失败应尽量映射为统一错误码，而不是直接透传第三方原始错误

## 15. 字段级 OpenAPI 草案说明

为了让后续新会话能够直接进入接口定义和前后端联调，本项目补充了一份字段级 OpenAPI 草案文件：

- `docs/openapi-v1-draft.yaml`
- `docs/implementation-kickoff.md`

该草案覆盖：

- 认证与用户接口
- 会话与消息接口
- 知识库与文档接口
- 管理后台中的任务、配额与系统参数接口
- 通用响应结构、分页结构、错误响应结构

核心 schema 包括：

- `RegisterRequest`
- `LoginRequest`
- `RefreshTokenResponse`
- `UserProfile`
- `CreateSessionRequest`
- `MessageSendRequest`
- `KnowledgeBaseCreateRequest`
- `DocumentUploadResponse`
- `TaskSummary`
- `ErrorResponse`

使用说明：

- OpenAPI 草案以 V1 的主接口为准，字段命名与本文档保持一致
- 流式问答接口在 OpenAPI 中以 `text/event-stream` 表示，并使用示例描述事件格式
- 如果后续需要自动生成前端类型或服务端接口桩，建议先基于该草案做一次字段审校，再补全剩余边缘接口

## 16. 结论

本设计将原有 AI 草稿改造成了可落地的详细设计方案，核心特点如下：

- 在架构上采用模块化单体 + worker，兼顾实现成本和清晰边界
- 在账户系统上引入 refresh token、单设备登录和真正可失效的退出机制
- 在知识库设计上补齐了分块、去重、检索、回退与引用记录
- 在大模型接入上通过 provider 抽象隔离具体厂商
- 在数据库设计上修正了数组字段和 JSONB 乱入的问题，补齐任务、配额、审计等关键表
- 在实现准备上补充了 DDL 草案、错误码体系和字段级 OpenAPI 草案

这份文档已经可以作为后续原型实现、数据库建模、迁移编写和接口联调的基础版本。若继续推进，下一步最合适的是：

- 将 DDL 草案拆分为实际 migration 文件
- 基于 OpenAPI 草案补全剩余边缘接口
- 增加状态流转图、时序图和测试用例矩阵
