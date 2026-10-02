import { Badge } from '@mantine/core';
import type { JSX } from 'react';
import type { InboundStatus } from '../api/types';

const meta: Record<InboundStatus, { color: string; label: string }> = {
  valid: { color: 'green', label: 'valid' },
  invalid_signature: { color: 'red', label: 'invalid signature' },
  unknown_shop: { color: 'orange', label: 'unknown shop' },
  malformed: { color: 'gray', label: 'malformed' },
};

/** Props for {@link InboundStatusBadge}. */
export interface InboundStatusBadgeProps {
  /** Verification status of the Inbound. */
  status: InboundStatus;
}

/** Badge for an Inbound verification status. */
export function InboundStatusBadge({ status }: InboundStatusBadgeProps): JSX.Element {
  const m = meta[status];
  return (
    <Badge color={m.color} variant="light" radius="sm">
      {m.label}
    </Badge>
  );
}
