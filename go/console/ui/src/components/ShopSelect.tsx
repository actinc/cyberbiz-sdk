import { Select, type SelectProps } from '@mantine/core';
import type { JSX } from 'react';
import { useShops } from '../api/hooks';

/** Props for {@link ShopSelect}. */
export interface ShopSelectProps extends Omit<SelectProps, 'data' | 'value' | 'onChange'> {
  /** Selected Shop id, null for none. */
  value: number | null;
  /** Change handler. */
  onChange: (id: number | null) => void;
}

/** Select listing every Shop by name and Shop Domain. */
export function ShopSelect({ value, onChange, ...rest }: ShopSelectProps): JSX.Element {
  const shops = useShops();
  const data = (shops.data ?? []).map((s) => ({
    value: String(s.id),
    label: `${s.name} (${s.shop_domain})`,
  }));
  return (
    <Select
      data={data}
      value={value === null ? null : String(value)}
      onChange={(v) => onChange(v === null ? null : Number(v))}
      placeholder={shops.isLoading ? 'Loading shops…' : 'Select a Shop'}
      searchable
      nothingFoundMessage="No Shops"
      {...rest}
    />
  );
}
