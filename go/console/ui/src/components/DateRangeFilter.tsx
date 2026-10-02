import { DatePickerInput } from '@mantine/dates';
import dayjs from 'dayjs';
import type { JSX } from 'react';

/** ISO `from`/`to` pair used by the list filters. */
export interface IsoRange {
  from?: string;
  to?: string;
}

/** Props for {@link DateRangeFilter}. */
export interface DateRangeFilterProps {
  /** Current range as ISO strings. */
  value: IsoRange;
  /** Change handler; emits start-of-day `from` and end-of-day `to` in ISO. */
  onChange: (range: IsoRange) => void;
}

/** Date range picker mapped to the backend's ISO `from`/`to` query parameters. */
export function DateRangeFilter({ value, onChange }: DateRangeFilterProps): JSX.Element {
  const picked: [string | null, string | null] = [
    value.from ? dayjs(value.from).format('YYYY-MM-DD') : null,
    value.to ? dayjs(value.to).format('YYYY-MM-DD') : null,
  ];
  return (
    <DatePickerInput
      type="range"
      placeholder="Date range"
      value={picked}
      onChange={([from, to]) => {
        onChange({
          from: from ? dayjs(from).startOf('day').toISOString() : undefined,
          to: to ? dayjs(to).endOf('day').toISOString() : undefined,
        });
      }}
      clearable
      allowSingleDateInRange
      size="sm"
      w={{ base: '100%', sm: 240 }}
    />
  );
}
