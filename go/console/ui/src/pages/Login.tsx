import { Button, Center, Paper, PasswordInput, Stack, Text, TextInput, Title } from '@mantine/core';
import { useForm } from '@mantine/form';
import type { JSX } from 'react';
import { useLocation, useNavigate } from 'react-router';
import { errorMessage } from '../api/client';
import { useLogin } from '../api/hooks';
import type { LoginRequest } from '../api/types';

/** Router state passed by the auth guard so login can return to the original page. */
interface LocationState {
  from?: string;
}

/** Username/password login page. */
export function LoginPage(): JSX.Element {
  const navigate = useNavigate();
  const location = useLocation();
  const login = useLogin();
  const form = useForm<LoginRequest>({
    initialValues: { username: '', password: '' },
    validate: {
      username: (v) => (v.trim() ? null : 'Username is required'),
      password: (v) => (v ? null : 'Password is required'),
    },
  });

  const from = (location.state as LocationState | null)?.from ?? '/';

  return (
    <Center mih="100vh" p="md">
      <Paper withBorder shadow="sm" p="xl" w="100%" maw={380}>
        <form
          onSubmit={form.onSubmit((values) => {
            login.mutate(values, {
              onSuccess: () => void navigate(from === '/login' ? '/' : from, { replace: true }),
            });
          })}
        >
          <Stack>
            <div>
              <Title order={3}>CYBERBIZ Console</Title>
              <Text size="sm" c="dimmed">
                Sign in to continue.
              </Text>
            </div>
            <TextInput
              label="Username"
              autoComplete="username"
              autoFocus
              {...form.getInputProps('username')}
            />
            <PasswordInput
              label="Password"
              autoComplete="current-password"
              {...form.getInputProps('password')}
            />
            {login.isError && (
              <Text size="sm" c="red" role="alert">
                {errorMessage(login.error)}
              </Text>
            )}
            <Button type="submit" loading={login.isPending} fullWidth>
              Sign in
            </Button>
          </Stack>
        </form>
      </Paper>
    </Center>
  );
}
