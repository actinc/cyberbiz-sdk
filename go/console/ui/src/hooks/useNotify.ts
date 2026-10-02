import { notifications } from '@mantine/notifications';
import { errorMessage } from '../api/client';

/** Shows a green success toast. */
export function notifySuccess(message: string, title?: string): void {
  notifications.show({ color: 'green', title, message });
}

/** Shows a red error toast from any thrown value. */
export function notifyError(err: unknown, title = 'Request failed'): void {
  notifications.show({ color: 'red', title, message: errorMessage(err) });
}
