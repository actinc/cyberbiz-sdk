import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError, api, buildUrl, isApiError, request, setNavigator } from './client';
import { ErrorCode } from './types';

/** Builds a fetch mock that answers every call with the given status and body. */
function mockFetch(status: number, body: unknown, statusText = ''): ReturnType<typeof vi.fn> {
  const fn = vi.fn(() =>
    Promise.resolve(
      new Response(body === undefined ? null : JSON.stringify(body), {
        status,
        statusText,
        headers: { 'Content-Type': 'application/json' },
      }),
    ),
  );
  vi.stubGlobal('fetch', fn);
  return fn;
}

describe('api client', () => {
  const toLogin = vi.fn();
  let restore: () => void;

  beforeEach(() => {
    toLogin.mockReset();
    restore = setNavigator({ toLogin });
  });

  afterEach(() => {
    restore();
    vi.unstubAllGlobals();
  });

  it('unwraps data from a successful envelope', async () => {
    const fetchMock = mockFetch(200, { success: true, data: { id: 7, username: 'admin' } });
    const user = await request<{ id: number; username: string }>('/api/auth/me');
    expect(user).toEqual({ id: 7, username: 'admin' });
    const [, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(init.credentials).toBe('same-origin');
  });

  it('serialises the body and sets the JSON content type on POST', async () => {
    const fetchMock = mockFetch(200, { success: true, data: {} });
    await api.post('/api/auth/login', { username: 'a', password: 'b' });
    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toBe('/api/auth/login');
    expect(init.method).toBe('POST');
    expect(init.body).toBe(JSON.stringify({ username: 'a', password: 'b' }));
    expect((init.headers as Record<string, string>)['Content-Type']).toBe('application/json');
  });

  it('throws ApiError with the envelope code, message and data', async () => {
    mockFetch(502, {
      success: false,
      error: 'CYBERBIZ upstream error',
      code: 50200,
      data: { status: 429, messages: ['rate limited'], request_id: 'req-1' },
    });
    const err = await request('/api/shops/1/verify', { method: 'POST' }).catch((e: unknown) => e);
    expect(isApiError(err, ErrorCode.Upstream)).toBe(true);
    const apiErr = err as ApiError;
    expect(apiErr.message).toBe('CYBERBIZ upstream error');
    expect(apiErr.status).toBe(502);
    expect(apiErr.data).toEqual({ status: 429, messages: ['rate limited'], request_id: 'req-1' });
    expect(toLogin).not.toHaveBeenCalled();
  });

  it('maps a 409 conflict to code 40900', async () => {
    mockFetch(409, { success: false, error: 'file exists', code: 40900 });
    const err = await api.post('/api/outbound/1/golden', {}).catch((e: unknown) => e);
    expect(isApiError(err, ErrorCode.Conflict)).toBe(true);
  });

  it('falls back to an HTTP message when the body is not an envelope', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => Promise.resolve(new Response('<html>', { status: 500, statusText: 'Boom' }))),
    );
    const err = await request('/api/stats').catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect((err as ApiError).code).toBe(0);
    expect((err as ApiError).message).toBe('HTTP 500 Boom');
  });

  it('redirects to /login on 40100', async () => {
    mockFetch(401, { success: false, error: 'not logged in', code: 40100 });
    await expect(request('/api/shops')).rejects.toMatchObject({ code: ErrorCode.Unauthenticated });
    expect(toLogin).toHaveBeenCalledTimes(1);
  });

  it('does not redirect on 40100 when noRedirect is set', async () => {
    mockFetch(401, { success: false, error: 'not logged in', code: 40100 });
    await expect(api.get('/api/auth/me', undefined, true)).rejects.toBeInstanceOf(ApiError);
    expect(toLogin).not.toHaveBeenCalled();
  });

  it('treats a bare 401 without envelope as 40100', async () => {
    mockFetch(401, undefined);
    await expect(request('/api/shops')).rejects.toMatchObject({ code: ErrorCode.Unauthenticated });
    expect(toLogin).toHaveBeenCalledTimes(1);
  });

  it('builds query strings and drops empty values', () => {
    expect(buildUrl('/api/outbound', { page: 2, q: '', shop_id: undefined, method: 'GET' })).toBe(
      '/api/outbound?page=2&method=GET',
    );
    expect(buildUrl('/api/outbound')).toBe('/api/outbound');
  });
});
