import dayjs from 'dayjs';

/** Formats a duration in milliseconds for display, e.g. `123 ms` or `1.20 s`. */
export function formatDuration(ms: number): string {
  if (ms < 1000) {
    return `${ms} ms`;
  }
  return `${(ms / 1000).toFixed(2)} s`;
}

/** Formats an ISO timestamp as local time. */
export function formatLocal(iso: string): string {
  return dayjs(iso).format('YYYY-MM-DD HH:mm:ss');
}

/** Formats a Date as the ISO string the backend expects in `from`/`to` filters. */
export function toIso(date: Date | null): string | undefined {
  return date ? date.toISOString() : undefined;
}

/** Colour name for an HTTP status code. */
export function statusColor(status: number): string {
  if (status === 0) {
    return 'gray';
  }
  if (status < 300) {
    return 'green';
  }
  if (status < 400) {
    return 'blue';
  }
  if (status < 500) {
    return 'orange';
  }
  return 'red';
}
