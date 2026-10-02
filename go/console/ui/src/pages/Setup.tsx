import { Alert, Center, Paper, Stack, Text, Title } from '@mantine/core';
import { IconInfoCircle } from '@tabler/icons-react';
import type { JSX } from 'react';
import { useNavigate } from 'react-router';
import { useCreateShop } from '../api/hooks';
import { ShopForm } from '../components/ShopForm';
import { notifyError, notifySuccess } from '../hooks/useNotify';

/**
 * First-run page shown while no Shop exists. Submitting calls `POST /api/shops`,
 * which verifies the credentials against CYBERBIZ before storing them.
 */
export function SetupPage(): JSX.Element {
  const navigate = useNavigate();
  const create = useCreateShop();

  return (
    <Center mih="100vh" p="md">
      <Paper withBorder shadow="sm" p="xl" w="100%" maw={560}>
        <Stack>
          <div>
            <Title order={3}>Add your first Shop</Title>
            <Text size="sm" c="dimmed">
              The Console needs one Shop with its App credentials and API Token before it can send
              Outbound requests or verify Inbound webhooks.
            </Text>
          </div>
          <Alert color="blue" variant="light" icon={<IconInfoCircle size={16} />}>
            "Verify" calls <code>GET /shop</code> through the SDK with these credentials. The Shop
            is only stored when CYBERBIZ accepts them; the call is recorded as an Outbound.
          </Alert>
          <ShopForm
            submitLabel="Verify and save"
            loading={create.isPending}
            onSubmit={(values) => {
              create.mutate(values, {
                onSuccess: (shop) => {
                  notifySuccess(
                    `${shop.name} (${shop.shop_domain}, CYBERBIZ id ${shop.cyberbiz_shop_id}) verified.`,
                    'Shop added',
                  );
                  void navigate('/', { replace: true });
                },
                onError: (err) => notifyError(err, 'Verification failed'),
              });
            }}
          />
        </Stack>
      </Paper>
    </Center>
  );
}
