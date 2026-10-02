import { useMediaQuery } from '@mantine/hooks';

/**
 * True when the viewport is narrower than Mantine's `sm` breakpoint (48em).
 * Pages switch from tables to cards below it.
 */
export function useIsMobile(): boolean {
  return useMediaQuery('(max-width: 48em)');
}
