import { Text } from '@mantine/core';
import type { JSX } from 'react';
import { useShops } from '../api/hooks';

/** Props for {@link ShopName}. */
export interface ShopNameProps {
  /** Shop id; null when the Inbound matched no Shop. */
  shopId: number | null;
  /** Shown instead of "deleted shop" when the id is null. */
  fallback?: string;
}

/** Shop name by id, or "deleted shop" when the id no longer resolves. */
export function ShopName({ shopId, fallback = '—' }: ShopNameProps): JSX.Element {
  const shops = useShops();
  if (shopId === null) {
    return (
      <Text size="sm" c="dimmed">
        {fallback}
      </Text>
    );
  }
  const shop = shops.data?.find((s) => s.id === shopId);
  if (!shop) {
    return (
      <Text size="sm" c="dimmed" fs="italic">
        deleted shop
      </Text>
    );
  }
  return <Text size="sm">{shop.name}</Text>;
}
