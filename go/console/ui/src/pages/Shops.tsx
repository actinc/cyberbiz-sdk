import { ActionIcon, Button, Code, Group, Modal, Stack, Text, Title, Tooltip } from '@mantine/core';
import { useDisclosure } from '@mantine/hooks';
import { IconCheck, IconEdit, IconPlus, IconTrash } from '@tabler/icons-react';
import { useState, type JSX } from 'react';
import { errorMessage } from '../api/client';
import { useCreateShop, useDeleteShop, useShops, useUpdateShop, useVerifyShop } from '../api/hooks';
import type { CreateShopRequest, JsonValue, Shop } from '../api/types';
import { ConfirmModal } from '../components/ConfirmModal';
import { DateTime } from '../components/DateTime';
import { EmptyState } from '../components/EmptyState';
import { JsonView } from '../components/JsonView';
import { ResponsiveTable, type Column } from '../components/ResponsiveTable';
import { ShopForm } from '../components/ShopForm';
import { notifyError, notifySuccess } from '../hooks/useNotify';

/** State of the create/edit modal. */
type Editing = { mode: 'create' } | { mode: 'edit'; shop: Shop } | null;

/** Result shown in the verify modal. */
interface VerifyResult {
  shop: Shop;
  info: JsonValue;
}

/** Shops page: list, add, edit, delete, verify. Only the token fingerprint is ever shown. */
export function ShopsPage(): JSX.Element {
  const shops = useShops();
  const create = useCreateShop();
  const update = useUpdateShop();
  const remove = useDeleteShop();
  const verify = useVerifyShop();

  const [editing, setEditing] = useState<Editing>(null);
  const [deleting, setDeleting] = useState<Shop | null>(null);
  const [verified, setVerified] = useState<VerifyResult | null>(null);
  const [verifyOpened, verifyModal] = useDisclosure(false);

  const handleSubmit = (values: CreateShopRequest): void => {
    if (editing?.mode === 'edit') {
      const body: Partial<CreateShopRequest> = { ...values };
      if (body.app_secret === '') {
        delete body.app_secret;
      }
      if (body.api_token === '') {
        delete body.api_token;
      }
      update.mutate(
        { id: editing.shop.id, body },
        {
          onSuccess: (shop) => {
            notifySuccess(`${shop.name} updated.`);
            setEditing(null);
          },
          onError: (err) => notifyError(err, 'Update failed'),
        },
      );
      return;
    }
    create.mutate(values, {
      onSuccess: (shop) => {
        notifySuccess(`${shop.name} verified and added.`);
        setEditing(null);
      },
      onError: (err) => notifyError(err, 'Verification failed'),
    });
  };

  const handleVerify = (shop: Shop): void => {
    verify.mutate(shop.id, {
      onSuccess: (res) => {
        setVerified({ shop, info: res.shop_info });
        verifyModal.open();
      },
      onError: (err) => notifyError(err, `Verify ${shop.name} failed`),
    });
  };

  const columns: Column<Shop>[] = [
    {
      key: 'name',
      header: 'Name',
      render: (s) => (
        <Stack gap={0}>
          <Text fw={500}>{s.name}</Text>
          <Text size="xs" c="dimmed">
            CYBERBIZ id {s.cyberbiz_shop_id}
          </Text>
        </Stack>
      ),
    },
    {
      key: 'domain',
      header: 'Shop Domain',
      render: (s) => (
        <Stack gap={0}>
          <Text size="sm" ff="monospace">
            {s.shop_domain}
          </Text>
          {s.custom_domain !== '' && (
            <Text size="xs" c="dimmed" ff="monospace">
              {s.custom_domain}
            </Text>
          )}
        </Stack>
      ),
    },
    {
      key: 'app',
      header: 'App',
      render: (s) => (
        <Stack gap={0}>
          <Text size="sm">{s.app_name}</Text>
          <Text size="xs" c="dimmed" ff="monospace">
            {s.app_id}
          </Text>
        </Stack>
      ),
    },
    {
      key: 'token',
      header: 'API Token',
      render: (s) => <Code>{s.token_fingerprint}</Code>,
    },
    {
      key: 'secret',
      header: 'App Secret',
      render: (s) => (
        <Text size="sm" c={s.has_app_secret ? undefined : 'orange'}>
          {s.has_app_secret ? 'stored' : 'missing'}
        </Text>
      ),
    },
    { key: 'updated', header: 'Updated', render: (s) => <DateTime value={s.updated_at} /> },
    {
      key: 'actions',
      header: '',
      hideLabelOnMobile: true,
      render: (s) => (
        <Group gap={4} wrap="nowrap" justify="flex-end">
          <Tooltip label="Verify (GET /shop)">
            <ActionIcon
              variant="subtle"
              color="green"
              onClick={() => handleVerify(s)}
              loading={verify.isPending && verify.variables === s.id}
              aria-label="Verify"
            >
              <IconCheck size={16} />
            </ActionIcon>
          </Tooltip>
          <Tooltip label="Edit">
            <ActionIcon
              variant="subtle"
              onClick={() => setEditing({ mode: 'edit', shop: s })}
              aria-label="Edit"
            >
              <IconEdit size={16} />
            </ActionIcon>
          </Tooltip>
          <Tooltip label="Delete">
            <ActionIcon
              variant="subtle"
              color="red"
              onClick={() => setDeleting(s)}
              aria-label="Delete"
            >
              <IconTrash size={16} />
            </ActionIcon>
          </Tooltip>
        </Group>
      ),
    },
  ];

  return (
    <Stack>
      <Group justify="space-between">
        <Title order={2}>Shops</Title>
        <Button leftSection={<IconPlus size={16} />} onClick={() => setEditing({ mode: 'create' })}>
          Add Shop
        </Button>
      </Group>

      {shops.isError && <Text c="red">{errorMessage(shops.error)}</Text>}
      {shops.data?.length === 0 ? (
        <EmptyState
          title="No Shops yet"
          description="Add a Shop with its App credentials and API Token."
        />
      ) : (
        <ResponsiveTable
          columns={columns}
          rows={shops.data ?? []}
          rowKey={(s) => s.id}
          minWidth={900}
        />
      )}

      <Modal
        opened={editing !== null}
        onClose={() => setEditing(null)}
        title={editing?.mode === 'edit' ? `Edit ${editing.shop.name}` : 'Add Shop'}
        size="lg"
      >
        {editing !== null && (
          <ShopForm
            shop={editing.mode === 'edit' ? editing.shop : undefined}
            submitLabel={editing.mode === 'edit' ? 'Save' : 'Verify and save'}
            loading={create.isPending || update.isPending}
            onSubmit={handleSubmit}
            onCancel={() => setEditing(null)}
          />
        )}
      </Modal>

      <ConfirmModal
        opened={deleting !== null}
        title="Delete Shop"
        message={
          deleting
            ? `Delete ${deleting.name} (${deleting.shop_domain})? Its Outbound and Inbound logs are kept and will show "deleted shop".`
            : ''
        }
        loading={remove.isPending}
        onClose={() => setDeleting(null)}
        onConfirm={() => {
          if (!deleting) {
            return;
          }
          remove.mutate(deleting.id, {
            onSuccess: () => {
              notifySuccess(`${deleting.name} deleted.`);
              setDeleting(null);
            },
            onError: (err) => notifyError(err, 'Delete failed'),
          });
        }}
      />

      <Modal
        opened={verifyOpened}
        onClose={verifyModal.close}
        title={verified ? `${verified.shop.name}: GET /shop` : 'Verify'}
        size="lg"
      >
        {verified && <JsonView value={verified.info} title="shop_info" />}
      </Modal>
    </Stack>
  );
}
