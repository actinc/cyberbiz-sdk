import type { JsonValue } from '../api/types';

/** Result of {@link tryParseJson}. */
export type ParseResult = { ok: true; value: JsonValue } | { ok: false; error: string };

/** Parses JSON text, returning a result instead of throwing. Empty text parses to null. */
export function tryParseJson(text: string): ParseResult {
  if (text.trim() === '') {
    return { ok: true, value: null };
  }
  try {
    return { ok: true, value: JSON.parse(text) as JsonValue };
  } catch (err) {
    return { ok: false, error: err instanceof Error ? err.message : 'Invalid JSON' };
  }
}

/** Pretty-prints any value; strings that hold JSON are re-indented, others are returned as-is. */
export function prettyJson(value: unknown): string {
  if (typeof value === 'string') {
    const parsed = tryParseJson(value);
    return parsed.ok && parsed.value !== null ? JSON.stringify(parsed.value, null, 2) : value;
  }
  return JSON.stringify(value, null, 2);
}

/** Flattens a headers object (string or string[] values) into display rows. */
export function headerRows(headers: Record<string, string | string[]> | null | undefined): {
  name: string;
  value: string;
}[] {
  if (!headers) {
    return [];
  }
  return Object.entries(headers)
    .map(([name, value]) => ({ name, value: Array.isArray(value) ? value.join(', ') : value }))
    .sort((a, b) => a.name.localeCompare(b.name));
}

/** True when the content type describes a body the UI should not render as text. */
export function isBinaryContentType(contentType: string | undefined): boolean {
  if (!contentType) {
    return false;
  }
  const ct = contentType.toLowerCase();
  return !(
    ct.includes('json') ||
    ct.startsWith('text/') ||
    ct.includes('xml') ||
    ct.includes('x-www-form-urlencoded')
  );
}

/** Looks a header up case-insensitively. */
export function headerValue(
  headers: Record<string, string | string[]> | null | undefined,
  name: string,
): string | undefined {
  if (!headers) {
    return undefined;
  }
  const key = Object.keys(headers).find((k) => k.toLowerCase() === name.toLowerCase());
  if (key === undefined) {
    return undefined;
  }
  const value = headers[key];
  return Array.isArray(value) ? value[0] : value;
}
