import type { Envelope, JsonValue } from './types';
import { ErrorCode } from './types';

/** Path the browser is sent to when a request fails with 40100. */
export const LOGIN_PATH = '/login';

/** Error thrown by the API client for any failed request. */
export class ApiError extends Error {
  /** Backend error code (40000, 40100, ...). 0 when the reply was not a valid envelope. */
  readonly code: number;
  /** HTTP status of the reply. */
  readonly status: number;
  /** Extra payload some errors carry (e.g. upstream status and messages on 50200). */
  readonly data: JsonValue | undefined;

  constructor(message: string, code: number, status: number, data?: JsonValue) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
    this.data = data;
  }
}

/** True when the value is an {@link ApiError} with the given code. */
export function isApiError(err: unknown, code?: number): err is ApiError {
  return err instanceof ApiError && (code === undefined || err.code === code);
}

/** Converts any thrown value into a display message. */
export function errorMessage(err: unknown): string {
  if (err instanceof Error) {
    return err.message;
  }
  return String(err);
}

/** Options accepted by {@link request}. */
export interface RequestOptions {
  /** HTTP method, GET by default. */
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE';
  /** JSON-serialised as the request body when present. */
  body?: unknown;
  /** Query parameters appended to the URL; undefined and empty values are dropped. */
  query?: Record<string, string | number | boolean | undefined>;
  /** Skip the redirect to /login on 40100 (used by the auth probe itself). */
  noRedirect?: boolean;
}

/** Redirect hook, replaceable in tests. */
export interface Navigator {
  /** Sends the browser to the login page. */
  toLogin: () => void;
}

const defaultNavigator: Navigator = {
  toLogin: () => {
    if (typeof window !== 'undefined' && window.location.pathname !== LOGIN_PATH) {
      window.location.assign(LOGIN_PATH);
    }
  },
};

let navigator: Navigator = defaultNavigator;

/** Replaces the redirect hook; returns a function that restores the previous one. */
export function setNavigator(next: Navigator): () => void {
  const prev = navigator;
  navigator = next;
  return () => {
    navigator = prev;
  };
}

/** Builds a URL with query parameters, dropping undefined and empty values. */
export function buildUrl(path: string, query?: RequestOptions['query']): string {
  if (!query) {
    return path;
  }
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === '') {
      continue;
    }
    params.set(key, String(value));
  }
  const qs = params.toString();
  return qs ? `${path}?${qs}` : path;
}

/** Parses a reply body as an envelope; returns null when it is not JSON. */
async function parseEnvelope(res: Response): Promise<Envelope<unknown> | null> {
  const text = await res.text();
  if (!text) {
    return null;
  }
  try {
    const parsed: unknown = JSON.parse(text);
    if (typeof parsed === 'object' && parsed !== null && 'success' in parsed) {
      return parsed as Envelope<unknown>;
    }
    return null;
  } catch {
    return null;
  }
}

/**
 * Sends a request to the Console backend and unwraps the envelope.
 * Throws {@link ApiError} on any failure; redirects to /login on 40100.
 */
export async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json' };
  const init: RequestInit = {
    method: options.method ?? 'GET',
    credentials: 'same-origin',
    headers,
  };
  if (options.body !== undefined) {
    headers['Content-Type'] = 'application/json';
    init.body = JSON.stringify(options.body);
  }
  const res = await fetch(buildUrl(path, options.query), init);
  const envelope = await parseEnvelope(res);

  if (envelope?.success === true) {
    return envelope.data as T;
  }

  const code = envelope?.code ?? (res.status === 401 ? ErrorCode.Unauthenticated : 0);
  const message = envelope?.error ?? `HTTP ${res.status} ${res.statusText}`.trim();
  const data = (envelope as { data?: JsonValue } | null)?.data;
  if (code === ErrorCode.Unauthenticated && !options.noRedirect) {
    navigator.toLogin();
  }
  throw new ApiError(message, code, res.status, data);
}

/** Convenience wrappers around {@link request}. */
export const api = {
  /** GET with optional query. */
  get: <T>(path: string, query?: RequestOptions['query'], noRedirect = false): Promise<T> =>
    request<T>(path, { query, noRedirect }),
  /** POST with a JSON body. */
  post: <T>(path: string, body?: unknown): Promise<T> => request<T>(path, { method: 'POST', body }),
  /** PUT with a JSON body. */
  put: <T>(path: string, body?: unknown): Promise<T> => request<T>(path, { method: 'PUT', body }),
  /** DELETE. */
  delete: <T>(path: string): Promise<T> => request<T>(path, { method: 'DELETE' }),
};
