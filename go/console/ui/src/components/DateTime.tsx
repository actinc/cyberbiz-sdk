import { Text, Tooltip } from '@mantine/core';
import type { JSX } from 'react';
import { formatLocal } from '../utils/format';

/** Props for {@link DateTime}. */
export interface DateTimeProps {
  /** ISO 8601 timestamp. */
  value: string;
}

/** Local time with the raw ISO value in a tooltip. */
export function DateTime({ value }: DateTimeProps): JSX.Element {
  return (
    <Tooltip label={value} withArrow>
      <Text size="sm" ff="monospace" style={{ whiteSpace: 'nowrap' }}>
        {formatLocal(value)}
      </Text>
    </Tooltip>
  );
}
