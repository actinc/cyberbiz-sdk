import { Group, Pagination, Select, Text } from '@mantine/core';
import type { JSX } from 'react';

/** Page sizes offered in the selector. */
export const PAGE_SIZES = [20, 50, 100] as const;

/** Props for {@link PaginationBar}. */
export interface PaginationBarProps {
  /** Total rows across all pages. */
  total: number;
  /** Current page (1-based). */
  page: number;
  /** Rows per page. */
  perPage: number;
  /** Page change handler. */
  onPageChange: (page: number) => void;
  /** Page size change handler. */
  onPerPageChange: (perPage: number) => void;
}

/** Pagination controls with a total count and a page-size select. */
export function PaginationBar({
  total,
  page,
  perPage,
  onPageChange,
  onPerPageChange,
}: PaginationBarProps): JSX.Element {
  const pages = Math.max(1, Math.ceil(total / perPage));
  return (
    <Group justify="space-between" wrap="wrap">
      <Text size="sm" c="dimmed">
        {total} total
      </Text>
      <Group gap="sm">
        <Select
          size="xs"
          w={90}
          value={String(perPage)}
          data={PAGE_SIZES.map((n) => ({ value: String(n), label: `${n} / page` }))}
          onChange={(v) => {
            if (v !== null) {
              onPerPageChange(Number(v));
            }
          }}
          allowDeselect={false}
        />
        <Pagination size="sm" total={pages} value={page} onChange={onPageChange} />
      </Group>
    </Group>
  );
}
