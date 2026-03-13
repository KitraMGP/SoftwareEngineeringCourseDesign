import type {
  ChatStreamDelta,
  ChatStreamDone,
  ChatStreamError,
  ChatStreamMeta,
  CreateSessionPayload,
  CreateSessionResult,
  PaginatedResult,
  SendMessagePayload,
  Session,
  SessionDetail
} from '../types/domain';
import type { ApiErrorEnvelope } from '../types/api';

import { ApiRequestError } from '../types/api';

import { apiClient, resolveApiUrl, unwrapData } from './http';

interface StreamSessionMessageOptions {
  accessToken?: string | null;
  signal?: AbortSignal;
  refreshAccessToken?: () => Promise<string | null>;
  onUnauthorized?: () => void;
  onMeta?: (payload: ChatStreamMeta) => void;
  onDelta?: (payload: ChatStreamDelta) => void;
  onDone?: (payload: ChatStreamDone) => void;
  onError?: (payload: ChatStreamError) => void;
}

function parseStreamEvent(chunk: string): { event: string; data: string } | null {
  const lines = chunk
    .split('\n')
    .map((line) => line.replace(/\r$/, ''))
    .filter(Boolean);

  if (!lines.length || lines[0]?.startsWith(':')) {
    return null;
  }

  let event = 'message';
  const dataLines: string[] = [];

  lines.forEach((line) => {
    if (line.startsWith('event:')) {
      event = line.slice(6).trim();
      return;
    }

    if (line.startsWith('data:')) {
      dataLines.push(line.slice(5).trimStart());
    }
  });

  return {
    event,
    data: dataLines.join('\n')
  };
}

function parseEventPayload<T>(data: string): T {
  try {
    return JSON.parse(data) as T;
  } catch {
    throw new ApiRequestError('流式响应解析失败');
  }
}

async function readApiError(response: Response): Promise<ApiRequestError> {
  const contentType = response.headers.get('content-type') || '';

  if (contentType.includes('application/json')) {
    try {
      const payload = (await response.json()) as ApiErrorEnvelope;
      return new ApiRequestError(payload.message || '请求失败', {
        status: response.status,
        code: payload.code,
        requestId: payload.request_id,
        details: payload.details ?? []
      });
    } catch {
      return new ApiRequestError(`请求失败 (${response.status})`, {
        status: response.status
      });
    }
  }

  const fallbackText = await response.text().catch(() => '');
  return new ApiRequestError(fallbackText || `请求失败 (${response.status})`, {
    status: response.status
  });
}

async function consumeEventStream(
  response: Response,
  options: StreamSessionMessageOptions
): Promise<void> {
  const reader = response.body?.getReader();

  if (!reader) {
    throw new ApiRequestError('未收到流式响应');
  }

  const decoder = new TextDecoder();
  let buffer = '';
  let streamError: ApiRequestError | null = null;

  const handleChunk = (chunk: string) => {
    const parsed = parseStreamEvent(chunk);
    if (!parsed || !parsed.data) {
      return;
    }

    if (parsed.event === 'meta') {
      options.onMeta?.(parseEventPayload<ChatStreamMeta>(parsed.data));
      return;
    }

    if (parsed.event === 'delta') {
      options.onDelta?.(parseEventPayload<ChatStreamDelta>(parsed.data));
      return;
    }

    if (parsed.event === 'done') {
      options.onDone?.(parseEventPayload<ChatStreamDone>(parsed.data));
      return;
    }

    if (parsed.event === 'error') {
      const payload = parseEventPayload<ChatStreamError>(parsed.data);
      options.onError?.(payload);
      streamError = new ApiRequestError(payload.message || '消息发送失败', {
        code: payload.code,
        status: response.status
      });
    }
  };

  let isDone = false;

  while (!isDone) {
    const { done, value } = await reader.read();
    buffer += decoder.decode(value || new Uint8Array(), { stream: !done });

    let separatorIndex = buffer.indexOf('\n\n');
    while (separatorIndex !== -1) {
      const chunk = buffer.slice(0, separatorIndex).trim();
      buffer = buffer.slice(separatorIndex + 2);

      if (chunk) {
        handleChunk(chunk);
      }

      if (streamError) {
        await reader.cancel().catch(() => undefined);
        throw streamError;
      }

      separatorIndex = buffer.indexOf('\n\n');
    }

    isDone = done;
  }

  const remainingChunk = buffer.trim();
  if (remainingChunk) {
    handleChunk(remainingChunk);
  }

  if (streamError) {
    throw streamError;
  }
}

export const chatApi = {
  async listSessions(params: { page?: number; size?: number; keyword?: string } = {}) {
    return unwrapData<PaginatedResult<Session>>(await apiClient.get('/sessions', { params }));
  },

  async createSession(payload: CreateSessionPayload) {
    return unwrapData<CreateSessionResult>(await apiClient.post('/sessions', payload));
  },

  async getSessionDetail(sessionId: string) {
    return unwrapData<SessionDetail>(await apiClient.get(`/sessions/${sessionId}`));
  },

  async streamSessionMessage(
    sessionId: string,
    payload: SendMessagePayload,
    options: StreamSessionMessageOptions = {}
  ): Promise<void> {
    const doRequest = (token?: string | null) =>
      fetch(resolveApiUrl(`/sessions/${sessionId}/messages`), {
        method: 'POST',
        credentials: 'include',
        headers: {
          Accept: 'text/event-stream',
          'Content-Type': 'application/json',
          ...(token ? { Authorization: `Bearer ${token}` } : {})
        },
        body: JSON.stringify(payload),
        signal: options.signal
      });

    let response = await doRequest(options.accessToken);

    if (response.status === 401 && options.refreshAccessToken) {
      const nextToken = await options.refreshAccessToken();

      if (nextToken) {
        response = await doRequest(nextToken);
      }
    }

    if (response.status === 401) {
      options.onUnauthorized?.();
    }

    if (!response.ok) {
      throw await readApiError(response);
    }

    await consumeEventStream(response, options);
  },

  async deleteSession(sessionId: string): Promise<void> {
    await apiClient.delete(`/sessions/${sessionId}`);
  }
};
