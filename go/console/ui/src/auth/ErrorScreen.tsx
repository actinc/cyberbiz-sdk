import { Alert, Button, Center, Stack } from '@mantine/core';
import { IconAlertTriangle } from '@tabler/icons-react';
import type { JSX } from 'react';
import { errorMessage } from '../api/client';

/** Props for {@link ErrorScreen}. */
export interface ErrorScreenProps {
  /** The failure to show. */
  error: unknown;
  /** Retry handler. */
  onRetry: () => void;
}

/** Full-page error with a retry button, used when the backend is unreachable. */
export function ErrorScreen({ error, onRetry }: ErrorScreenProps): JSX.Element {
  return (
    <Center h="100vh" p="md">
      <Stack maw={480}>
        <Alert
          color="red"
          icon={<IconAlertTriangle size={18} />}
          title="Console backend unreachable"
        >
          {errorMessage(error)}
        </Alert>
        <Button onClick={onRetry} variant="light">
          Retry
        </Button>
      </Stack>
    </Center>
  );
}
