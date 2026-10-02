import { Group, Table, Text } from '@mantine/core';
import type { JSX } from 'react';
import { headerRows } from '../utils/json';
import { CopyIconButton } from './CopyIconButton';

/** Props for {@link HeadersTable}. */
export interface HeadersTableProps {
  /** Header map as stored by the backend. */
  headers: Record<string, string | string[]> | null | undefined;
}

/** Two-column table of HTTP headers with a copy-all button. */
export function HeadersTable({ headers }: HeadersTableProps): JSX.Element {
  const rows = headerRows(headers);
  if (rows.length === 0) {
    return (
      <Text size="sm" c="dimmed">
        No headers.
      </Text>
    );
  }
  const asText = rows.map((r) => `${r.name}: ${r.value}`).join('\n');
  return (
    <div>
      <Group justify="flex-end" mb={4}>
        <CopyIconButton value={asText} label="Copy headers" />
      </Group>
      <Table withTableBorder striped fz="xs">
        <Table.Tbody>
          {rows.map((r) => (
            <Table.Tr key={r.name}>
              <Table.Td ff="monospace" fw={500} style={{ whiteSpace: 'nowrap' }}>
                {r.name}
              </Table.Td>
              <Table.Td ff="monospace" style={{ wordBreak: 'break-all' }}>
                {r.value}
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </div>
  );
}
