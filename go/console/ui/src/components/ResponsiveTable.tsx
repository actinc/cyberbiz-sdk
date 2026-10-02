import { Card, Group, ScrollArea, Stack, Table, Text } from '@mantine/core';
import type { JSX, ReactNode } from 'react';
import { useIsMobile } from '../hooks/useIsMobile';

/** One column of a {@link ResponsiveTable}. */
export interface Column<T> {
  /** Unique key. */
  key: string;
  /** Header label; also the label on mobile cards. */
  header: string;
  /** Cell renderer. */
  render: (row: T) => ReactNode;
  /** Hide the label on mobile cards (for cells that are self-describing, e.g. badges). */
  hideLabelOnMobile?: boolean;
  /** Column width hint for the desktop table. */
  width?: number | string;
}

/** Props for {@link ResponsiveTable}. */
export interface ResponsiveTableProps<T> {
  /** Column definitions. */
  columns: Column<T>[];
  /** Rows to render. */
  rows: T[];
  /** Stable key per row. */
  rowKey: (row: T) => string | number;
  /** Row click handler; rows get a pointer cursor when set. */
  onRowClick?: (row: T) => void;
  /** Minimum width of the desktop table before it scrolls horizontally. */
  minWidth?: number;
}

/** Table on desktop, stacked cards below the `sm` breakpoint. */
export function ResponsiveTable<T>({
  columns,
  rows,
  rowKey,
  onRowClick,
  minWidth = 800,
}: ResponsiveTableProps<T>): JSX.Element {
  const isMobile = useIsMobile();

  if (isMobile) {
    return (
      <Stack gap="sm">
        {rows.map((row) => (
          <Card
            key={rowKey(row)}
            withBorder
            padding="sm"
            onClick={onRowClick ? () => onRowClick(row) : undefined}
            style={{ cursor: onRowClick ? 'pointer' : undefined }}
          >
            <Stack gap={6}>
              {columns.map((col) => (
                <Group key={col.key} justify="space-between" wrap="nowrap" gap="xs">
                  {!col.hideLabelOnMobile && (
                    <Text size="xs" c="dimmed" style={{ flexShrink: 0 }}>
                      {col.header}
                    </Text>
                  )}
                  <div style={{ minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {col.render(row)}
                  </div>
                </Group>
              ))}
            </Stack>
          </Card>
        ))}
      </Stack>
    );
  }

  return (
    <ScrollArea>
      <Table striped highlightOnHover withTableBorder miw={minWidth}>
        <Table.Thead>
          <Table.Tr>
            {columns.map((col) => (
              <Table.Th key={col.key} w={col.width}>
                {col.header}
              </Table.Th>
            ))}
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {rows.map((row) => (
            <Table.Tr
              key={rowKey(row)}
              onClick={onRowClick ? () => onRowClick(row) : undefined}
              style={{ cursor: onRowClick ? 'pointer' : undefined }}
            >
              {columns.map((col) => (
                <Table.Td key={col.key}>{col.render(row)}</Table.Td>
              ))}
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>
    </ScrollArea>
  );
}
