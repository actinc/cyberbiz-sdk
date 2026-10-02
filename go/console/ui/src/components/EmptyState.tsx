import { Center, Stack, Text } from '@mantine/core';
import { IconInbox } from '@tabler/icons-react';
import type { JSX, ReactNode } from 'react';

/** Props for {@link EmptyState}. */
export interface EmptyStateProps {
  /** Headline. */
  title: string;
  /** Supporting text. */
  description?: string;
  /** Optional call to action rendered under the text. */
  action?: ReactNode;
}

/** Placeholder for empty lists. */
export function EmptyState({ title, description, action }: EmptyStateProps): JSX.Element {
  return (
    <Center py="xl">
      <Stack align="center" gap="xs">
        <IconInbox size={40} stroke={1.2} color="var(--mantine-color-dimmed)" />
        <Text fw={500}>{title}</Text>
        {description && (
          <Text size="sm" c="dimmed" ta="center">
            {description}
          </Text>
        )}
        {action}
      </Stack>
    </Center>
  );
}
