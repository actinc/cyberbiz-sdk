import { Drawer, Loader, Center, Text } from '@mantine/core';
import type { JSX, ReactNode } from 'react';
import { errorMessage } from '../api/client';
import { useIsMobile } from '../hooks/useIsMobile';

/** Props for {@link LogDrawer}. */
export interface LogDrawerProps {
  /** Whether the drawer is open. */
  opened: boolean;
  /** Close handler. */
  onClose: () => void;
  /** Drawer title. */
  title: ReactNode;
  /** Whether the detail query is still loading. */
  loading: boolean;
  /** Detail query error, if any. */
  error: unknown;
  /** Content once loaded. */
  children: ReactNode;
}

/** Right-hand drawer (full width on mobile) with loading and error states. */
export function LogDrawer({
  opened,
  onClose,
  title,
  loading,
  error,
  children,
}: LogDrawerProps): JSX.Element {
  const isMobile = useIsMobile();
  return (
    <Drawer
      opened={opened}
      onClose={onClose}
      title={title}
      position="right"
      size={isMobile ? '100%' : 'xl'}
      padding="md"
    >
      {loading && (
        <Center py="xl">
          <Loader />
        </Center>
      )}
      {error !== null && error !== undefined && <Text c="red">{errorMessage(error)}</Text>}
      {!loading && children}
    </Drawer>
  );
}
