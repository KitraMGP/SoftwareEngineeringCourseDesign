import axios from 'axios';

import { ApiRequestError, type FieldError, type ApiErrorEnvelope } from '../types/api';

export function normalizeApiError(error: unknown): ApiRequestError {
  if (error instanceof ApiRequestError) {
    return error;
  }

  if (axios.isAxiosError<ApiErrorEnvelope>(error)) {
    const payload = error.response?.data;

    return new ApiRequestError(payload?.message ?? error.message ?? '请求失败', {
      status: error.response?.status,
      code: payload?.code,
      requestId: payload?.request_id,
      details: payload?.details ?? []
    });
  }

  if (error instanceof Error) {
    return new ApiRequestError(error.message);
  }

  return new ApiRequestError('请求失败');
}

export function toFieldErrorMap(error: unknown): Record<string, string> {
  const normalized = normalizeApiError(error);

  return normalized.details.reduce<Record<string, string>>((result, detail) => {
    if (detail.field && !result[detail.field]) {
      result[detail.field] = detail.message;
    }
    return result;
  }, {});
}

export function getErrorMessage(error: unknown, fallback = '操作未完成，请稍后重试。'): string {
  const normalized = normalizeApiError(error);
  return normalized.message || fallback;
}

export function mergeFieldErrors(details: FieldError[]): Record<string, string[]> {
  return details.reduce<Record<string, string[]>>((result, detail) => {
    if (!result[detail.field]) {
      result[detail.field] = [];
    }
    result[detail.field].push(detail.message);
    return result;
  }, {});
}
