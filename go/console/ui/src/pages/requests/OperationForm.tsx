import {
  Accordion,
  Button,
  Group,
  Select,
  SimpleGrid,
  Stack,
  Text,
  TextInput,
  Textarea,
  Title,
} from '@mantine/core';
import { IconSend } from '@tabler/icons-react';
import { useEffect, useState, type JSX } from 'react';
import type { Operation, Param, SendRequestBody } from '../../api/types';
import { JsonView } from '../../components/JsonView';
import { MethodBadge } from '../../components/MethodBadge';
import { ShopSelect } from '../../components/ShopSelect';
import {
  buildInitialValues,
  buildRequest,
  paramWidget,
  type FormErrors,
  type OperationFormValues,
} from './formBuilder';

/** Returns a copy of the errors map without the given key. */
function omitKey(errors: FormErrors, key: string): FormErrors {
  return Object.fromEntries(Object.entries(errors).filter(([k]) => k !== key));
}

/** Props for {@link ParamInput}. */
interface ParamInputProps {
  param: Param;
  value: string;
  error: string | undefined;
  onChange: (value: string) => void;
}

/** Text, number, boolean or enum input for one parameter. */
function ParamInput({ param, value, error, onChange }: ParamInputProps): JSX.Element {
  const label = `${param.name}${param.required ? ' *' : ''}`;
  const description = param.description || `${param.in} · ${param.type}`;
  const widget = paramWidget(param);
  if (widget === 'enum' || widget === 'boolean') {
    const data = widget === 'boolean' ? ['true', 'false'] : (param.enum ?? []);
    return (
      <Select
        label={label}
        description={description}
        data={data}
        value={value === '' ? null : value}
        onChange={(v) => onChange(v ?? '')}
        error={error}
        clearable={!param.required}
        size="sm"
      />
    );
  }
  return (
    <TextInput
      label={label}
      description={description}
      value={value}
      onChange={(e) => onChange(e.currentTarget.value)}
      error={error}
      inputMode={widget === 'number' ? 'decimal' : undefined}
      placeholder={widget === 'number' ? '0' : undefined}
      size="sm"
      ff={widget === 'number' ? 'monospace' : undefined}
    />
  );
}

/** Props for {@link OperationForm}. */
export interface OperationFormProps {
  /** Selected operation. */
  operation: Operation;
  /** Selected Shop id. */
  shopId: number | null;
  /** Shop selection handler. */
  onShopChange: (id: number | null) => void;
  /** Called with the validated request body when Send is pressed. */
  onSend: (request: SendRequestBody) => void;
  /** Whether a request is in flight. */
  sending: boolean;
}

/** Docs, parameter form, JSON body editor, Shop selector and Send button for one operation. */
export function OperationForm({
  operation,
  shopId,
  onShopChange,
  onSend,
  sending,
}: OperationFormProps): JSX.Element {
  const [values, setValues] = useState<OperationFormValues>(() => buildInitialValues(operation));
  const [errors, setErrors] = useState<FormErrors>({});

  useEffect(() => {
    setValues(buildInitialValues(operation));
    setErrors({});
  }, [operation]);

  const setParam = (location: 'path' | 'query', name: string, value: string): void => {
    setValues((prev) => ({ ...prev, [location]: { ...prev[location], [name]: value } }));
    setErrors((prev) => omitKey(prev, `${location}.${name}`));
  };

  const handleSend = (): void => {
    const result = buildRequest(operation, values);
    if (!result.ok) {
      setErrors(result.errors);
      return;
    }
    setErrors({});
    onSend(result.request);
  };

  const renderParams = (
    location: 'path' | 'query',
    params: Param[],
    title: string,
  ): JSX.Element | null => {
    if (params.length === 0) {
      return null;
    }
    return (
      <div>
        <Text size="sm" fw={600} mb={4}>
          {title}
        </Text>
        <SimpleGrid cols={{ base: 1, md: 2 }} spacing="sm">
          {params.map((p) => (
            <ParamInput
              key={p.name}
              param={p}
              value={values[location][p.name] ?? ''}
              error={errors[`${location}.${p.name}`]}
              onChange={(v) => setParam(location, p.name, v)}
            />
          ))}
        </SimpleGrid>
      </div>
    );
  };

  return (
    <Stack>
      <div>
        <Group gap="sm" wrap="nowrap" align="flex-start">
          <MethodBadge method={operation.method} />
          <Title order={4} ff="monospace" style={{ wordBreak: 'break-all' }}>
            {operation.path}
          </Title>
        </Group>
        <Text fw={500} mt={4}>
          {operation.summary}
        </Text>
        {operation.description !== '' && (
          <Text size="sm" c="dimmed" style={{ whiteSpace: 'pre-wrap' }}>
            {operation.description}
          </Text>
        )}
      </div>

      {renderParams('path', operation.path_params, 'Path parameters')}
      {renderParams('query', operation.query_params, 'Query parameters')}

      {operation.request_body !== null && (
        <Textarea
          label={`Body (${operation.request_body.content_type})`}
          description="Prefilled with the documented Sample. Must be valid JSON."
          value={values.body}
          onChange={(e) => {
            const body = e.currentTarget.value;
            setValues((prev) => ({ ...prev, body }));
            setErrors((prev) => omitKey(prev, 'body'));
          }}
          error={errors.body}
          autosize
          minRows={6}
          maxRows={24}
          ff="monospace"
          styles={{ input: { fontSize: 12 } }}
        />
      )}

      <Group align="flex-end" gap="sm" wrap="wrap">
        <ShopSelect
          value={shopId}
          onChange={onShopChange}
          label="Shop"
          required
          w={{ base: '100%', sm: 300 }}
        />
        <Button
          leftSection={<IconSend size={16} />}
          onClick={handleSend}
          loading={sending}
          disabled={shopId === null}
        >
          Send
        </Button>
      </Group>

      {operation.responses.length > 0 && (
        <Accordion variant="contained">
          <Accordion.Item value="responses">
            <Accordion.Control>
              <Text size="sm">
                Documented responses ({operation.responses.map((r) => r.status).join(', ')})
              </Text>
            </Accordion.Control>
            <Accordion.Panel>
              <Stack>
                {operation.responses.map((r) => (
                  <div key={r.status}>
                    <Text size="sm" fw={500}>
                      {r.status} — {r.description}
                    </Text>
                    {r.example !== null && <JsonView value={r.example} maxHeight={240} />}
                  </div>
                ))}
              </Stack>
            </Accordion.Panel>
          </Accordion.Item>
        </Accordion>
      )}
    </Stack>
  );
}
