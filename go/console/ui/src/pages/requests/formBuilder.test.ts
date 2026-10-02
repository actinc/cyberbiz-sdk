import { describe, expect, it } from 'vitest';
import type { Operation, Param } from '../../api/types';
import {
  buildInitialValues,
  buildQuery,
  buildRequest,
  exampleToString,
  paramWidget,
  substitutePath,
  validateForm,
  validateParam,
} from './formBuilder';

/** Builds a Param with defaults. */
function param(overrides: Partial<Param> & { name: string }): Param {
  return {
    in: 'query',
    required: false,
    type: 'string',
    description: '',
    enum: null,
    example: null,
    ...overrides,
  };
}

const getOrder: Operation = {
  id: 'getOrder',
  method: 'GET',
  path: '/v1/orders/{id}',
  tag: 'Orders',
  summary: 'Get an order',
  description: '',
  path_params: [param({ name: 'id', in: 'path', required: true, type: 'integer', example: 12345 })],
  query_params: [
    param({ name: 'fields', type: 'array', example: ['id', 'status'] }),
    param({ name: 'include_items', type: 'boolean', example: true }),
    param({ name: 'status', enum: ['open', 'closed'], example: 'open' }),
  ],
  request_body: null,
  responses: [{ status: 200, description: 'OK', example: null }],
};

const createProduct: Operation = {
  ...getOrder,
  id: 'createProduct',
  method: 'POST',
  path: '/v1/products',
  path_params: [],
  query_params: [],
  request_body: {
    content_type: 'application/json',
    schema: {},
    example: { product: { title: 'Sample', price: 100 } },
  },
};

describe('buildInitialValues', () => {
  it('uses examples as defaults and prettifies the example body', () => {
    expect(buildInitialValues(getOrder)).toEqual({
      path: { id: '12345' },
      query: { fields: 'id,status', include_items: 'true', status: 'open' },
      body: '',
    });
    const withBody = buildInitialValues(createProduct);
    expect(withBody.body).toBe(
      JSON.stringify({ product: { title: 'Sample', price: 100 } }, null, 2),
    );
  });

  it('leaves parameters without examples blank', () => {
    const op: Operation = {
      ...getOrder,
      path_params: [param({ name: 'id', in: 'path', required: true })],
    };
    expect(buildInitialValues(op).path).toEqual({ id: '' });
  });
});

describe('exampleToString / paramWidget', () => {
  it('stringifies scalars, arrays and objects', () => {
    expect(exampleToString(5)).toBe('5');
    expect(exampleToString(false)).toBe('false');
    expect(exampleToString(['a', 1])).toBe('a,1');
    expect(exampleToString({ a: 1 })).toBe('{"a":1}');
    expect(exampleToString(null)).toBe('');
  });

  it('picks the widget from enum first, then type', () => {
    expect(paramWidget(param({ name: 'x', type: 'integer', enum: ['1', '2'] }))).toBe('enum');
    expect(paramWidget(param({ name: 'x', type: 'integer' }))).toBe('number');
    expect(paramWidget(param({ name: 'x', type: 'boolean' }))).toBe('boolean');
    expect(paramWidget(param({ name: 'x', type: 'string' }))).toBe('text');
    expect(paramWidget(param({ name: 'x', enum: [] }))).toBe('text');
  });
});

describe('validateParam', () => {
  it('requires required params and accepts blank optional ones', () => {
    expect(validateParam(param({ name: 'id', required: true }), ' ')).toBe('id is required');
    expect(validateParam(param({ name: 'id' }), '')).toBeNull();
  });

  it('checks integer, number, boolean and enum values', () => {
    expect(validateParam(param({ name: 'n', type: 'integer' }), '1.5')).toBe(
      'n must be an integer',
    );
    expect(validateParam(param({ name: 'n', type: 'integer' }), '-3')).toBeNull();
    expect(validateParam(param({ name: 'n', type: 'number' }), 'abc')).toBe('n must be a number');
    expect(validateParam(param({ name: 'n', type: 'number' }), '1.5e3')).toBeNull();
    expect(validateParam(param({ name: 'b', type: 'boolean' }), 'yes')).toBe(
      'b must be true or false',
    );
    expect(validateParam(param({ name: 'b', type: 'boolean' }), 'false')).toBeNull();
    expect(validateParam(param({ name: 's', enum: ['a', 'b'] }), 'c')).toBe(
      's must be one of a, b',
    );
  });
});

describe('validateForm / buildRequest', () => {
  it('collects errors keyed by location and name, including invalid JSON', () => {
    const errors = validateForm(getOrder, {
      path: { id: '' },
      query: { fields: '', include_items: 'maybe', status: '' },
      body: '{ nope',
    });
    expect(errors['path.id']).toBe('id is required');
    expect(errors['query.include_items']).toBe('include_items must be true or false');
    expect(errors.body).toMatch(/^Body is not valid JSON/);
    expect(Object.keys(errors)).toHaveLength(3);
  });

  it('builds the request body with substituted path and split array query', () => {
    const result = buildRequest(getOrder, buildInitialValues(getOrder));
    expect(result.ok).toBe(true);
    if (!result.ok) {
      return;
    }
    expect(result.request).toEqual({
      method: 'GET',
      path: '/v1/orders/12345',
      query: { fields: ['id', 'status'], include_items: ['true'], status: ['open'] },
      body: null,
      headers: {},
    });
  });

  it('parses the JSON body for operations that take one', () => {
    const result = buildRequest(createProduct, {
      path: {},
      query: {},
      body: '{"product":{"title":"X"}}',
    });
    expect(result.ok && result.request.body).toEqual({ product: { title: 'X' } });
  });

  it('returns errors instead of a request when invalid', () => {
    const result = buildRequest(getOrder, { path: { id: 'abc' }, query: {}, body: '' });
    expect(result).toEqual({ ok: false, errors: { 'path.id': 'id must be an integer' } });
  });
});

describe('substitutePath / buildQuery', () => {
  it('URL-encodes substituted values and leaves unknown placeholders intact', () => {
    expect(substitutePath('/v1/a/{x}/b/{y}', { x: 'p q/r' })).toBe('/v1/a/p%20q%2Fr/b/{y}');
  });

  it('drops blank query values and trims entries', () => {
    expect(
      buildQuery([param({ name: 'a' }), param({ name: 'b', type: 'array' })], {
        a: ' ',
        b: ' x , y,, ',
      }),
    ).toEqual({ b: ['x', 'y'] });
  });
});
