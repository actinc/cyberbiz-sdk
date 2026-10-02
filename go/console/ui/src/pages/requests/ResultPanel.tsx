import { Alert, Anchor, Group, Paper, Stack, Tabs, Text, Title } from '@mantine/core';
import { IconDownload } from '@tabler/icons-react';
import type { JSX } from 'react';
import { errorMessage, isApiError } from '../../api/client';
import type { OutboundLog, UpstreamErrorData } from '../../api/types';
import { ErrorCode } from '../../api/types';
import { GoldenFileButton } from '../../components/GoldenFileButton';
import { HeadersTable } from '../../components/HeadersTable';
import { JsonView } from '../../components/JsonView';
import { StatusBadge } from '../../components/StatusBadge';
import { formatDuration } from '../../utils/format';
import { headerValue, isBinaryContentType } from '../../utils/json';

/** Props for {@link ResultPanel}. */
export interface ResultPanelProps {
  /** The recorded Outbound returned by the backend, when the request completed. */
  result: OutboundLog | null;
  /** The failure, when the request did not complete. */
  error: unknown;
}

/** Narrows an error's data to the 50200 upstream payload. */
function upstreamData(data: unknown): UpstreamErrorData | null {
  if (typeof data !== 'object' || data === null) {
    return null;
  }
  const d = data as Partial<UpstreamErrorData>;
  return typeof d.status === 'number' ? (d as UpstreamErrorData) : null;
}

/** Status, duration, headers, body and Golden File button for the last sent request. */
export function ResultPanel({ result, error }: ResultPanelProps): JSX.Element | null {
  if (error !== null && error !== undefined) {
    const upstream = isApiError(error, ErrorCode.Upstream) ? upstreamData(error.data) : null;
    return (
      <Alert
        color="red"
        title={upstream ? `CYBERBIZ returned ${upstream.status}` : 'Request failed'}
      >
        <Stack gap={4}>
          <Text size="sm">{errorMessage(error)}</Text>
          {upstream?.messages.map((m) => (
            <Text size="sm" key={m}>
              • {m}
            </Text>
          ))}
          {upstream?.request_id && (
            <Text size="xs" c="dimmed" ff="monospace">
              request id {upstream.request_id}
            </Text>
          )}
        </Stack>
      </Alert>
    );
  }
  if (result === null) {
    return null;
  }

  const contentType = headerValue(result.response_headers, 'Content-Type');
  const binary = isBinaryContentType(contentType);

  return (
    <Paper withBorder p="md">
      <Stack>
        <Group justify="space-between" wrap="wrap">
          <Group gap="md">
            <Title order={5}>Result</Title>
            <StatusBadge status={result.response_status} />
            <Text size="sm">{formatDuration(result.duration_ms)}</Text>
            {result.attempt > 1 && (
              <Text size="sm" c="dimmed">
                after {result.attempt} attempts
              </Text>
            )}
            {result.request_id !== '' && (
              <Text size="sm" c="dimmed" ff="monospace">
                {result.request_id}
              </Text>
            )}
          </Group>
          <GoldenFileButton kind="outbound" id={result.id} size="xs" />
        </Group>
        {result.error !== null && result.error !== '' && (
          <Alert color="red" title="Transport error">
            {result.error}
          </Alert>
        )}
        <Tabs defaultValue="body">
          <Tabs.List>
            <Tabs.Tab value="body">Body</Tabs.Tab>
            <Tabs.Tab value="headers">Headers</Tabs.Tab>
            <Tabs.Tab value="request">Request</Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="body" pt="sm">
            {binary ? (
              <Alert color="blue" variant="light" title="Binary response">
                <Text size="sm">
                  {contentType}.{' '}
                  <Anchor href={`/api/outbound/${result.id}/body`} download>
                    <IconDownload size={14} style={{ verticalAlign: 'middle' }} /> Download
                  </Anchor>
                </Text>
              </Alert>
            ) : (
              <JsonView value={result.response_body} title="Response body" maxHeight={600} />
            )}
          </Tabs.Panel>
          <Tabs.Panel value="headers" pt="sm">
            <HeadersTable headers={result.response_headers} />
          </Tabs.Panel>
          <Tabs.Panel value="request" pt="sm">
            <Stack>
              <Text size="sm" ff="monospace">
                {result.method} {result.path}
                {result.query !== '' ? `?${result.query}` : ''}
              </Text>
              <HeadersTable headers={result.request_headers} />
              <JsonView value={result.request_body} title="Request body" />
            </Stack>
          </Tabs.Panel>
        </Tabs>
      </Stack>
    </Paper>
  );
}
