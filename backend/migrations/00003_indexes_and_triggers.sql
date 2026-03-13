-- +goose Up
CREATE UNIQUE INDEX uq_users_username_ci
    ON users (lower(username))
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_users_email_ci
    ON users (lower(email))
    WHERE deleted_at IS NULL;

CREATE INDEX idx_users_role_status
    ON users (role, status)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_user_sessions_single_active
    ON user_sessions (user_id)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_user_sessions_user_created_at
    ON user_sessions (user_id, created_at DESC);

CREATE UNIQUE INDEX uq_provider_configs_provider
    ON provider_configs (provider);

CREATE INDEX idx_system_settings_category
    ON system_settings (category);

CREATE INDEX idx_knowledge_bases_user_updated_at
    ON knowledge_bases (user_id, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_sessions_user_updated_at
    ON sessions (user_id, updated_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_messages_session_created_at
    ON messages (session_id, created_at ASC);

CREATE INDEX idx_messages_reply_to
    ON messages (reply_to_message_id)
    WHERE reply_to_message_id IS NOT NULL;

CREATE UNIQUE INDEX uq_files_storage_object
    ON files (storage_provider, bucket_name, object_key);

CREATE INDEX idx_files_user_created_at
    ON files (user_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_documents_kb_created_at
    ON documents (knowledge_base_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_documents_kb_file_active
    ON documents (knowledge_base_id, file_id)
    WHERE deleted_at IS NULL AND file_id IS NOT NULL;

CREATE UNIQUE INDEX uq_document_chunks_document_index
    ON document_chunks (document_id, chunk_index);

CREATE INDEX idx_document_chunks_kb_document
    ON document_chunks (knowledge_base_id, document_id);

CREATE INDEX idx_document_chunks_embedding_hnsw
    ON document_chunks
    USING hnsw (embedding vector_cosine_ops);

CREATE UNIQUE INDEX uq_message_citations_message_rank
    ON message_citations (message_id, rank_no);

CREATE INDEX idx_message_citations_message_id
    ON message_citations (message_id);

CREATE UNIQUE INDEX uq_quota_policies_system_default
    ON quota_policies (scope_type)
    WHERE scope_type = 'system_default';

CREATE UNIQUE INDEX uq_quota_policies_user
    ON quota_policies (scope_id)
    WHERE scope_type = 'user';

CREATE UNIQUE INDEX uq_daily_usage_user_date
    ON daily_usage_counters (user_id, usage_date);

CREATE INDEX idx_tasks_status_next_run_at
    ON tasks (status, next_run_at, created_at);

CREATE INDEX idx_tasks_resource
    ON tasks (resource_type, resource_id);

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

CREATE TRIGGER trg_provider_configs_set_updated_at
    BEFORE UPDATE ON provider_configs
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

CREATE TRIGGER trg_tasks_set_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_system_settings_set_updated_at
    BEFORE UPDATE ON system_settings
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trg_system_settings_set_updated_at ON system_settings;
DROP TRIGGER IF EXISTS trg_tasks_set_updated_at ON tasks;
DROP TRIGGER IF EXISTS trg_user_resource_usage_set_updated_at ON user_resource_usage;
DROP TRIGGER IF EXISTS trg_daily_usage_counters_set_updated_at ON daily_usage_counters;
DROP TRIGGER IF EXISTS trg_quota_policies_set_updated_at ON quota_policies;
DROP TRIGGER IF EXISTS trg_documents_set_updated_at ON documents;
DROP TRIGGER IF EXISTS trg_messages_set_updated_at ON messages;
DROP TRIGGER IF EXISTS trg_sessions_set_updated_at ON sessions;
DROP TRIGGER IF EXISTS trg_knowledge_bases_set_updated_at ON knowledge_bases;
DROP TRIGGER IF EXISTS trg_provider_configs_set_updated_at ON provider_configs;
DROP TRIGGER IF EXISTS trg_user_sessions_set_updated_at ON user_sessions;
DROP TRIGGER IF EXISTS trg_users_set_updated_at ON users;

DROP INDEX IF EXISTS idx_audit_logs_action_created_at;
DROP INDEX IF EXISTS idx_audit_logs_actor_created_at;
DROP INDEX IF EXISTS idx_tasks_resource;
DROP INDEX IF EXISTS idx_tasks_status_next_run_at;
DROP INDEX IF EXISTS uq_daily_usage_user_date;
DROP INDEX IF EXISTS uq_quota_policies_user;
DROP INDEX IF EXISTS uq_quota_policies_system_default;
DROP INDEX IF EXISTS idx_message_citations_message_id;
DROP INDEX IF EXISTS uq_message_citations_message_rank;
DROP INDEX IF EXISTS idx_document_chunks_embedding_hnsw;
DROP INDEX IF EXISTS idx_document_chunks_kb_document;
DROP INDEX IF EXISTS uq_document_chunks_document_index;
DROP INDEX IF EXISTS uq_documents_kb_file_active;
DROP INDEX IF EXISTS idx_documents_kb_created_at;
DROP INDEX IF EXISTS idx_files_user_created_at;
DROP INDEX IF EXISTS uq_files_storage_object;
DROP INDEX IF EXISTS idx_messages_reply_to;
DROP INDEX IF EXISTS idx_messages_session_created_at;
DROP INDEX IF EXISTS idx_sessions_user_updated_at;
DROP INDEX IF EXISTS idx_knowledge_bases_user_updated_at;
DROP INDEX IF EXISTS idx_system_settings_category;
DROP INDEX IF EXISTS uq_provider_configs_provider;
DROP INDEX IF EXISTS idx_user_sessions_user_created_at;
DROP INDEX IF EXISTS uq_user_sessions_single_active;
DROP INDEX IF EXISTS idx_users_role_status;
DROP INDEX IF EXISTS uq_users_email_ci;
DROP INDEX IF EXISTS uq_users_username_ci;
