import { Badge } from '@mantine/core';
import type { JSX } from 'react';

const colors: Record<string, string> = {
  GET: 'blue',
  POST: 'green',
  PUT: 'orange',
  PATCH: 'yellow',
  DELETE: 'red',
};

/** Props for {@link MethodBadge}. */
export interface MethodBadgeProps {
  /** HTTP method. */
  method: string;
}

/** Colour-coded HTTP method badge with a fixed width so lists line up. */
export function MethodBadge({ method }: MethodBadgeProps): JSX.Element {
  return (
    <Badge
      color={colors[method] ?? 'gray'}
      variant="filled"
      radius="sm"
      w={64}
      style={{ flexShrink: 0 }}
    >
      {method}
    </Badge>
  );
}
