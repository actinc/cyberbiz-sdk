import { Button, Group, Modal, Stack, Text, TextInput } from '@mantine/core';
import { useDisclosure } from '@mantine/hooks';
import { IconFileCheck } from '@tabler/icons-react';
import { useState, type JSX } from 'react';
import { isApiError } from '../api/client';
import { useSaveGolden, type GoldenKind } from '../api/hooks';
import { ErrorCode } from '../api/types';
import { notifyError, notifySuccess } from '../hooks/useNotify';

/** Props for {@link GoldenFileButton}. */
export interface GoldenFileButtonProps {
  /** Which log kind the Golden File is written from. */
  kind: GoldenKind;
  /** Log id. */
  id: number;
  /** Button size. */
  size?: 'xs' | 'sm' | 'md';
  /** Button variant. */
  variant?: 'filled' | 'light' | 'outline' | 'subtle';
}

/**
 * "Save as Golden File" button. Opens a modal asking for an optional name suffix;
 * when the backend answers 40900 (file exists) it offers to overwrite.
 */
export function GoldenFileButton({
  kind,
  id,
  size = 'sm',
  variant = 'light',
}: GoldenFileButtonProps): JSX.Element {
  const [opened, modal] = useDisclosure(false);
  const [name, setName] = useState('');
  const [conflict, setConflict] = useState<string | null>(null);
  const save = useSaveGolden(kind);

  const submit = (overwrite: boolean): void => {
    save.mutate(
      { id, body: { name: name.trim() || undefined, overwrite: overwrite || undefined } },
      {
        onSuccess: (res) => {
          notifySuccess(res.path, 'Golden File written');
          setConflict(null);
          modal.close();
        },
        onError: (err) => {
          if (isApiError(err, ErrorCode.Conflict)) {
            setConflict(err.message);
            return;
          }
          notifyError(err, 'Could not write Golden File');
        },
      },
    );
  };

  return (
    <>
      <Button
        size={size}
        variant={variant}
        leftSection={<IconFileCheck size={16} />}
        onClick={() => {
          setConflict(null);
          modal.open();
        }}
      >
        Save as Golden File
      </Button>
      <Modal opened={opened} onClose={modal.close} title="Save as Golden File">
        <Stack>
          <Text size="sm" c="dimmed">
            The redacted {kind === 'outbound' ? 'response body' : 'Inbound body'} and headers are
            written to the Golden File directory. Sensitive values are redacted by the backend.
          </Text>
          <TextInput
            label="Name suffix (optional)"
            description="Appended to the generated file name, e.g. `_paid` or `_empty`."
            value={name}
            onChange={(e) => setName(e.currentTarget.value)}
            data-autofocus
          />
          {conflict !== null && (
            <Text size="sm" c="orange">
              {conflict} — overwrite it?
            </Text>
          )}
          <Group justify="flex-end">
            <Button variant="default" onClick={modal.close}>
              Cancel
            </Button>
            {conflict === null ? (
              <Button onClick={() => submit(false)} loading={save.isPending}>
                Save
              </Button>
            ) : (
              <Button color="orange" onClick={() => submit(true)} loading={save.isPending}>
                Overwrite
              </Button>
            )}
          </Group>
        </Stack>
      </Modal>
    </>
  );
}
