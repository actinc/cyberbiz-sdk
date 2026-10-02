import { Alert, Code, Group, Paper, SimpleGrid, Skeleton, Stack, Text, Title } from '@mantine/core';
import { IconInfoCircle } from '@tabler/icons-react';
import type { JSX } from 'react';
import { errorMessage } from '../api/client';
import { useShops, useStats } from '../api/hooks';
import { CopyIconButton } from '../components/CopyIconButton';

/** Props for {@link StatCard}. */
interface StatCardProps {
  label: string;
  value: number | undefined;
  color?: string;
}

/** One dashboard counter. */
function StatCard({ label, value, color }: StatCardProps): JSX.Element {
  return (
    <Paper withBorder p="md">
      <Text size="xs" c="dimmed" tt="uppercase" fw={600}>
        {label}
      </Text>
      {value === undefined ? (
        <Skeleton h={32} mt={4} w={80} />
      ) : (
        <Text fz={28} fw={700} c={color}>
          {value}
        </Text>
      )}
    </Paper>
  );
}

/** Dashboard: stats cards and the webhook URL to paste into CYBERBIZ. */
export function DashboardPage(): JSX.Element {
  const stats = useStats();
  const shops = useShops();
  const webhookUrl = shops.data?.find((s) => s.webhook_url !== '')?.webhook_url ?? '';

  return (
    <Stack>
      <Title order={2}>Dashboard</Title>
      {stats.isError && (
        <Alert color="red" title="Could not load stats">
          {errorMessage(stats.error)}
        </Alert>
      )}
      <SimpleGrid cols={{ base: 2, sm: 4 }}>
        <StatCard label="Shops" value={stats.data?.shops} />
        <StatCard label="Outbound (24h)" value={stats.data?.outbound_24h} />
        <StatCard label="Inbound (24h)" value={stats.data?.inbound_24h} />
        <StatCard
          label="Invalid Inbound (24h)"
          value={stats.data?.inbound_invalid_24h}
          color={stats.data && stats.data.inbound_invalid_24h > 0 ? 'red' : undefined}
        />
      </SimpleGrid>

      <Paper withBorder p="md">
        <Stack gap="sm">
          <Title order={4}>Webhook URL</Title>
          <Text size="sm" c="dimmed">
            Paste this URL into the CYBERBIZ app settings as the webhook endpoint. Every Inbound is
            verified with the Shop's App Secret and recorded, whatever the result.
          </Text>
          {webhookUrl === '' ? (
            <Alert color="yellow" variant="light" icon={<IconInfoCircle size={16} />}>
              <Code>CONSOLE_PUBLIC_URL</Code> is not set, so no public webhook URL is available. The
              endpoint is <Code>/webhooks/cyberbiz</Code> on this server.
            </Alert>
          ) : (
            <Group gap="xs" wrap="nowrap">
              <Code style={{ wordBreak: 'break-all' }}>{webhookUrl}</Code>
              <CopyIconButton value={webhookUrl} label="Copy webhook URL" />
            </Group>
          )}
          <Alert color="blue" variant="light" icon={<IconInfoCircle size={16} />}>
            CYBERBIZ must reach this Console over the public internet. When running locally, expose
            it with a tunnel such as <Code>ngrok http 8787</Code> or{' '}
            <Code>cloudflared tunnel --url http://localhost:8787</Code> and set{' '}
            <Code>CONSOLE_PUBLIC_URL</Code> to the tunnel URL.
          </Alert>
        </Stack>
      </Paper>
    </Stack>
  );
}
