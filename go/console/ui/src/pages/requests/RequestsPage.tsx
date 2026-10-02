import { Alert, Center, Grid, Loader, Paper, Stack, Text, Title } from '@mantine/core';
import { useEffect, useState, type JSX } from 'react';
import { errorMessage } from '../../api/client';
import { useCatalog, useSendRequest, useShops } from '../../api/hooks';
import type { Operation, OutboundLog } from '../../api/types';
import { EmptyState } from '../../components/EmptyState';
import { OperationForm } from './OperationForm';
import { OperationList } from './OperationList';
import { ResultPanel } from './ResultPanel';

/** The API tester: operation list on the left, form and result on the right. */
export function RequestsPage(): JSX.Element {
  const catalog = useCatalog();
  const shops = useShops();
  const send = useSendRequest();
  const [operation, setOperation] = useState<Operation | null>(null);
  const [shopId, setShopId] = useState<number | null>(null);
  const [result, setResult] = useState<OutboundLog | null>(null);

  useEffect(() => {
    if (shopId === null && shops.data && shops.data.length > 0) {
      setShopId(shops.data[0]?.id ?? null);
    }
  }, [shops.data, shopId]);

  if (catalog.isPending) {
    return (
      <Center py="xl">
        <Loader />
      </Center>
    );
  }
  if (catalog.isError) {
    return (
      <Alert color="red" title="Could not load the API catalog">
        {errorMessage(catalog.error)}
      </Alert>
    );
  }

  return (
    <Stack>
      <Title order={2}>API tester</Title>
      <Grid gutter="md">
        <Grid.Col span={{ base: 12, md: 4, lg: 3 }}>
          <Paper withBorder p="sm">
            <OperationList
              catalog={catalog.data}
              selectedId={operation?.id ?? null}
              onSelect={(op) => {
                setOperation(op);
                setResult(null);
                send.reset();
              }}
            />
          </Paper>
        </Grid.Col>
        <Grid.Col span={{ base: 12, md: 8, lg: 9 }}>
          <Stack>
            <Paper withBorder p="md">
              {operation === null ? (
                <EmptyState
                  title="Pick an operation"
                  description={`${catalog.data.operations.length} operations across ${catalog.data.tags.length} tags.`}
                />
              ) : (
                <OperationForm
                  operation={operation}
                  shopId={shopId}
                  onShopChange={setShopId}
                  sending={send.isPending}
                  onSend={(request) => {
                    if (shopId === null) {
                      return;
                    }
                    setResult(null);
                    send.mutate({ shopId, body: request }, { onSuccess: (log) => setResult(log) });
                  }}
                />
              )}
            </Paper>
            {shops.data?.length === 0 && (
              <Text size="sm" c="orange">
                No Shops configured; add one before sending requests.
              </Text>
            )}
            <ResultPanel result={result} error={send.error} />
          </Stack>
        </Grid.Col>
      </Grid>
    </Stack>
  );
}
