import {
  Alert,
  Badge,
  Button,
  Code,
  Group,
  Loader,
  Select,
  Stack,
  Table,
  Tabs,
  Text,
  TextInput,
  Title,
} from '@mantine/core';
import { IconSearch } from '@tabler/icons-react';
import { useState, type JSX } from 'react';
import { errorMessage } from '../api/client';
import { useInbound, useInboundDetail } from '../api/hooks';
import type { InboundFilter, InboundLog, InboundLogSummary, InboundStatus } from '../api/types';
import { DateRangeFilter } from '../components/DateRangeFilter';
import { DateTime } from '../components/DateTime';
import { EmptyState } from '../components/EmptyState';
import { GoldenFileButton } from '../components/GoldenFileButton';
import { HeadersTable } from '../components/HeadersTable';
import { InboundStatusBadge } from '../components/InboundStatusBadge';
import { JsonView } from '../components/JsonView';
import { LogDrawer } from '../components/LogDrawer';
import { PaginationBar } from '../components/PaginationBar';
import { ResponsiveTable, type Column } from '../components/ResponsiveTable';
import { ShopName } from '../components/ShopName';
import { ShopSelect } from '../components/ShopSelect';
import { StatusBadge } from '../components/StatusBadge';

const STATUSES: { value: InboundStatus; label: string }[] = [
  { value: 'valid', label: 'valid' },
  { value: 'invalid_signature', label: 'invalid signature' },
  { value: 'unknown_shop', label: 'unknown shop' },
  { value: 'malformed', label: 'malformed' },
];

/** Yes/no/unknown text for a nullable boolean. */
function yesNo(value: boolean | null): JSX.Element {
  if (value === null) {
    return (
      <Text size="sm" c="dimmed">
        header absent
      </Text>
    );
  }
  return (
    <Text size="sm" c={value ? 'green' : 'red'} fw={500}>
      {value ? 'valid' : 'invalid'}
    </Text>
  );
}

/** Props for {@link InboundDetail}. */
interface InboundDetailProps {
  log: InboundLog;
}

/** Headers / Body / Verification tabs for one Inbound. */
function InboundDetail({ log }: InboundDetailProps): JSX.Element {
  return (
    <Stack>
      <Group gap="md" wrap="wrap">
        <Badge variant="filled" color="grape" radius="sm">
          {log.event || '(no event)'}
        </Badge>
        <InboundStatusBadge status={log.status} />
        <StatusBadge status={log.response_status} />
        {log.duplicate_of !== null && (
          <Badge color="yellow" variant="light" radius="sm">
            duplicate of #{log.duplicate_of}
          </Badge>
        )}
        <DateTime value={log.created_at} />
      </Group>
      <Group gap="md" wrap="wrap">
        <Text size="sm">
          Shop Domain: <Code>{log.shop_domain || '—'}</Code>
        </Text>
        {log.custom_domain !== '' && (
          <Text size="sm">
            Custom domain: <Code>{log.custom_domain}</Code>
          </Text>
        )}
        <Text size="sm" c="dimmed">
          <ShopName shopId={log.shop_id} fallback="no matching Shop" />
        </Text>
        <Text size="sm" c="dimmed">
          from {log.remote_addr}
        </Text>
      </Group>
      <Group>
        <GoldenFileButton kind="inbound" id={log.id} />
      </Group>
      <Tabs defaultValue="body">
        <Tabs.List>
          <Tabs.Tab value="headers">Headers</Tabs.Tab>
          <Tabs.Tab value="body">Body</Tabs.Tab>
          <Tabs.Tab value="verification">Verification</Tabs.Tab>
        </Tabs.List>
        <Tabs.Panel value="headers" pt="sm">
          <HeadersTable headers={log.headers} />
        </Tabs.Panel>
        <Tabs.Panel value="body" pt="sm">
          <JsonView value={log.body} title="Inbound body" />
        </Tabs.Panel>
        <Tabs.Panel value="verification" pt="sm">
          <Stack>
            {log.status !== 'valid' && (
              <Alert color={log.status === 'malformed' ? 'gray' : 'red'} variant="light">
                {log.status === 'invalid_signature' &&
                  'The Signature did not match the HMAC-SHA256 of the body keyed by the App Secret.'}
                {log.status === 'unknown_shop' &&
                  'No Shop matched X-Cyberbiz-Domain, so the Signature could not be checked.'}
                {log.status === 'malformed' && 'Missing headers, oversized body, or invalid JSON.'}
              </Alert>
            )}
            <Table withTableBorder fz="sm">
              <Table.Tbody>
                <Table.Tr>
                  <Table.Td fw={500} w={200}>
                    Signature
                  </Table.Td>
                  <Table.Td>{yesNo(log.signature_valid)}</Table.Td>
                </Table.Tr>
                <Table.Tr>
                  <Table.Td fw={500}>X-Cyberbiz-Hmac-Sha256</Table.Td>
                  <Table.Td ff="monospace" style={{ wordBreak: 'break-all' }}>
                    {log.signature || '—'}
                  </Table.Td>
                </Table.Tr>
                <Table.Tr>
                  <Table.Td fw={500}>Domain Signature</Table.Td>
                  <Table.Td>{yesNo(log.domain_signature_valid)}</Table.Td>
                </Table.Tr>
                <Table.Tr>
                  <Table.Td fw={500}>X-Cyberbiz-Domain-Hmac-Sha256</Table.Td>
                  <Table.Td ff="monospace" style={{ wordBreak: 'break-all' }}>
                    {log.domain_signature || '—'}
                  </Table.Td>
                </Table.Tr>
                <Table.Tr>
                  <Table.Td fw={500}>Response returned</Table.Td>
                  <Table.Td>
                    <StatusBadge status={log.response_status} />
                  </Table.Td>
                </Table.Tr>
              </Table.Tbody>
            </Table>
          </Stack>
        </Tabs.Panel>
      </Tabs>
    </Stack>
  );
}

/** Inbound list with filters, pagination and a detail drawer. */
export function InboundPage(): JSX.Element {
  const [filter, setFilter] = useState<InboundFilter>({ page: 1, per_page: 20 });
  const [selected, setSelected] = useState<number | null>(null);
  const list = useInbound(filter);
  const detail = useInboundDetail(selected);

  const patch = (changes: Partial<InboundFilter>): void =>
    setFilter((prev) => ({ ...prev, ...changes, page: 1 }));

  const columns: Column<InboundLogSummary>[] = [
    { key: 'time', header: 'Time', render: (r) => <DateTime value={r.created_at} /> },
    {
      key: 'domain',
      header: 'Shop Domain',
      render: (r) => (
        <Stack gap={0}>
          <Text size="sm" ff="monospace">
            {r.shop_domain || '—'}
          </Text>
          <ShopName shopId={r.shop_id} fallback="" />
        </Stack>
      ),
    },
    {
      key: 'event',
      header: 'Event',
      render: (r) => (
        <Text size="sm" ff="monospace">
          {r.event || '—'}
        </Text>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      hideLabelOnMobile: true,
      render: (r) => (
        <Group gap={4} wrap="nowrap">
          <InboundStatusBadge status={r.status} />
          {r.duplicate_of !== null && (
            <Badge size="xs" color="yellow" variant="light">
              duplicate
            </Badge>
          )}
        </Group>
      ),
    },
    {
      key: 'response',
      header: 'Response',
      hideLabelOnMobile: true,
      render: (r) => <StatusBadge status={r.response_status} />,
    },
  ];

  return (
    <Stack>
      <Title order={2}>Inbound</Title>

      <Group gap="sm" align="flex-end" wrap="wrap">
        <ShopSelect
          value={filter.shop_id ?? null}
          onChange={(id) => patch({ shop_id: id ?? undefined })}
          clearable
          placeholder="All Shops"
          size="sm"
          w={{ base: '100%', sm: 220 }}
        />
        <TextInput
          value={filter.event ?? ''}
          onChange={(e) => patch({ event: e.currentTarget.value || undefined })}
          placeholder="Event, e.g. orders/paid"
          size="sm"
          w={{ base: '48%', sm: 200 }}
        />
        <Select
          data={STATUSES}
          value={filter.status ?? null}
          onChange={(v) => patch({ status: (v as InboundStatus | null) ?? undefined })}
          placeholder="Status"
          clearable
          size="sm"
          w={{ base: '48%', sm: 170 }}
        />
        <DateRangeFilter
          value={{ from: filter.from, to: filter.to }}
          onChange={(r) => patch({ from: r.from, to: r.to })}
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
          title="No Inbound received"
          description="Point CYBERBIZ at the webhook URL shown on the Dashboard; every delivery is recorded here."
        />
      ) : (
        <ResponsiveTable
          columns={columns}
          rows={list.data?.items ?? []}
          rowKey={(r) => r.id}
          onRowClick={(r) => setSelected(r.id)}
          minWidth={800}
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
        title={selected !== null ? `Inbound #${selected}` : ''}
        loading={detail.isPending && selected !== null}
        error={detail.error}
      >
        {detail.data && <InboundDetail log={detail.data} />}
      </LogDrawer>
    </Stack>
  );
}
