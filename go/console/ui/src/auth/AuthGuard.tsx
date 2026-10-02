import { Center, Loader } from '@mantine/core';
import type { JSX } from 'react';
import { Navigate, Outlet, useLocation } from 'react-router';
import { isApiError } from '../api/client';
import { useMe } from '../api/hooks';
import { ErrorCode } from '../api/types';
import { ErrorScreen } from './ErrorScreen';

/**
 * Probes `/api/auth/me` once. Unauthenticated users go to /login; while
 * `setup_required` is true every route except /setup redirects there, and
 * /setup redirects back to / once a Shop exists.
 */
export function AuthGuard(): JSX.Element {
  const me = useMe();
  const location = useLocation();

  if (me.isPending) {
    return (
      <Center h="100vh">
        <Loader />
      </Center>
    );
  }

  if (me.isError) {
    if (isApiError(me.error, ErrorCode.Unauthenticated)) {
      return <Navigate to="/login" replace state={{ from: location.pathname }} />;
    }
    return <ErrorScreen error={me.error} onRetry={() => void me.refetch()} />;
  }

  const onSetup = location.pathname === '/setup';
  if (me.data.setup_required && !onSetup) {
    return <Navigate to="/setup" replace />;
  }
  if (!me.data.setup_required && onSetup) {
    return <Navigate to="/" replace />;
  }

  return <Outlet />;
}
