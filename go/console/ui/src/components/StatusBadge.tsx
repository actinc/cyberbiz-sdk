import { Badge } from '@mantine/core';
import type { JSX } from 'react';
import { statusColor } from '../utils/format';

/** Props for {@link StatusBadge}. */
export interface StatusBadgeProps {
  /** HTTP status code; 0 means the request never got a reply. */
  status: number;
}

/** Colour-coded HTTP status badge (2xx green, 3xx blue, 4xx orange, 5xx red). */
export function StatusBadge({ status }: StatusBadgeProps): JSX.Element {
  return (
    <Badge color={statusColor(status)} variant="light" radius="sm">
      {status === 0 ? 'no reply' : status}
    </Badge>
  );
}
