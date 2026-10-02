import { Button, Group, PasswordInput, Stack, TextInput } from '@mantine/core';
import { useForm } from '@mantine/form';
import type { JSX } from 'react';
import type { CreateShopRequest, Shop } from '../api/types';

/** Props for {@link ShopForm}. */
export interface ShopFormProps {
  /** Shop being edited; undefined for the create form. */
  shop?: Shop;
  /** Submit handler; on edit, blank secret/token mean "keep the stored one". */
  onSubmit: (values: CreateShopRequest) => void;
  /** Whether the submit is in flight. */
  loading?: boolean;
  /** Label of the submit button. */
  submitLabel?: string;
  /** Optional cancel handler; renders a Cancel button when set. */
  onCancel?: () => void;
}

const domainPattern = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?(\.[a-z0-9]([a-z0-9-]*[a-z0-9])?)+$/i;

/** Initial values for the Shop form, blank secrets on edit. */
function shopFormInitialValues(shop?: Shop): CreateShopRequest {
  return {
    name: shop?.name ?? '',
    shop_domain: shop?.shop_domain ?? '',
    app_name: shop?.app_name ?? '',
    app_id: shop?.app_id ?? '',
    app_secret: '',
    api_token: '',
  };
}

/** Shop create/edit form: name, Shop Domain, App Name, App ID, App Secret, API Token. */
export function ShopForm({
  shop,
  onSubmit,
  loading = false,
  submitLabel = 'Save',
  onCancel,
}: ShopFormProps): JSX.Element {
  const editing = shop !== undefined;
  const form = useForm<CreateShopRequest>({
    initialValues: shopFormInitialValues(shop),
    validate: {
      name: (v) => (v.trim() ? null : 'Name is required'),
      shop_domain: (v) =>
        domainPattern.test(v.trim()) ? null : 'Enter the Shop Domain, e.g. example.cyberbiz.co',
      app_name: (v) => (v.trim() ? null : 'App Name is required'),
      app_id: (v) => (v.trim() ? null : 'App ID is required'),
      app_secret: (v) => (editing || v.trim() ? null : 'App Secret is required'),
      api_token: (v) => (editing || v.trim() ? null : 'API Token is required'),
    },
  });

  return (
    <form
      onSubmit={form.onSubmit((values) => {
        onSubmit({
          ...values,
          name: values.name.trim(),
          shop_domain: values.shop_domain.trim(),
          app_name: values.app_name.trim(),
          app_id: values.app_id.trim(),
          app_secret: values.app_secret.trim(),
          api_token: values.api_token.trim(),
        });
      })}
    >
      <Stack>
        <TextInput label="Name" placeholder="My Shop" required {...form.getInputProps('name')} />
        <TextInput
          label="Shop Domain"
          placeholder="example.cyberbiz.co"
          description="Sent by CYBERBIZ as X-Cyberbiz-Domain on every Inbound."
          required
          {...form.getInputProps('shop_domain')}
        />
        <TextInput label="App Name" required {...form.getInputProps('app_name')} />
        <TextInput label="App ID" required {...form.getInputProps('app_id')} />
        <PasswordInput
          label="App Secret"
          description={
            editing
              ? shop.has_app_secret
                ? 'Leave blank to keep the stored App Secret.'
                : 'No App Secret stored; Inbound Signatures cannot be verified until one is set.'
              : 'Signs Inbound webhooks. Stored encrypted; never shown again.'
          }
          required={!editing}
          {...form.getInputProps('app_secret')}
        />
        <PasswordInput
          label="API Token"
          description={
            editing
              ? `Leave blank to keep the stored token (${shop.token_fingerprint}).`
              : 'Bearer JWT for Outbound requests. Stored encrypted; only its fingerprint is shown.'
          }
          required={!editing}
          {...form.getInputProps('api_token')}
        />
        <Group justify="flex-end">
          {onCancel && (
            <Button variant="default" onClick={onCancel}>
              Cancel
            </Button>
          )}
          <Button type="submit" loading={loading}>
            {submitLabel}
          </Button>
        </Group>
      </Stack>
    </form>
  );
}
