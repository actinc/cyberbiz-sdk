import type { JSX } from 'react';
import { Navigate, Route, Routes } from 'react-router';
import { AuthGuard } from './auth/AuthGuard';
import { AppLayout } from './layout/AppLayout';
import { DashboardPage } from './pages/Dashboard';
import { InboundPage } from './pages/Inbound';
import { LoginPage } from './pages/Login';
import { OutboundPage } from './pages/Outbound';
import { RequestsPage } from './pages/requests/RequestsPage';
import { SetupPage } from './pages/Setup';
import { ShopsPage } from './pages/Shops';

/** Route table. Everything except /login sits behind the auth guard. */
export function App(): JSX.Element {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route element={<AuthGuard />}>
        <Route path="/setup" element={<SetupPage />} />
        <Route element={<AppLayout />}>
          <Route index element={<DashboardPage />} />
          <Route path="/shops" element={<ShopsPage />} />
          <Route path="/requests" element={<RequestsPage />} />
          <Route path="/outbound" element={<OutboundPage />} />
          <Route path="/inbound" element={<InboundPage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
