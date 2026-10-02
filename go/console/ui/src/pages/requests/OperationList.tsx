import { Group, NavLink, ScrollArea, Select, Stack, Text, TextInput } from '@mantine/core';
import { IconSearch } from '@tabler/icons-react';
import { useMemo, useState, type JSX } from 'react';
import type { Catalog, Operation } from '../../api/types';
import { EmptyState } from '../../components/EmptyState';
import { MethodBadge } from '../../components/MethodBadge';

/** Props for {@link OperationList}. */
export interface OperationListProps {
  /** Catalog from `/api/catalog`. */
  catalog: Catalog;
  /** Currently selected operation id. */
  selectedId: string | null;
  /** Selection handler. */
  onSelect: (op: Operation) => void;
  /** Max height of the scrollable list. */
  height?: number | string;
}

/** True when the operation matches the free-text query. */
function matches(op: Operation, query: string): boolean {
  if (query === '') {
    return true;
  }
  const q = query.toLowerCase();
  return (
    op.path.toLowerCase().includes(q) ||
    op.summary.toLowerCase().includes(q) ||
    op.id.toLowerCase().includes(q) ||
    op.method.toLowerCase() === q
  );
}

/** Tag filter plus searchable operation list, grouped by tag. */
export function OperationList({
  catalog,
  selectedId,
  onSelect,
  height = 'calc(100vh - 220px)',
}: OperationListProps): JSX.Element {
  const [tag, setTag] = useState<string | null>(null);
  const [query, setQuery] = useState('');

  const groups = useMemo(() => {
    const filtered = catalog.operations.filter(
      (op) => (tag === null || op.tag === tag) && matches(op, query.trim()),
    );
    const byTag = new Map<string, Operation[]>();
    for (const op of filtered) {
      const list = byTag.get(op.tag) ?? [];
      byTag.set(op.tag, [...list, op]);
    }
    const order = catalog.tags.map((t) => t.name);
    return [...byTag.entries()].sort(([a], [b]) => {
      const ia = order.indexOf(a);
      const ib = order.indexOf(b);
      return (
        (ia === -1 ? order.length : ia) - (ib === -1 ? order.length : ib) || a.localeCompare(b)
      );
    });
  }, [catalog, tag, query]);

  return (
    <Stack gap="xs" h="100%">
      <Select
        data={catalog.tags.map((t) => ({ value: t.name, label: t.name }))}
        value={tag}
        onChange={setTag}
        placeholder="All tags"
        clearable
        searchable
        size="sm"
      />
      <TextInput
        value={query}
        onChange={(e) => setQuery(e.currentTarget.value)}
        placeholder="Search path or summary"
        leftSection={<IconSearch size={14} />}
        size="sm"
      />
      <ScrollArea h={height} type="auto">
        {groups.length === 0 && <EmptyState title="No operations match" />}
        {groups.map(([tagName, ops]) => (
          <div key={tagName}>
            <Text size="xs" c="dimmed" fw={600} tt="uppercase" px="xs" pt="sm" pb={4}>
              {tagName}
            </Text>
            {ops.map((op) => (
              <NavLink
                key={op.id}
                active={op.id === selectedId}
                onClick={() => onSelect(op)}
                px="xs"
                py={6}
                label={
                  <Group gap="xs" wrap="nowrap">
                    <MethodBadge method={op.method} />
                    <Stack gap={0} style={{ minWidth: 0 }}>
                      <Text size="sm" ff="monospace" truncate>
                        {op.path}
                      </Text>
                      <Text size="xs" c="dimmed" truncate>
                        {op.summary}
                      </Text>
                    </Stack>
                  </Group>
                }
              />
            ))}
          </div>
        ))}
      </ScrollArea>
    </Stack>
  );
}
