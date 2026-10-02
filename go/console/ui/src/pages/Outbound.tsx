import {
  Alert,
  Anchor,
  Badge,
  Button,
  Group,
  Loader,
  NumberInput,
  Select,
  Stack,
  Tabs,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { IconDownload, IconSearch } from '@tabler/icons-react';
import { useState, type JSX } from 'react';
import { errorMessage } from '../api/client';
import { useOutbound, useOutboundDetail } from '../api/hooks';
import type { OutboundFilter, OutboundLog, OutboundLogSummary } from '../api/types';
import { DateRangeFilter } from '../components/DateRangeFilter';
import { DateTime } from '../components/DateTime';
import { EmptyState } from '../components/EmptyState';
import { GoldenFileButton } from '../components/GoldenFileButton';
import { HeadersTable } from '../components/HeadersTable';
import { JsonView } from '../components/JsonView';
import { LogDrawer } from '../components/LogDrawer';
import { MethodBadge } from '../components/MethodBadge';
import { PaginationBar } from '../components/PaginationBar';
import { ResponsiveTable, type Column } from '../components/ResponsiveTable';
import { ShopName } from '../components/ShopName';
import { ShopSelect } from '../components/ShopSelect';
import { StatusBadge } from '../components/StatusBadge';
import { formatDuration } from '../utils/format';
import { headerValue, isBinaryContentType } from '../utils/json';

const METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'];

/** Props for {@link OutboundDetail}. */
interface OutboundDetailProps {
  log: OutboundLog;
}

/** Request/Response tabs for one Outbound. */
function OutboundDetail({ log }: OutboundDetailProps): JSX.Element {
  const contentType = headerValue(log.response_headers, 'Content-Type');
  const binary = isBinaryContentType(contentType);
  const url = log.query ? `${log.path}?${log.query}` : log.path;

  return (
    <Stack>
      <Group gap="xs" wrap="wrap">
        <MethodBadge method={log.method} />
        <Text ff="monospace" size="sm" style={{ wordBreak: 'break-all' }}>
          {url}
        </Text>
      </Group>
      <Group gap="md" wrap="wrap">
        <StatusBadge status={log.response_status} />
        <Text size="sm">{formatDuration(log.duration_ms)}</Text>
        <Text size="sm" c="dimmed">
          attempt {log.attempt}
        </Text>
        {log.request_id !== '' && (
          <Text size="sm" c="dimmed" ff="monospace">
            request id {log.request_id}
          </Text>
        )}
        <Text size="sm" c="dimmed">
          <ShopName shopId={log.shop_id} />
        </Text>
        <DateTime value={log.created_at} />
      </Group>
      {log.error !== null && log.error !== '' && (
        <Alert color="red" title="Transport error">
          {log.error}
        </Alert>
      )}
      <Group>
        <GoldenFileButton kind="outbound" id={log.id} />
      </Group>
      <Tabs defaultValue="response">
        <Tabs.List>
          <Tabs.Tab value="request">Request</Tabs.Tab>
          <Tabs.Tab value="response">Response</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="request" pt="sm">
          <Stack>
            <HeadersTable headers={log.request_headers} />
            <JsonView value={log.request_body} title="Request body" />
          </Stack>
        </Tabs.Panel>
        <Tabs.Panel value="response" pt="sm">
          <Stack>
            <HeadersTable headers={log.response_headers} />
            {binary ? (
              <Alert color="blue" variant="light" title="Binary response">
                <Text size="sm">
                  {contentType} ({log.response_body.length} base64 chars).{' '}
                  <Anchor href={`/api/outbound/${log.id}/body`} download>
                    <IconDownload size={14} style={{ verticalAlign: 'middle' }} /> Download
                  </Anchor>
                </Text>
              </Alert>
            ) : (
              <JsonView value={log.response_body} title="Response body" />
            )}
          </Stack>
        </Tabs.Panel>
      </Tabs>
    </Stack>
  );
}

/** Outbound list with filters, pagination and a detail drawer. */
export function OutboundPage(): JSX.Element {
  const [filter, setFilter] = useState<OutboundFilter>({ page: 1, per_page: 20 });
  const [selected, setSelected] = useState<number | null>(null);
  const list = useOutbound(filter);
  const detail = useOutboundDetail(selected);

  const patch = (changes: Partial<OutboundFilter>): void =>
    setFilter((prev) => ({ ...prev, ...changes, page: 1 }));

  const columns: Column<OutboundLogSummary>[] = [
    { key: 'time', header: 'Time', render: (r) => <DateTime value={r.created_at} /> },
    { key: 'shop', header: 'Shop', render: (r) => <ShopName shopId={r.shop_id} /> },
    {
      key: 'method',
      header: 'Method',
      hideLabelOnMobile: true,
      render: (r) => <MethodBadge method={r.method} />,
    },
    {
      key: 'path',
      header: 'Path',
      render: (r) => (
        <Text size="sm" ff="monospace" style={{ wordBreak: 'break-all' }}>
          {r.path}
          {r.query !== '' && (
            <Text span c="dimmed">
              ?{r.query}
            </Text>
          )}
        </Text>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      hideLabelOnMobile: true,
      render: (r) => (
        <Group gap={4} wrap="nowrap">
          <StatusBadge status={r.response_status} />
          {r.attempt > 1 && (
            <Badge size="xs" variant="outline" color="gray">
              retry {r.attempt}
            </Badge>
          )}
        </Group>
      ),
    },
    {
      key: 'duration',
      header: 'Duration',
      render: (r) => <Text size="sm">{formatDuration(r.duration_ms)}</Text>,
    },
  ];

  return (
    <Stack>
      <Title order={2}>Outbound</Title>

      <Group gap="sm" align="flex-end" wrap="wrap">
        <ShopSelect
          value={filter.shop_id ?? null}
          onChange={(id) => patch({ shop_id: id ?? undefined })}
          clearable
          placeholder="All Shops"
          size="sm"
          w={{ base: '100%', sm: 220 }}
        />
        <Select
          data={METHODS}
          value={filter.method ?? null}
          onChange={(v) => patch({ method: v ?? undefined })}
          placeholder="Method"
          clearable
          size="sm"
          w={{ base: '48%', sm: 120 }}
        />
        <NumberInput
          value={filter.status ?? ''}
          onChange={(v) => patch({ status: typeof v === 'number' ? v : undefined })}
          placeholder="Status"
          min={0}
          max={599}
          size="sm"
          w={{ base: '48%', sm: 100 }}
        />
        <DateRangeFilter
          value={{ from: filter.from, to: filter.to }}
          onChange={(r) => patch({ from: r.from, to: r.to })}
        />
        <TextInput
          value={filter.path ?? ''}
          onChange={(e) => patch({ path: e.currentTarget.value || undefined })}
          placeholder="Path contains"
          size="sm"
          w={{ base: '100%', sm: 200 }}
        />
        <TextInput
          value={filter.q ?? ''}
          onChange={(e) => patch({ q: e.currentTarget.value || undefined })}
          placeholder="Search bodies"
          leftSection={<IconSearch size={14} />}
          size="sm"
          w={{ base: '100%', sm: 200 }}
        />
        <Button
          variant="subtle"
          size="sm"
          onClick={() => setFilter({ page: 1, per_page: filter.per_page })}
        >
          Reset
        </Button>
        {list.isFetching && <Loader size="xs" />}
      </Group>

      {list.isError && <Text c="red">{errorMessage(list.error)}</Text>}
      {list.data?.items.length === 0 ? (
        <EmptyState
          title="No Outbound recorded"
          description="Send a request from the API tester or verify a Shop; every attempt is recorded here."
        />
      ) : (
        <ResponsiveTable
          columns={columns}
          rows={list.data?.items ?? []}
          rowKey={(r) => r.id}
          onRowClick={(r) => setSelected(r.id)}
          minWidth={900}
        />
      )}
      {list.data && (
        <PaginationBar
          total={list.data.total}
          page={filter.page ?? 1}
          perPage={filter.per_page ?? 20}
          onPageChange={(page) => setFilter((prev) => ({ ...prev, page }))}
          onPerPageChange={(per_page) => setFilter((prev) => ({ ...prev, per_page, page: 1 }))}
        />
      )}

      <LogDrawer
        opened={selected !== null}
        onClose={() => setSelected(null)}
        title={selected !== null ? `Outbound #${selected}` : ''}
        loading={detail.isPending && selected !== null}
        error={detail.error}
      >
        {detail.data && <OutboundDetail log={detail.data} />}
      </LogDrawer>
    </Stack>
  );
}
