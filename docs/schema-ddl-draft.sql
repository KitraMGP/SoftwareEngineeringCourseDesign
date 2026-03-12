-- Draft PostgreSQL DDL for the V1 knowledge QA system.
-- Assumptions:
-- 1. PostgreSQL 15+
-- 2. Embedding dimension is fixed to 1536 for the first draft.
-- 3. Business logic still validates permissions and state transitions at application level.

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS vector;

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT NOT NULL CHECK (char_length(username) BETWEEN 3 AND 50),
    email TEXT NOT NULL CHECK (char_length(email) <= 320),
    password_hash TEXT NOT NULL,
    nickname TEXT,
    avatar_url TEXT,
    role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'frozen')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_users_username_ci
    ON users (lower(username))
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_users_email_ci
    ON users (lower(email))
    WHERE deleted_at IS NULL;

CREATE INDEX idx_users_role_status
    ON users (role, status)
    WHERE deleted_at IS NULL;

CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT NOT NULL,
    device_label TEXT,
    user_agent TEXT,
    ip_address TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    last_active_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_user_sessions_single_active
    ON user_sessions (user_id)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_user_sessions_user_created_at
    ON user_sessions (user_id, created_at DESC);

CREATE TABLE knowledge_bases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 200),
    description TEXT,
    embedding_model TEXT NOT NULL,
    prompt_template TEXT,
    retrieval_top_k INTEGER NOT NULL DEFAULT 5 CHECK (retrieval_top_k BETWEEN 1 AND 20),
    similarity_threshold DOUBLE PRECISION,
    last_indexed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_knowledge_bases_user_updated_at
    ON knowledge_bases (user_id, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT CHECK (name IS NULL OR char_length(name) <= 200),
    model TEXT NOT NULL CHECK (char_length(model) <= 100),
    knowledge_base_id UUID REFERENCES knowledge_bases(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_sessions_user_updated_at
    ON sessions (user_id, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    reply_to_message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
    content TEXT NOT NULL CHECK (char_length(btrim(content)) > 0),
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('completed', 'failed')),
    model_used TEXT,
    grounded BOOLEAN NOT NULL DEFAULT FALSE,
    prompt_tokens INTEGER NOT NULL DEFAULT 0 CHECK (prompt_tokens >= 0),
    completion_tokens INTEGER NOT NULL DEFAULT 0 CHECK (completion_tokens >= 0),
    total_tokens INTEGER NOT NULL DEFAULT 0 CHECK (total_tokens >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_messages_session_created_at
    ON messages (session_id, created_at ASC);

CREATE INDEX idx_messages_reply_to
    ON messages (reply_to_message_id)
    WHERE reply_to_message_id IS NOT NULL;

CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    storage_provider TEXT NOT NULL,
    bucket_name TEXT NOT NULL,
    object_key TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    sha256 TEXT NOT NULL CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_files_storage_object
    ON files (storage_provider, bucket_name, object_key);

CREATE INDEX idx_files_user_created_at
    ON files (user_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_base_id UUID NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    file_id UUID REFERENCES files(id) ON DELETE SET NULL,
    title TEXT CHECK (title IS NULL OR char_length(title) <= 500),
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'available', 'failed', 'deleting')),
    error_message TEXT,
    content_text TEXT,
    content_length INTEGER NOT NULL DEFAULT 0 CHECK (content_length >= 0),
    chunk_count INTEGER NOT NULL DEFAULT 0 CHECK (chunk_count >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_documents_kb_created_at
    ON documents (knowledge_base_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_documents_kb_file_active
    ON documents (knowledge_base_id, file_id)
    WHERE deleted_at IS NULL AND file_id IS NOT NULL;

CREATE TABLE document_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    knowledge_base_id UUID NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL CHECK (chunk_index >= 0),
    heading_path TEXT,
    content TEXT NOT NULL CHECK (char_length(btrim(content)) > 0),
    token_count INTEGER NOT NULL DEFAULT 0 CHECK (token_count >= 0),
    source_page INTEGER CHECK (source_page IS NULL OR source_page > 0),
    embedding VECTOR(1536) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_document_chunks_document_index
    ON document_chunks (document_id, chunk_index);

CREATE INDEX idx_document_chunks_kb_document
    ON document_chunks (knowledge_base_id, document_id);

CREATE INDEX idx_document_chunks_embedding_hnsw
    ON document_chunks
    USING hnsw (embedding vector_cosine_ops);

CREATE TABLE message_citations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    document_chunk_id UUID NOT NULL REFERENCES document_chunks(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    knowledge_base_id UUID NOT NULL REFERENCES knowledge_bases(id) ON DELETE CASCADE,
    rank_no INTEGER NOT NULL CHECK (rank_no >= 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_message_citations_message_rank
    ON message_citations (message_id, rank_no);

CREATE INDEX idx_message_citations_message_id
    ON message_citations (message_id);

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_type TEXT NOT NULL
        CHECK (task_type IN ('document_ingest', 'knowledge_base_reindex', 'resource_cleanup')),
    resource_type TEXT NOT NULL
        CHECK (resource_type IN ('document', 'knowledge_base', 'session', 'file', 'system')),
    resource_id UUID,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed', 'cancelled')),
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    result JSONB NOT NULL DEFAULT '{}'::jsonb,
    attempt_count INTEGER NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
    max_attempts INTEGER NOT NULL DEFAULT 3 CHECK (max_attempts >= 0),
    next_run_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    error_code TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tasks_status_next_run_at
    ON tasks (status, next_run_at, created_at);

CREATE INDEX idx_tasks_resource
    ON tasks (resource_type, resource_id);

CREATE TABLE provider_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL,
    base_url TEXT NOT NULL,
    default_chat_model TEXT NOT NULL,
    default_embedding_model TEXT NOT NULL,
    encrypted_api_key TEXT NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_provider_configs_provider
    ON provider_configs (provider);

CREATE TABLE system_settings (
    key TEXT PRIMARY KEY CHECK (char_length(key) BETWEEN 1 AND 100),
    category TEXT NOT NULL CHECK (char_length(category) BETWEEN 1 AND 100),
    value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_system_settings_category
    ON system_settings (category);

CREATE TABLE quota_policies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope_type TEXT NOT NULL CHECK (scope_type IN ('system_default', 'user')),
    scope_id UUID,
    daily_total_tokens_limit BIGINT NOT NULL CHECK (daily_total_tokens_limit > 0),
    storage_bytes_limit BIGINT NOT NULL CHECK (storage_bytes_limit > 0),
    document_count_limit INTEGER NOT NULL CHECK (document_count_limit > 0),
    warn_ratio NUMERIC(5,4) NOT NULL DEFAULT 0.8000 CHECK (warn_ratio > 0 AND warn_ratio < 1),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_quota_policies_scope_consistency CHECK (
        (scope_type = 'system_default' AND scope_id IS NULL) OR
        (scope_type = 'user' AND scope_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX uq_quota_policies_system_default
    ON quota_policies (scope_type)
    WHERE scope_type = 'system_default';

CREATE UNIQUE INDEX uq_quota_policies_user
    ON quota_policies (scope_id)
    WHERE scope_type = 'user';

CREATE TABLE daily_usage_counters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    usage_date DATE NOT NULL,
    request_count INTEGER NOT NULL DEFAULT 0 CHECK (request_count >= 0),
    prompt_tokens BIGINT NOT NULL DEFAULT 0 CHECK (prompt_tokens >= 0),
    completion_tokens BIGINT NOT NULL DEFAULT 0 CHECK (completion_tokens >= 0),
    total_tokens BIGINT NOT NULL DEFAULT 0 CHECK (total_tokens >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_daily_usage_user_date
    ON daily_usage_counters (user_id, usage_date);

CREATE TABLE user_resource_usage (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    knowledge_base_count INTEGER NOT NULL DEFAULT 0 CHECK (knowledge_base_count >= 0),
    document_count INTEGER NOT NULL DEFAULT 0 CHECK (document_count >= 0),
    storage_bytes BIGINT NOT NULL DEFAULT 0 CHECK (storage_bytes >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_role TEXT NOT NULL CHECK (actor_role IN ('user', 'admin', 'system')),
    action TEXT NOT NULL,
    resource_type TEXT,
    resource_id UUID,
    target_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    result TEXT NOT NULL CHECK (result IN ('success', 'failure')),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_actor_created_at
    ON audit_logs (actor_user_id, created_at DESC);

CREATE INDEX idx_audit_logs_action_created_at
    ON audit_logs (action, created_at DESC);

CREATE TRIGGER trg_users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_user_sessions_set_updated_at
    BEFORE UPDATE ON user_sessions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_knowledge_bases_set_updated_at
    BEFORE UPDATE ON knowledge_bases
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_sessions_set_updated_at
    BEFORE UPDATE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_messages_set_updated_at
    BEFORE UPDATE ON messages
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_documents_set_updated_at
    BEFORE UPDATE ON documents
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_tasks_set_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_provider_configs_set_updated_at
    BEFORE UPDATE ON provider_configs
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_system_settings_set_updated_at
    BEFORE UPDATE ON system_settings
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_quota_policies_set_updated_at
    BEFORE UPDATE ON quota_policies
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_daily_usage_counters_set_updated_at
    BEFORE UPDATE ON daily_usage_counters
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_user_resource_usage_set_updated_at
    BEFORE UPDATE ON user_resource_usage
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
