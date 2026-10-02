import { Button, Group, Modal, Stack, Text } from '@mantine/core';
import type { JSX } from 'react';

/** Props for {@link ConfirmModal}. */
export interface ConfirmModalProps {
  /** Whether the modal is shown. */
  opened: boolean;
  /** Title. */
  title: string;
  /** Body text. */
  message: string;
  /** Confirm button label. */
  confirmLabel?: string;
  /** Whether the confirm action is in flight. */
  loading?: boolean;
  /** Confirm handler. */
  onConfirm: () => void;
  /** Cancel/close handler. */
  onClose: () => void;
}

/** Simple destructive-action confirmation. */
export function ConfirmModal({
  opened,
  title,
  message,
  confirmLabel = 'Delete',
  loading = false,
  onConfirm,
  onClose,
}: ConfirmModalProps): JSX.Element {
  return (
    <Modal opened={opened} onClose={onClose} title={title}>
      <Stack>
        <Text size="sm">{message}</Text>
        <Group justify="flex-end">
          <Button variant="default" onClick={onClose}>
            Cancel
          </Button>
          <Button color="red" onClick={onConfirm} loading={loading}>
            {confirmLabel}
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
