import type { JsonValue, Operation, Param, SendRequestBody } from '../../api/types';
import { prettyJson, tryParseJson } from '../../utils/json';

/**
 * Form state for one operation. Every parameter is edited as a string so the
 * same inputs work for text, numbers, booleans and enums; conversion and
 * validation happen in {@link buildRequest}.
 */
export interface OperationFormValues {
  /** Path parameter values keyed by name. */
  path: Record<string, string>;
  /** Query parameter values keyed by name; array types are comma-separated. */
  query: Record<string, string>;
  /** JSON body text (empty means no body). */
  body: string;
}

/** Validation errors keyed by field: `path.<name>`, `query.<name>` or `body`. */
export type FormErrors = Record<string, string>;

/** Outcome of {@link buildRequest}. */
export type BuildResult =
  { ok: true; request: SendRequestBody } | { ok: false; errors: FormErrors };

/** Widget the form should render for a parameter. */
export type ParamWidget = 'text' | 'number' | 'boolean' | 'enum';

/** Picks the input widget for a parameter. */
export function paramWidget(param: Param): ParamWidget {
  if (param.enum !== null && param.enum.length > 0) {
    return 'enum';
  }
  switch (param.type) {
    case 'integer':
    case 'number':
      return 'number';
    case 'boolean':
      return 'boolean';
    default:
      return 'text';
  }
}

/** Renders an example value as the string the form edits. */
export function exampleToString(example: JsonValue | null): string {
  if (example === null) {
    return '';
  }
  if (Array.isArray(example)) {
    return example.map((v) => (typeof v === 'string' ? v : JSON.stringify(v))).join(',');
  }
  if (typeof example === 'object') {
    return JSON.stringify(example);
  }
  return String(example);
}

/** Builds a name → value map from a parameter list using examples as defaults. */
function initialParamValues(params: Param[]): Record<string, string> {
  return Object.fromEntries(params.map((p) => [p.name, exampleToString(p.example)]));
}

/** Initial form values for an operation: examples for parameters, the example body prettified. */
export function buildInitialValues(op: Operation): OperationFormValues {
  const example = op.request_body?.example;
  return {
    path: initialParamValues(op.path_params),
    query: initialParamValues(op.query_params),
    body: example === undefined || example === null ? '' : prettyJson(example),
  };
}

const integerPattern = /^-?\d+$/;
const numberPattern = /^-?(\d+\.?\d*|\.\d+)([eE][+-]?\d+)?$/;

/** Validates one raw value against its parameter definition; returns an error message or null. */
export function validateParam(param: Param, raw: string): string | null {
  const value = raw.trim();
  if (value === '') {
    return param.required ? `${param.name} is required` : null;
  }
  if (param.enum !== null && param.enum.length > 0 && !param.enum.includes(value)) {
    return `${param.name} must be one of ${param.enum.join(', ')}`;
  }
  switch (param.type) {
    case 'integer':
      return integerPattern.test(value) ? null : `${param.name} must be an integer`;
    case 'number':
      return numberPattern.test(value) ? null : `${param.name} must be a number`;
    case 'boolean':
      return value === 'true' || value === 'false' ? null : `${param.name} must be true or false`;
    default:
      return null;
  }
}

/** Validates every parameter and the JSON body; empty map means valid. */
export function validateForm(op: Operation, values: OperationFormValues): FormErrors {
  const errors: FormErrors = {};
  for (const p of op.path_params) {
    const err = validateParam(p, values.path[p.name] ?? '');
    if (err !== null) {
      errors[`path.${p.name}`] = err;
    }
  }
  for (const p of op.query_params) {
    const err = validateParam(p, values.query[p.name] ?? '');
    if (err !== null) {
      errors[`query.${p.name}`] = err;
    }
  }
  const parsed = tryParseJson(values.body);
  if (!parsed.ok) {
    errors.body = `Body is not valid JSON: ${parsed.error}`;
  }
  return errors;
}

/** Replaces `{name}` placeholders in a path template with URL-encoded values. */
export function substitutePath(template: string, params: Record<string, string>): string {
  return template.replace(/\{([^}]+)\}/g, (match, name: string) => {
    const value = params[name];
    return value === undefined || value === '' ? match : encodeURIComponent(value.trim());
  });
}

/** Turns query form values into the `{name: [values]}` map the backend expects. */
export function buildQuery(
  params: Param[],
  values: Record<string, string>,
): Record<string, string[]> {
  const query: Record<string, string[]> = {};
  for (const p of params) {
    const raw = (values[p.name] ?? '').trim();
    if (raw === '') {
      continue;
    }
    query[p.name] =
      p.type === 'array'
        ? raw
            .split(',')
            .map((v) => v.trim())
            .filter((v) => v !== '')
        : [raw];
  }
  return query;
}

/** Validates the form and, when valid, builds the `POST /api/shops/:id/requests` body. */
export function buildRequest(op: Operation, values: OperationFormValues): BuildResult {
  const errors = validateForm(op, values);
  if (Object.keys(errors).length > 0) {
    return { ok: false, errors };
  }
  const parsed = tryParseJson(values.body);
  const body = parsed.ok ? parsed.value : null;
  return {
    ok: true,
    request: {
      method: op.method,
      path: substitutePath(op.path, values.path),
      query: buildQuery(op.query_params, values.query),
      body: op.request_body === null && body === null ? null : body,
      headers: {},
    },
  };
}
