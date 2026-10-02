import { Box, Code, Group, ScrollArea, Text } from '@mantine/core';
import type { JSX } from 'react';
import { prettyJson } from '../utils/json';
import { CopyIconButton } from './CopyIconButton';

/** Props for {@link JsonView}. */
export interface JsonViewProps {
  /** Value to render: an object, or a string that may contain JSON. */
  value: unknown;
  /** Optional caption shown above the block. */
  title?: string;
  /** Maximum height before the block scrolls. */
  maxHeight?: number;
}

/** Pretty-printed JSON block with a copy button. */
export function JsonView({ value, title, maxHeight = 480 }: JsonViewProps): JSX.Element {
  const text = value === undefined || value === '' ? '' : prettyJson(value);
  return (
    <Box>
      <Group justify="space-between" mb={4}>
        <Text size="xs" c="dimmed">
          {title ?? ''}
        </Text>
        {text !== '' && <CopyIconButton value={text} label="Copy JSON" />}
      </Group>
      <ScrollArea.Autosize mah={maxHeight} type="auto">
        <Code block style={{ whiteSpace: 'pre', fontSize: 12 }}>
          {text === '' ? '(empty)' : text}
        </Code>
      </ScrollArea.Autosize>
    </Box>
  );
}
