export interface FieldError {
  field: string;
  message: string;
}

export interface ApiSuccessEnvelope<T> {
  code: number;
  message: string;
  data: T;
}

export interface ApiErrorEnvelope {
  code: number;
  message: string;
  data: null;
  request_id?: string;
  details?: FieldError[];
}

export class ApiRequestError extends Error {
  status?: number;
  code?: number;
  requestId?: string;
  details: FieldError[];

  constructor(message: string, options: Partial<ApiRequestError> = {}) {
    super(message);
    this.name = 'ApiRequestError';
    this.status = options.status;
    this.code = options.code;
    this.requestId = options.requestId;
    this.details = options.details ?? [];
  }
}
