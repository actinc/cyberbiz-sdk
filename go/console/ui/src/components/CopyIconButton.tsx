import { ActionIcon, CopyButton, Tooltip } from '@mantine/core';
import { IconCheck, IconCopy } from '@tabler/icons-react';
import type { JSX } from 'react';

/** Props for {@link CopyIconButton}. */
export interface CopyIconButtonProps {
  /** Text placed on the clipboard. */
  value: string;
  /** Tooltip label; defaults to "Copy". */
  label?: string;
  /** Icon size in pixels. */
  size?: number;
}

/** Icon button that copies `value` and shows a check mark briefly afterwards. */
export function CopyIconButton({
  value,
  label = 'Copy',
  size = 16,
}: CopyIconButtonProps): JSX.Element {
  return (
    <CopyButton value={value} timeout={1500}>
      {({ copied, copy }) => (
        <Tooltip label={copied ? 'Copied' : label} withArrow>
          <ActionIcon
            variant="subtle"
            color={copied ? 'teal' : 'gray'}
            onClick={copy}
            aria-label={label}
          >
            {copied ? <IconCheck size={size} /> : <IconCopy size={size} />}
          </ActionIcon>
        </Tooltip>
      )}
    </CopyButton>
  );
}
