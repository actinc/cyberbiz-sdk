import {
  ActionIcon,
  AppShell,
  Badge,
  Burger,
  Group,
  NavLink,
  Text,
  Tooltip,
  useMantineColorScheme,
} from '@mantine/core';
import { useDisclosure } from '@mantine/hooks';
import {
  IconArrowDownLeft,
  IconArrowUpRight,
  IconBuildingStore,
  IconLayoutDashboard,
  IconLogout,
  IconMoon,
  IconSend,
  IconSun,
} from '@tabler/icons-react';
import type { JSX } from 'react';
import { Link, Outlet, useLocation, useNavigate } from 'react-router';
import { useLogout, useMe, useShops } from '../api/hooks';
import { notifyError } from '../hooks/useNotify';

const navItems = [
  { path: '/', label: 'Dashboard', icon: IconLayoutDashboard },
  { path: '/requests', label: 'API tester', icon: IconSend },
  { path: '/outbound', label: 'Outbound', icon: IconArrowUpRight },
  { path: '/inbound', label: 'Inbound', icon: IconArrowDownLeft },
  { path: '/shops', label: 'Shops', icon: IconBuildingStore },
] as const;

/** Mantine AppShell with a collapsible navbar, shop count, theme toggle and logout. */
export function AppLayout(): JSX.Element {
  const [opened, { toggle, close }] = useDisclosure();
  const location = useLocation();
  const navigate = useNavigate();
  const me = useMe();
  const shops = useShops();
  const logout = useLogout();
  const { colorScheme, toggleColorScheme } = useMantineColorScheme();

  const handleLogout = (): void => {
    logout.mutate(undefined, {
      onSuccess: () => void navigate('/login', { replace: true }),
      onError: (err) => notifyError(err, 'Logout failed'),
    });
  };

  return (
    <AppShell
      header={{ height: 56 }}
      navbar={{ width: 220, breakpoint: 'sm', collapsed: { mobile: !opened } }}
      padding={{ base: 'sm', sm: 'md' }}
    >
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between" wrap="nowrap">
          <Group gap="sm" wrap="nowrap">
            <Burger opened={opened} onClick={toggle} hiddenFrom="sm" size="sm" />
            <Text fw={700} size="lg" component={Link} to="/" c="inherit" td="none">
              CYBERBIZ Console
            </Text>
          </Group>
          <Group gap="xs" wrap="nowrap">
            <Tooltip label="Shops configured">
              <Badge
                variant="light"
                component={Link}
                to="/shops"
                leftSection={<IconBuildingStore size={12} />}
                style={{ cursor: 'pointer' }}
              >
                {shops.data?.length ?? '…'} {shops.data?.length === 1 ? 'Shop' : 'Shops'}
              </Badge>
            </Tooltip>
            <Text size="sm" c="dimmed" visibleFrom="xs">
              {me.data?.user.username}
            </Text>
            <Tooltip label={colorScheme === 'dark' ? 'Light theme' : 'Dark theme'}>
              <ActionIcon
                variant="subtle"
                color="gray"
                onClick={toggleColorScheme}
                aria-label="Toggle theme"
              >
                {colorScheme === 'dark' ? <IconSun size={18} /> : <IconMoon size={18} />}
              </ActionIcon>
            </Tooltip>
            <Tooltip label="Log out">
              <ActionIcon
                variant="subtle"
                color="gray"
                onClick={handleLogout}
                loading={logout.isPending}
                aria-label="Log out"
              >
                <IconLogout size={18} />
              </ActionIcon>
            </Tooltip>
          </Group>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="sm">
        {navItems.map((item) => (
          <NavLink
            key={item.path}
            component={Link}
            to={item.path}
            label={item.label}
            leftSection={<item.icon size={16} stroke={1.5} />}
            active={location.pathname === item.path}
            onClick={close}
          />
        ))}
      </AppShell.Navbar>

      <AppShell.Main>
        <Outlet />
      </AppShell.Main>
    </AppShell>
  );
}
