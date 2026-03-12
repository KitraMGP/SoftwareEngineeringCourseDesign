-- +goose Up
INSERT INTO system_settings (key, category, value, description)
VALUES
    (
        'auth.access_token_ttl_seconds',
        'auth',
        '1800'::jsonb,
        'Access token validity period in seconds'
    ),
    (
        'auth.refresh_token_ttl_seconds',
        'auth',
        '604800'::jsonb,
        'Refresh token validity period in seconds'
    ),
    (
        'rag.default_retrieval_top_k',
        'rag',
        '5'::jsonb,
        'Default number of retrieved chunks'
    ),
    (
        'rag.max_context_chunks',
        'rag',
        '5'::jsonb,
        'Maximum number of retrieved chunks injected into the prompt'
    ),
    (
        'rag.chunk_target_tokens',
        'rag',
        '500'::jsonb,
        'Target token count for document chunking'
    ),
    (
        'rag.chunk_overlap_tokens',
        'rag',
        '80'::jsonb,
        'Overlap token count between neighboring chunks'
    ),
    (
        'rag.history_rounds',
        'rag',
        '5'::jsonb,
        'Number of recent chat rounds included in prompt assembly'
    ),
    (
        'upload.max_file_size_bytes',
        'upload',
        '20971520'::jsonb,
        'Maximum single upload file size in bytes'
    ),
    (
        'sse.heartbeat_seconds',
        'sse',
        '15'::jsonb,
        'SSE heartbeat interval in seconds'
    ),
    (
        'app.soft_delete_retention_days',
        'app',
        '7'::jsonb,
        'Soft delete retention window before asynchronous cleanup'
    )
ON CONFLICT (key) DO UPDATE
SET
    category = EXCLUDED.category,
    value = EXCLUDED.value,
    description = EXCLUDED.description,
    updated_at = CURRENT_TIMESTAMP;

-- +goose Down
DELETE FROM system_settings
WHERE key IN (
    'auth.access_token_ttl_seconds',
    'auth.refresh_token_ttl_seconds',
    'rag.default_retrieval_top_k',
    'rag.max_context_chunks',
    'rag.chunk_target_tokens',
    'rag.chunk_overlap_tokens',
    'rag.history_rounds',
    'upload.max_file_size_bytes',
    'sse.heartbeat_seconds',
    'app.soft_delete_retention_days'
);
